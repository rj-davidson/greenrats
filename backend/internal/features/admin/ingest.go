package admin

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/placement"
	"github.com/rj-davidson/greenrats/ent/season"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/exa"
	"github.com/rj-davidson/greenrats/internal/external/googlemaps"
	"github.com/rj-davidson/greenrats/internal/external/openai"
	"github.com/rj-davidson/greenrats/internal/features/golfers"
	"github.com/rj-davidson/greenrats/internal/sync"
)

type IngestService struct {
	db          *ent.Client
	config      *config.Config
	ballDontLie *balldontlie.Client
	syncService *sync.Service
	exa         *exa.Client
	openai      *openai.Client
	golfers     *golfers.Service
	logger      *slog.Logger
}

func NewIngestService(
	db *ent.Client,
	cfg *config.Config,
	ballDontLie *balldontlie.Client,
	googlemapsClient *googlemaps.Client,
	exaClient *exa.Client,
	openaiClient *openai.Client,
	golfersSvc *golfers.Service,
	logger *slog.Logger,
) *IngestService {
	return &IngestService{
		db:          db,
		config:      cfg,
		ballDontLie: ballDontLie,
		syncService: sync.NewService(db, googlemapsClient, logger),
		exa:         exaClient,
		openai:      openaiClient,
		golfers:     golfersSvc,
		logger:      logger,
	}
}

func (s *IngestService) SyncTournaments(ctx context.Context) error {
	s.logger.Info("starting tournament sync")

	seasonEnt, err := s.syncService.UpsertSeason(ctx, s.config.CurrentSeason)
	if err != nil {
		return fmt.Errorf("failed to ensure season exists: %w", err)
	}

	tournaments, err := s.ballDontLie.GetTournaments(ctx, s.config.CurrentSeason)
	if err != nil {
		return fmt.Errorf("failed to fetch tournaments: %w", err)
	}

	s.logger.Info("fetched tournaments", "count", len(tournaments), "season", s.config.CurrentSeason)

	for idx := range tournaments {
		if _, err := s.syncService.UpsertTournament(ctx, &tournaments[idx], seasonEnt); err != nil {
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
		if err := s.syncService.UpsertPlayer(ctx, &players[idx]); err != nil {
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
		if err := s.syncService.UpsertPlacement(ctx, t, &results[idx]); err != nil {
			s.logger.Warn("failed to upsert placement", "player", results[idx].Player.DisplayName, "error", err)
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

	placements, err := s.db.Placement.Query().
		Where(placement.HasTournamentWith(tournament.IDEQ(t.ID))).
		WithGolfer().
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query placements: %w", err)
	}

	if len(placements) == 0 {
		s.logger.Info("no placements, skipping earnings sync", "tournament", t.Name)
		return nil
	}

	golferInputs := make([]openai.GolferInput, 0, len(placements))
	placementByGolferID := make(map[string]*ent.Placement)
	for _, p := range placements {
		if p.Edges.Golfer == nil || p.Status == placement.StatusCut || (p.PositionNumeric == nil || *p.PositionNumeric == 0) {
			continue
		}
		g := p.Edges.Golfer
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
		placementByGolferID[g.ID.String()] = p
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

	s.logger.Info("using OpenAI for golfer matching", "count", len(golferInputs), "tournament", t.Name)
	updated := s.matchWithOpenAI(ctx, t, exaContent.String(), golferInputs, placementByGolferID)

	s.logger.Info("earnings sync completed", "tournament", t.Name, "updated", updated)
	return nil
}

func (s *IngestService) SyncField(ctx context.Context, tournamentID uuid.UUID) error {
	t, err := s.db.Tournament.Get(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("failed to find tournament: %w", err)
	}

	if t.BdlID == nil {
		return fmt.Errorf("tournament has no BallDontLie ID")
	}

	s.logger.Info("starting field sync", "tournament", t.Name)

	fields, err := s.ballDontLie.GetTournamentField(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("failed to fetch tournament field: %w", err)
	}

	s.logger.Info("fetched field", "tournament", t.Name, "count", len(fields))

	processed := 0
	for idx := range fields {
		if err := s.syncService.UpsertFieldEntry(ctx, t, &fields[idx]); err != nil {
			s.logger.Warn("failed to upsert field entry", "player", fields[idx].Player.DisplayName, "error", err)
			continue
		}
		processed++
	}

	s.logger.Info("field sync completed", "tournament", t.Name, "processed", processed)
	return nil
}

const earningsBatchSize = 10

func (s *IngestService) matchWithOpenAI(
	ctx context.Context,
	t *ent.Tournament,
	content string,
	golferInputs []openai.GolferInput,
	placementByGolferID map[string]*ent.Placement,
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
		p, ok := placementByGolferID[r.GolferID]
		if !ok {
			continue
		}

		if p.Earnings != r.Earnings {
			_, err := p.Update().
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

func (s *IngestService) SyncCourses(ctx context.Context) error {
	s.logger.Info("starting courses sync")

	courses, err := s.ballDontLie.GetCourses(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch courses: %w", err)
	}

	s.logger.Info("fetched courses", "count", len(courses))

	coursesProcessed := 0
	holesProcessed := 0

	for idx := range courses {
		c := &courses[idx]

		entCourse, err := s.syncService.UpsertCourse(ctx, c)
		if err != nil {
			s.logger.Warn("failed to upsert course", "name", c.Name, "error", err)
			continue
		}
		coursesProcessed++

		holes, err := s.ballDontLie.GetCourseHoles(ctx, c.ID)
		if err != nil {
			s.logger.Warn("failed to fetch holes", "course", c.Name, "error", err)
			continue
		}

		for hIdx := range holes {
			if err := s.syncService.UpsertCourseHole(ctx, entCourse.ID, &holes[hIdx]); err != nil {
				s.logger.Warn("failed to upsert hole", "course", c.Name, "hole", holes[hIdx].HoleNumber, "error", err)
				continue
			}
			holesProcessed++
		}
	}

	s.logger.Info("courses sync completed", "courses", coursesProcessed, "holes", holesProcessed)
	return nil
}

var relevantStatIDs = []int{
	120, // Scoring Average
	110, // Top 10 Finishes
	107, // Cuts Made
	106, // Events Played
	109, // Wins
	102, // Driving Distance
	101, // Driving Accuracy Percentage
	103, // Greens in Regulation Percentage
	104, // Putting Average
	130, // Scrambling
}

func (s *IngestService) SyncGolferSeasonStats(ctx context.Context) error {
	s.logger.Info("starting golfer season stats sync")

	currentSeason, err := s.db.Season.Query().
		Where(season.IsCurrent(true)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current season: %w", err)
	}

	stats, err := s.ballDontLie.GetPlayerSeasonStats(ctx, currentSeason.Year, relevantStatIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch player season stats: %w", err)
	}

	s.logger.Info("fetched player season stats", "count", len(stats))

	processed := 0
	for idx := range stats {
		stat := &stats[idx]

		g, err := s.db.Golfer.Query().
			Where(golfer.BdlID(stat.Player.ID)).
			Only(ctx)
		if err != nil {
			continue
		}

		if err := s.syncService.UpsertGolferSeasonStat(ctx, g.ID, currentSeason.ID, stat); err != nil {
			s.logger.Warn("failed to upsert stat", "golfer", stat.Player.DisplayName, "error", err)
			continue
		}
		processed++
	}

	s.logger.Info("golfer season stats sync completed", "stats_processed", processed)
	return nil
}
