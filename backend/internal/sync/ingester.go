package sync

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/email"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/pgatour"
	"github.com/rj-davidson/greenrats/internal/features/fields"
)

const (
	TournamentSyncInterval  = 24 * time.Hour
	LeaderboardSyncInterval = 1 * time.Hour
	PlayerSyncInterval      = 7 * 24 * time.Hour
	ReminderCheckInterval   = 1 * time.Hour
	EarningsCheckInterval   = 6 * time.Hour
	FieldSyncInterval       = 12 * time.Hour
)

type Ingester struct {
	db            *ent.Client
	config        *config.Config
	ballDontLie   *balldontlie.Client
	pgatourClient *pgatour.Client
	syncService   *Service
	fieldsService *fields.Service
	email         *email.Client
	logger        *slog.Logger
}

func NewIngester(
	db *ent.Client,
	cfg *config.Config,
	ballDontLie *balldontlie.Client,
	pgatourClient *pgatour.Client,
	syncService *Service,
	fieldsService *fields.Service,
	emailClient *email.Client,
	logger *slog.Logger,
) *Ingester {
	return &Ingester{
		db:            db,
		config:        cfg,
		ballDontLie:   ballDontLie,
		pgatourClient: pgatourClient,
		syncService:   syncService,
		fieldsService: fieldsService,
		email:         emailClient,
		logger:        logger,
	}
}

func (i *Ingester) Run(ctx context.Context) {
	tournamentTicker := time.NewTicker(TournamentSyncInterval)
	defer tournamentTicker.Stop()

	leaderboardTicker := time.NewTicker(LeaderboardSyncInterval)
	defer leaderboardTicker.Stop()

	playerTicker := time.NewTicker(PlayerSyncInterval)
	defer playerTicker.Stop()

	reminderTicker := time.NewTicker(ReminderCheckInterval)
	defer reminderTicker.Stop()

	earningsTicker := time.NewTicker(EarningsCheckInterval)
	defer earningsTicker.Stop()

	fieldTicker := time.NewTicker(FieldSyncInterval)
	defer fieldTicker.Stop()

	log.Println("Checking for required initial syncs...")
	if i.shouldSync(ctx, "tournaments", TournamentSyncInterval) {
		i.syncTournaments(ctx)
	} else {
		log.Println("Skipping tournament sync (data is fresh)")
	}
	if i.shouldSync(ctx, "players", PlayerSyncInterval) {
		i.syncPlayers(ctx)
	} else {
		log.Println("Skipping player sync (data is fresh)")
	}
	if i.shouldSync(ctx, "leaderboards", LeaderboardSyncInterval) {
		i.syncLeaderboards(ctx)
	} else {
		log.Println("Skipping leaderboard sync (data is fresh)")
	}
	if i.shouldSync(ctx, "earnings", EarningsCheckInterval) {
		i.syncEarnings(ctx)
	} else {
		log.Println("Skipping earnings sync (data is fresh)")
	}
	if i.shouldSync(ctx, "fields", FieldSyncInterval) {
		i.syncFields(ctx)
	} else {
		log.Println("Skipping field sync (data is fresh)")
	}
	i.sendPickReminders(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-tournamentTicker.C:
			i.syncTournaments(ctx)
		case <-leaderboardTicker.C:
			i.syncLeaderboards(ctx)
		case <-playerTicker.C:
			i.syncPlayers(ctx)
		case <-earningsTicker.C:
			i.syncEarnings(ctx)
		case <-fieldTicker.C:
			i.syncFields(ctx)
		case <-reminderTicker.C:
			i.sendPickReminders(ctx)
		}
	}
}
