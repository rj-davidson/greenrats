package admin

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/exa"
	"github.com/rj-davidson/greenrats/internal/external/googlemaps"
	"github.com/rj-davidson/greenrats/internal/external/openai"
	"github.com/rj-davidson/greenrats/internal/external/pgatour"
	"github.com/rj-davidson/greenrats/internal/features/fields"
	"github.com/rj-davidson/greenrats/internal/features/golfers"
	"github.com/rj-davidson/greenrats/internal/sync"
)

type IngestService struct {
	db            *ent.Client
	config        *config.Config
	ballDontLie   *balldontlie.Client
	syncService   *sync.Service
	fieldsService *fields.Service
	exa           *exa.Client
	openai        *openai.Client
	golfers       *golfers.Service
	logger        *slog.Logger
}

func NewIngestService(
	db *ent.Client,
	cfg *config.Config,
	ballDontLie *balldontlie.Client,
	pgatourClient *pgatour.Client,
	googlemapsClient *googlemaps.Client,
	exaClient *exa.Client,
	openaiClient *openai.Client,
	golfersSvc *golfers.Service,
	logger *slog.Logger,
) *IngestService {
	return &IngestService{
		db:            db,
		config:        cfg,
		ballDontLie:   ballDontLie,
		syncService:   sync.NewService(db, googlemapsClient, logger),
		fieldsService: fields.NewService(db, pgatourClient, logger),
		exa:           exaClient,
		openai:        openaiClient,
		golfers:       golfersSvc,
		logger:        logger,
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
		if _, err := s.syncService.UpsertTournament(ctx, &tournaments[idx]); err != nil {
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
		if err := s.syncService.UpsertTournamentEntry(ctx, t, &results[idx]); err != nil {
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

	s.logger.Info("using OpenAI for golfer matching", "count", len(golferInputs), "tournament", t.Name)
	updated := s.matchWithOpenAI(ctx, t, exaContent.String(), golferInputs, entryByGolferID)

	s.logger.Info("earnings sync completed", "tournament", t.Name, "updated", updated)
	return nil
}

func (s *IngestService) SyncField(ctx context.Context, tournamentID uuid.UUID) error {
	result, err := s.fieldsService.SyncTournamentField(ctx, tournamentID)
	if err != nil {
		return err
	}

	s.logger.Info("field sync completed",
		"tournament", result.TournamentName,
		"total", result.TotalPlayers,
		"matched", result.MatchedPlayers,
		"new", result.NewEntries,
		"updated", result.UpdatedEntries,
	)

	return nil
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
