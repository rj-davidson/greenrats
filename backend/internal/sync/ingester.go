package sync

import (
	"context"
	"log/slog"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/email"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/pgatour"
	"github.com/rj-davidson/greenrats/internal/features/fields"
)

const (
	TournamentSyncInterval  = 24 * time.Hour
	LeaderboardPlayInterval = 5 * time.Minute
	LeaderboardIdleInterval = 1 * time.Hour
	ScorecardSyncInterval   = 10 * time.Minute
	ReminderCheckInterval   = 1 * time.Hour
	EarningsCheckInterval   = 6 * time.Hour
	SchedulerTickInterval   = 1 * time.Minute
	FieldSyncHour           = 9
	PlayerSyncHour          = 21
	PlayHoursStart          = 8
	PlayHoursEnd            = 20
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
	ticker := time.NewTicker(SchedulerTickInterval)
	defer ticker.Stop()

	var lastTournamentSync, lastLeaderboardSync, lastScorecardSync time.Time
	var lastFieldSync, lastEarningsSync, lastPlayerSync, lastReminderSync time.Time

	i.logger.Info("ingester started, checking for required initial syncs")
	if i.shouldSync(ctx, "tournaments", TournamentSyncInterval) {
		i.syncTournaments(ctx)
		lastTournamentSync = time.Now()
	} else {
		i.logger.Debug("skipping tournament sync (data is fresh)")
		lastTournamentSync = time.Now()
	}
	if i.shouldSyncPlayersNow() {
		i.syncPlayers(ctx)
		lastPlayerSync = time.Now()
	} else {
		i.logger.Debug("skipping player sync (not scheduled time)")
	}
	if i.shouldSync(ctx, "leaderboards", i.getLeaderboardInterval(ctx)) {
		i.syncLeaderboards(ctx)
		lastLeaderboardSync = time.Now()
	} else {
		i.logger.Debug("skipping leaderboard sync (data is fresh)")
		lastLeaderboardSync = time.Now()
	}
	if i.isAnyTournamentInPlayHours(ctx) {
		if i.shouldSync(ctx, "scorecards", ScorecardSyncInterval) {
			i.syncScorecards(ctx)
			lastScorecardSync = time.Now()
		}
	}
	if i.shouldSync(ctx, "earnings", EarningsCheckInterval) {
		i.syncEarnings(ctx)
		lastEarningsSync = time.Now()
	} else {
		i.logger.Debug("skipping earnings sync (data is fresh)")
		lastEarningsSync = time.Now()
	}
	if i.shouldSyncFieldsNow() {
		i.syncFields(ctx)
		lastFieldSync = time.Now()
	} else {
		i.logger.Debug("skipping field sync (not scheduled time)")
	}
	i.sendPickReminders(ctx)
	lastReminderSync = time.Now()

	i.logger.Info("initial syncs complete, entering scheduled sync loop")
	for {
		select {
		case <-ctx.Done():
			i.logger.Info("ingester shutting down")
			return
		case <-ticker.C:
			now := time.Now()

			if now.Sub(lastTournamentSync) >= TournamentSyncInterval {
				i.syncTournaments(ctx)
				lastTournamentSync = now
			}

			leaderboardInterval := i.getLeaderboardInterval(ctx)
			if now.Sub(lastLeaderboardSync) >= leaderboardInterval {
				i.syncLeaderboards(ctx)
				lastLeaderboardSync = now
			}

			if i.isAnyTournamentInPlayHours(ctx) && now.Sub(lastScorecardSync) >= ScorecardSyncInterval {
				i.syncScorecards(ctx)
				lastScorecardSync = now
			}

			if i.shouldSyncFieldsNow() && now.Sub(lastFieldSync) >= 23*time.Hour {
				i.syncFields(ctx)
				lastFieldSync = now
			}

			if now.Sub(lastEarningsSync) >= EarningsCheckInterval {
				i.syncEarnings(ctx)
				lastEarningsSync = now
			}

			if i.shouldSyncPlayersNow() && now.Sub(lastPlayerSync) >= 23*time.Hour {
				i.syncPlayers(ctx)
				lastPlayerSync = now
			}

			if now.Sub(lastReminderSync) >= ReminderCheckInterval {
				i.sendPickReminders(ctx)
				lastReminderSync = now
			}
		}
	}
}

func (i *Ingester) getLeaderboardInterval(ctx context.Context) time.Duration {
	if i.isAnyTournamentInPlayHours(ctx) {
		return LeaderboardPlayInterval
	}
	return LeaderboardIdleInterval
}

func (i *Ingester) isAnyTournamentInPlayHours(ctx context.Context) bool {
	now := time.Now().UTC()

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.PickWindowClosesAtLT(now),
		).
		All(ctx)
	if err != nil {
		i.logger.Error("failed to query active tournaments for play hours check", "error", err)
		return false
	}

	for _, t := range tournaments {
		if i.isTournamentInPlayHours(t) {
			return true
		}
	}
	return false
}

func (i *Ingester) isTournamentInPlayHours(t *ent.Tournament) bool {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		i.logger.Error("failed to load timezone", "error", err)
		loc = time.UTC
	}
	now := time.Now().In(loc)

	day := now.Weekday()
	isPlayDay := day == time.Thursday || day == time.Friday || day == time.Saturday || day == time.Sunday

	hour := now.Hour()
	isPlayHour := hour >= PlayHoursStart && hour < PlayHoursEnd

	return isPlayDay && isPlayHour
}

func (i *Ingester) shouldSyncFieldsNow() bool {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		i.logger.Error("failed to load timezone", "error", err)
		return false
	}
	now := time.Now().In(loc)
	return now.Hour() == FieldSyncHour && now.Minute() < 5
}

func (i *Ingester) shouldSyncPlayersNow() bool {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		i.logger.Error("failed to load timezone", "error", err)
		return false
	}
	now := time.Now().In(loc)
	day := now.Weekday()
	isTargetDay := day == time.Tuesday || day == time.Wednesday || day == time.Thursday
	return isTargetDay && now.Hour() == PlayerSyncHour && now.Minute() < 5
}
