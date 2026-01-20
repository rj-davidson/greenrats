package admin

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/exa"
	"github.com/rj-davidson/greenrats/internal/external/openai"
	"github.com/rj-davidson/greenrats/internal/external/scrapedo"
	"github.com/rj-davidson/greenrats/internal/features/golfers"
)

type IngestService struct {
	db          *ent.Client
	config      *config.Config
	ballDontLie *balldontlie.Client
	exa         *exa.Client
	openai      *openai.Client
	scrapedo    *scrapedo.Client
	golfers     *golfers.Service
	logger      *slog.Logger
}

func NewIngestService(
	db *ent.Client,
	cfg *config.Config,
	ballDontLie *balldontlie.Client,
	exaClient *exa.Client,
	openaiClient *openai.Client,
	scrapeDoClient *scrapedo.Client,
	golfersSvc *golfers.Service,
	logger *slog.Logger,
) *IngestService {
	return &IngestService{
		db:          db,
		config:      cfg,
		ballDontLie: ballDontLie,
		exa:         exaClient,
		openai:      openaiClient,
		scrapedo:    scrapeDoClient,
		golfers:     golfersSvc,
		logger:      logger,
	}
}

func (s *IngestService) SyncTournaments(ctx context.Context) error {
	s.logger.Info("starting tournament sync")

	tournaments, err := s.ballDontLie.GetTournaments(ctx, s.config.CurrentSeason)
	if err != nil {
		return fmt.Errorf("failed to fetch tournaments: %w", err)
	}

	s.logger.Info("fetched tournaments", "count", len(tournaments), "season", s.config.CurrentSeason)

	for idx := range tournaments {
		if err := s.upsertTournament(ctx, &tournaments[idx]); err != nil {
			s.logger.Warn("failed to upsert tournament", "name", tournaments[idx].Name, "error", err)
			continue
		}
	}

	s.logger.Info("tournament sync completed")
	return nil
}

func (s *IngestService) SyncPlayers(ctx context.Context) error {
	s.logger.Info("starting player sync")

	players, err := s.ballDontLie.GetPlayers(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch players: %w", err)
	}

	s.logger.Info("fetched players", "count", len(players))

	for idx := range players {
		if err := s.upsertPlayer(ctx, &players[idx]); err != nil {
			s.logger.Warn("failed to upsert player", "name", players[idx].DisplayName, "error", err)
			continue
		}
	}

	s.logger.Info("player sync completed")
	return nil
}

func (s *IngestService) SyncLeaderboard(ctx context.Context, tournamentID uuid.UUID) error {
	t, err := s.db.Tournament.Get(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("failed to find tournament: %w", err)
	}

	if t.BdlID == nil {
		return fmt.Errorf("tournament has no BallDontLie ID")
	}

	s.logger.Info("starting leaderboard sync", "tournament", t.Name)

	results, err := s.ballDontLie.GetTournamentResults(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("failed to fetch tournament results: %w", err)
	}

	s.logger.Info("fetched results", "tournament", t.Name, "count", len(results))

	for idx := range results {
		if err := s.upsertTournamentEntry(ctx, t, &results[idx]); err != nil {
			s.logger.Warn("failed to upsert entry", "player", results[idx].Player.DisplayName, "error", err)
			continue
		}
	}

	s.logger.Info("leaderboard sync completed", "tournament", t.Name)
	return nil
}

func (s *IngestService) SyncEarnings(ctx context.Context, tournamentID uuid.UUID) error {
	t, err := s.db.Tournament.Get(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("failed to find tournament: %w", err)
	}

	s.logger.Info("starting earnings sync", "tournament", t.Name)

	year := t.SeasonYear
	if year == 0 {
		year = t.EndDate.Year()
	}

	entries, err := s.db.TournamentEntry.Query().
		Where(tournamententry.HasTournamentWith(tournament.IDEQ(t.ID))).
		WithGolfer().
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query tournament entries: %w", err)
	}

	if len(entries) == 0 {
		s.logger.Info("no tournament entries, skipping earnings sync", "tournament", t.Name)
		return nil
	}

	golferInputs := make([]openai.GolferInput, 0, len(entries))
	entryByGolferID := make(map[string]*ent.TournamentEntry)
	for _, entry := range entries {
		if entry.Edges.Golfer == nil || entry.Cut || entry.Position == 0 {
			continue
		}
		g := entry.Edges.Golfer
		input := openai.GolferInput{
			GolferID: g.ID.String(),
			Name:     g.Name,
		}
		if g.FirstName != nil {
			input.FirstName = *g.FirstName
		}
		if g.LastName != nil {
			input.LastName = *g.LastName
		}
		golferInputs = append(golferInputs, input)
		entryByGolferID[g.ID.String()] = entry
	}

	if len(golferInputs) == 0 {
		s.logger.Info("no eligible golfers for earnings sync", "tournament", t.Name)
		return nil
	}

	exaResponse, err := s.exa.SearchEarnings(ctx, t.Name, year)
	if err != nil {
		return fmt.Errorf("failed to search earnings via Exa: %w", err)
	}

	if len(exaResponse.Results) == 0 {
		s.logger.Warn("no Exa results found", "tournament", t.Name)
		return nil
	}

	s.logger.Debug("exa results", "tournament", t.Name, "count", len(exaResponse.Results))

	var exaContent strings.Builder
	for _, result := range exaResponse.Results {
		exaContent.WriteString(result.Text)
		exaContent.WriteString("\n\n")
	}

	matchedGolferIDs := make(map[string]bool)
	var updated int

	if s.config.ScrapeDoAPIKey != "" {
		updated, matchedGolferIDs = s.tryScrapedoEarnings(ctx, t, exaResponse, golferInputs, entryByGolferID)
	}

	unmatchedGolfers := make([]openai.GolferInput, 0)
	for _, g := range golferInputs {
		if !matchedGolferIDs[g.GolferID] {
			unmatchedGolfers = append(unmatchedGolfers, g)
		}
	}

	if len(unmatchedGolfers) > 0 {
		s.logger.Info("using OpenAI for remaining golfers", "count", len(unmatchedGolfers), "tournament", t.Name)
		openaiUpdated := s.matchWithOpenAI(ctx, t, exaContent.String(), unmatchedGolfers, entryByGolferID)
		updated += openaiUpdated
	}

	s.logger.Info("earnings sync completed", "tournament", t.Name, "updated", updated)
	return nil
}

func (s *IngestService) upsertTournament(ctx context.Context, t *balldontlie.Tournament) error {
	const iso8601Millis = "2006-01-02T15:04:05.000Z"
	startDate, err := time.Parse(iso8601Millis, t.StartDate)
	if err != nil {
		return fmt.Errorf("failed to parse start date: %w", err)
	}

	endDate := s.parseEndDate(t.EndDate, startDate)

	now := time.Now().UTC()
	status := tournament.StatusUpcoming
	if now.After(endDate) {
		status = tournament.StatusCompleted
	} else if now.After(startDate) || now.Equal(startDate) {
		status = tournament.StatusActive
	}

	existing, err := s.db.Tournament.Query().
		Where(tournament.BdlID(t.ID)).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		builder := s.db.Tournament.Create().
			SetBdlID(t.ID).
			SetName(t.Name).
			SetStartDate(startDate).
			SetEndDate(endDate).
			SetStatus(status).
			SetSeasonYear(t.Season)

		if t.CourseName != nil && *t.CourseName != "" {
			builder.SetCourse(*t.CourseName)
		}
		if t.City != nil && *t.City != "" {
			builder.SetLocation(*t.City)
		}
		if t.Purse != nil && *t.Purse != "" {
			if purse, err := strconv.Atoi(*t.Purse); err == nil && purse > 0 {
				builder.SetPurse(purse)
			}
		}

		_, err = builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create tournament: %w", err)
		}
		s.logger.Debug("created tournament", "name", t.Name)
	case err != nil:
		return fmt.Errorf("failed to query tournament: %w", err)
	default:
		updater := existing.Update().
			SetName(t.Name).
			SetStartDate(startDate).
			SetEndDate(endDate).
			SetStatus(status)

		if t.CourseName != nil && *t.CourseName != "" {
			updater.SetCourse(*t.CourseName)
		}
		if t.City != nil && *t.City != "" {
			updater.SetLocation(*t.City)
		}
		if t.Purse != nil && *t.Purse != "" {
			if purse, err := strconv.Atoi(*t.Purse); err == nil && purse > 0 {
				updater.SetPurse(purse)
			}
		}

		_, err = updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update tournament: %w", err)
		}
		s.logger.Debug("updated tournament", "name", t.Name, "status", status)
	}

	return nil
}

func (s *IngestService) parseEndDate(endDateStr *string, startDate time.Time) time.Time {
	if endDateStr == nil || *endDateStr == "" {
		return startDate.AddDate(0, 0, 4).Add(6 * time.Hour)
	}

	const iso8601Millis = "2006-01-02T15:04:05.000Z"
	if endDate, err := time.Parse(iso8601Millis, *endDateStr); err == nil {
		return endDate.AddDate(0, 0, 1).Add(6 * time.Hour)
	}

	str := strings.TrimSpace(*endDateStr)
	if idx := strings.LastIndex(str, " - "); idx != -1 {
		endPart := strings.TrimSpace(str[idx+3:])

		if day, err := strconv.Atoi(endPart); err == nil && day >= 1 && day <= 31 {
			endDate := time.Date(startDate.Year(), startDate.Month(), day, 0, 0, 0, 0, time.UTC)
			if endDate.Before(startDate) {
				endDate = endDate.AddDate(0, 1, 0)
			}
			return endDate.AddDate(0, 0, 1).Add(6 * time.Hour)
		}

		formats := []string{"Jan 2", "January 2"}
		for _, format := range formats {
			if parsed, err := time.Parse(format, endPart); err == nil {
				endDate := time.Date(startDate.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
				if endDate.Before(startDate) {
					endDate = endDate.AddDate(1, 0, 0)
				}
				return endDate.AddDate(0, 0, 1).Add(6 * time.Hour)
			}
		}
	}

	return startDate.AddDate(0, 0, 4).Add(6 * time.Hour)
}

func (s *IngestService) upsertPlayer(ctx context.Context, p *balldontlie.Player) error {
	name := p.DisplayName
	if name == "" && p.FirstName != nil && p.LastName != nil {
		name = fmt.Sprintf("%s %s", *p.FirstName, *p.LastName)
	}

	countryCode := "UNK"
	if p.CountryCode != nil && *p.CountryCode != "" {
		countryCode = *p.CountryCode
	}

	existing, err := s.db.Golfer.Query().
		Where(golfer.BdlID(p.ID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		existing, err = s.db.Golfer.Query().
			Where(golfer.Name(name)).
			Only(ctx)
	}

	switch {
	case ent.IsNotFound(err):
		builder := s.db.Golfer.Create().
			SetBdlID(p.ID).
			SetName(name).
			SetCountryCode(countryCode).
			SetActive(p.Active)

		if p.FirstName != nil {
			builder.SetFirstName(*p.FirstName)
		}
		if p.LastName != nil {
			builder.SetLastName(*p.LastName)
		}
		if p.Country != nil && *p.Country != "" {
			builder.SetCountry(*p.Country)
		}
		if p.OWGR != nil && *p.OWGR > 0 {
			builder.SetOwgr(*p.OWGR)
		}

		_, err = builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create golfer: %w", err)
		}
		s.logger.Debug("created golfer", "name", name)
	case err != nil:
		return fmt.Errorf("failed to query golfer: %w", err)
	default:
		updater := existing.Update().
			SetName(name).
			SetCountryCode(countryCode).
			SetActive(p.Active).
			SetBdlID(p.ID)

		if p.FirstName != nil {
			updater.SetFirstName(*p.FirstName)
		}
		if p.LastName != nil {
			updater.SetLastName(*p.LastName)
		}
		if p.Country != nil && *p.Country != "" {
			updater.SetCountry(*p.Country)
		}
		if p.OWGR != nil && *p.OWGR > 0 {
			updater.SetOwgr(*p.OWGR)
		}

		_, err = updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update golfer: %w", err)
		}
	}

	return nil
}

func (s *IngestService) upsertTournamentEntry(ctx context.Context, t *ent.Tournament, r *balldontlie.TournamentResult) error {
	g, err := s.db.Golfer.Query().
		Where(golfer.BdlID(r.Player.ID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to query golfer: %w", err)
	}

	status := tournamententry.StatusActive
	cut := false
	if r.Tournament.Status != nil {
		switch *r.Tournament.Status {
		case "COMPLETED":
			status = tournamententry.StatusFinished
		case "IN_PROGRESS":
			status = tournamententry.StatusActive
		}
	}

	position := 0
	if r.PositionNumeric != nil {
		position = *r.PositionNumeric
	}
	if r.Position != nil && *r.Position == "CUT" {
		cut = true
		status = tournamententry.StatusFinished
	}

	score := 0
	if r.TotalScore != nil {
		score = *r.TotalScore
	}
	totalStrokes := 0

	existing, err := s.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.ID(t.ID)),
			tournamententry.HasGolferWith(golfer.ID(g.ID)),
		).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		_, err = s.db.TournamentEntry.Create().
			SetTournament(t).
			SetGolfer(g).
			SetPosition(position).
			SetCut(cut).
			SetScore(score).
			SetTotalStrokes(totalStrokes).
			SetStatus(status).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create tournament entry: %w", err)
		}
	case err != nil:
		return fmt.Errorf("failed to query tournament entry: %w", err)
	default:
		_, err = existing.Update().
			SetPosition(position).
			SetCut(cut).
			SetScore(score).
			SetTotalStrokes(totalStrokes).
			SetStatus(status).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update tournament entry: %w", err)
		}
	}

	return nil
}

func (s *IngestService) tryScrapedoEarnings(
	ctx context.Context,
	t *ent.Tournament,
	exaResponse *exa.SearchResponse,
	golferInputs []openai.GolferInput,
	entryByGolferID map[string]*ent.TournamentEntry,
) (updated int, matchedIDs map[string]bool) {
	matchedIDs = make(map[string]bool)
	targetMatches := int(math.Ceil(scrapedo.EarningsMatchThreshold * float64(len(golferInputs))))

	for _, result := range exaResponse.Results {
		s.logger.Debug("trying scrape.do", "url", result.URL, "tournament", t.Name)

		scrapeResp, err := s.scrapedo.Scrape(ctx, result.URL, scrapedo.ScrapeOptions{Render: true})
		if err != nil {
			s.logger.Warn("scrape.do failed", "url", result.URL, "error", err)
			continue
		}

		if scrapeResp.StatusCode != 200 {
			s.logger.Warn("scrape.do returned non-200 status", "url", result.URL, "status", scrapeResp.StatusCode)
			continue
		}

		parseResult := scrapedo.ParseLeaderboard(scrapeResp.Content)
		if !parseResult.Success {
			s.logger.Debug("programmatic parsing failed", "url", result.URL, "entries", len(parseResult.Entries))
			continue
		}

		for _, golferInput := range golferInputs {
			if matchedIDs[golferInput.GolferID] {
				continue
			}

			for _, parsed := range parseResult.Entries {
				if !s.namesMatch(golferInput, parsed.Name) {
					continue
				}
				entry := entryByGolferID[golferInput.GolferID]
				if entry != nil && entry.Earnings != parsed.Earnings {
					_, err := entry.Update().
						SetEarnings(parsed.Earnings).
						Save(ctx)
					if err != nil {
						s.logger.Warn("failed to update earnings", "golfer_id", golferInput.GolferID, "error", err)
						continue
					}
					updated++
				}
				matchedIDs[golferInput.GolferID] = true
				break
			}
		}

		if len(matchedIDs) >= targetMatches {
			break
		}
	}

	return updated, matchedIDs
}

func (s *IngestService) namesMatch(golferInput openai.GolferInput, parsedName string) bool {
	parsedCandidates := golfers.NormalizeName(parsedName)
	golferCandidates := golfers.NormalizeName(golferInput.Name)

	for _, pc := range parsedCandidates {
		for _, gc := range golferCandidates {
			if strings.EqualFold(pc, gc) {
				return true
			}
		}
	}

	if golferInput.FirstName != "" && golferInput.LastName != "" {
		fullName := golferInput.FirstName + " " + golferInput.LastName
		for _, pc := range parsedCandidates {
			if strings.EqualFold(pc, fullName) {
				return true
			}
		}
	}

	return false
}

const earningsBatchSize = 10

func (s *IngestService) matchWithOpenAI(
	ctx context.Context,
	t *ent.Tournament,
	content string,
	golferInputs []openai.GolferInput,
	entryByGolferID map[string]*ent.TournamentEntry,
) int {
	leaderboard, err := s.openai.ParseLeaderboardContent(ctx, content, t.Name)
	if err != nil {
		s.logger.Warn("failed to parse leaderboard content via OpenAI", "error", err)
		return 0
	}

	if len(leaderboard.Entries) == 0 {
		s.logger.Debug("no leaderboard entries parsed", "tournament", t.Name)
		return 0
	}

	var allResults []openai.EarningsResult
	numBatches := (len(golferInputs) + earningsBatchSize - 1) / earningsBatchSize
	for batch := range numBatches {
		start := batch * earningsBatchSize
		end := min(start+earningsBatchSize, len(golferInputs))
		batchGolfers := golferInputs[start:end]

		results, err := s.openai.MatchPlayersToLeaderboard(ctx, leaderboard, batchGolfers)
		if err != nil {
			s.logger.Warn("failed to match batch", "batch", batch+1, "tournament", t.Name, "error", err)
			continue
		}

		allResults = append(allResults, results...)
	}

	var updated int
	for _, r := range allResults {
		entry, ok := entryByGolferID[r.GolferID]
		if !ok {
			continue
		}

		if entry.Earnings != r.Earnings {
			_, err := entry.Update().
				SetEarnings(r.Earnings).
				Save(ctx)
			if err != nil {
				s.logger.Warn("failed to update earnings", "golfer_id", r.GolferID, "error", err)
				continue
			}
			updated++
		}
	}

	return updated
}
