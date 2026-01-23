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
)

const (
	TournamentSyncInterval  = 24 * time.Hour
	LeaderboardPlayInterval = 5 * time.Minute
	LeaderboardIdleInterval = 1 * time.Hour
	ScorecardSyncInterval   = 10 * time.Minute
	ReminderCheckInterval   = 1 * time.Hour
	EarningsCheckInterval   = 6 * time.Hour
	GolferStatsSyncInterval = 7 * 24 * time.Hour
	SchedulerTickInterval   = 1 * time.Minute
	FieldSyncHour           = 9
	PlayerSyncHour          = 21
	CourseSyncHour          = 3
	CourseSyncDay           = time.Sunday
	GolferStatsSyncDay      = time.Monday
	PlayHoursStart          = 8
	PlayHoursEnd            = 20
)

type Ingester struct {
	db          *ent.Client
	config      *config.Config
	ballDontLie *balldontlie.Client
	syncService *Service
	email       *email.Client
	logger      *slog.Logger
}

func NewIngester(
	db *ent.Client,
	cfg *config.Config,
	ballDontLie *balldontlie.Client,
	syncService *Service,
	emailClient *email.Client,
	logger *slog.Logger,
) *Ingester {
	return &Ingester{
		db:          db,
		config:      cfg,
		ballDontLie: ballDontLie,
		syncService: syncService,
		email:       emailClient,
		logger:      logger,
	}
}

func (i *Ingester) Run(ctx context.Context) {
	if i.needsInitialization(ctx) {
		i.initialize(ctx)
	}

	ticker := time.NewTicker(SchedulerTickInterval)
	defer ticker.Stop()

	var lastTournamentSync, lastLeaderboardSync, lastScorecardSync time.Time
	var lastFieldSync, lastEarningsSync, lastPlayerSync, lastReminderSync time.Time
	var lastCourseSync, lastGolferStatsSync time.Time

	i.logger.Info("ingester started, checking for required initial syncs")
	if i.shouldSync(ctx, "tournaments", TournamentSyncInterval) {
		i.syncTournaments(ctx)
		lastTournamentSync = time.Now()
	} else {
		lastTournamentSync = time.Now()
	}
	if i.shouldSyncPlayersNow() {
		i.syncPlayers(ctx)
		lastPlayerSync = time.Now()
	}
	if i.shouldSync(ctx, "leaderboards", i.getLeaderboardInterval(ctx)) {
		i.syncLeaderboards(ctx)
		lastLeaderboardSync = time.Now()
	} else {
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
		lastEarningsSync = time.Now()
	}
	if i.shouldSyncFieldsNow() {
		i.syncFields(ctx)
		lastFieldSync = time.Now()
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

			if i.shouldSyncCoursesNow() && now.Sub(lastCourseSync) >= 6*24*time.Hour {
				i.syncCourses(ctx)
				lastCourseSync = now
			}

			if i.shouldSyncGolferStatsNow() && now.Sub(lastGolferStatsSync) >= 6*24*time.Hour {
				i.syncGolferSeasonStats(ctx)
				lastGolferStatsSync = now
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

func (i *Ingester) shouldSyncCoursesNow() bool {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		i.logger.Error("failed to load timezone", "error", err)
		return false
	}
	now := time.Now().In(loc)
	return now.Weekday() == CourseSyncDay && now.Hour() == CourseSyncHour && now.Minute() < 5
}

func (i *Ingester) shouldSyncGolferStatsNow() bool {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		i.logger.Error("failed to load timezone", "error", err)
		return false
	}
	now := time.Now().In(loc)
	return now.Weekday() == GolferStatsSyncDay && now.Hour() == 4 && now.Minute() < 5
}

func (i *Ingester) needsInitialization(ctx context.Context) bool {
	count, err := i.db.Season.Query().Count(ctx)
	if err != nil {
		i.logger.Error("failed to check season count", "error", err)
		return false
	}
	return count == 0
}

func (i *Ingester) initialize(ctx context.Context) {
	i.logger.Info("running initialization - fresh database detected")
	start := time.Now()

	i.syncTournaments(ctx)
	i.syncPlayers(ctx)
	i.syncCourses(ctx)
	i.syncAllFields(ctx)
	i.syncAllLeaderboards(ctx)
	i.syncAllRounds(ctx)
	i.syncAllEarnings(ctx)
	i.syncGolferSeasonStats(ctx)

	i.logger.Info("initialization complete", "duration", time.Since(start))
}

func (i *Ingester) syncAllFields(ctx context.Context) {
	start := time.Now()
	i.logger.Info("sync started", "type", "fields_init")

	now := time.Now().UTC()
	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.PickWindowOpensAtLTE(now),
			tournament.BdlIDNotNil(),
		).
		All(ctx)
	if err != nil {
		i.logger.Error("failed to query tournaments for field initialization", "error", err)
		return
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no tournaments with opened pick windows")
		return
	}

	synced := 0
	for _, t := range tournaments {
		if t.BdlID == nil {
			continue
		}

		hasField, err := i.tournamentHasField(ctx, t)
		if err != nil {
			i.logger.Error("failed to check field", "tournament", t.Name, "error", err)
			continue
		}

		if hasField {
			continue
		}

		if err := i.syncTournamentField(ctx, t); err != nil {
			i.logger.Error("failed to sync field", "tournament", t.Name, "error", err)
			continue
		}
		synced++
	}

	i.logger.Info("sync completed", "type", "fields_init", "duration", time.Since(start), "tournaments_synced", synced)
}

func (i *Ingester) syncAllRounds(ctx context.Context) {
	start := time.Now()
	i.logger.Info("sync started", "type", "rounds_init")

	now := time.Now().UTC()
	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.PickWindowClosesAtLT(now),
			tournament.BdlIDNotNil(),
		).
		All(ctx)
	if err != nil {
		i.logger.Error("failed to query tournaments for rounds initialization", "error", err)
		return
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no tournaments found for rounds initialization")
		return
	}

	synced := 0
	for _, t := range tournaments {
		if err := i.syncTournamentScorecards(ctx, t); err != nil {
			i.logger.Error("failed to sync rounds", "tournament", t.Name, "error", err)
			continue
		}
		synced++
	}

	i.logger.Info("sync completed", "type", "rounds_init", "duration", time.Since(start), "tournaments_synced", synced)
}

func (i *Ingester) syncAllLeaderboards(ctx context.Context) {
	start := time.Now()
	i.logger.Info("sync started", "type", "leaderboards_init")

	now := time.Now().UTC()
	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.PickWindowClosesAtLT(now),
			tournament.BdlIDNotNil(),
		).
		All(ctx)
	if err != nil {
		i.logger.Error("failed to query active tournaments for initialization", "error", err)
		return
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no active tournaments found for leaderboard initialization")
		return
	}

	synced := 0
	for _, t := range tournaments {
		if err := i.syncTournamentLeaderboard(ctx, t); err != nil {
			i.logger.Error("failed to sync leaderboard", "tournament", t.Name, "error", err)
			continue
		}
		synced++
	}

	i.logger.Info("sync completed", "type", "leaderboards_init", "duration", time.Since(start), "tournaments_synced", synced)
}

func (i *Ingester) syncAllEarnings(ctx context.Context) {
	start := time.Now()
	i.logger.Info("sync started", "type", "earnings_init")

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.HasChampion(),
			tournament.BdlIDNotNil(),
		).
		All(ctx)
	if err != nil {
		i.logger.Error("failed to query completed tournaments for initialization", "error", err)
		return
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no completed tournaments found for earnings initialization")
		return
	}

	synced := 0
	for _, t := range tournaments {
		hasEarnings, err := i.tournamentHasEarnings(ctx, t)
		if err != nil {
			i.logger.Error("failed to check earnings", "tournament", t.Name, "error", err)
			continue
		}

		if hasEarnings {
			continue
		}

		if err := i.syncTournamentLeaderboard(ctx, t); err != nil {
			i.logger.Error("failed to sync earnings", "tournament", t.Name, "error", err)
			continue
		}
		synced++
	}

	i.logger.Info("sync completed", "type", "earnings_init", "duration", time.Since(start), "tournaments_synced", synced)
}
