package sync

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
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
	db                   *ent.Client
	config               *config.Config
	ballDontLie          *balldontlie.Client
	syncService          *Service
	email                *email.Client
	logger               *slog.Logger
	scorecardSyncRunning atomic.Bool
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
		i.runSync(ctx, "sync_tournaments", i.syncTournaments)
		lastTournamentSync = time.Now()
	} else {
		lastTournamentSync = time.Now()
	}
	if i.shouldSyncPlayersNow() {
		i.runSync(ctx, "sync_players", i.syncPlayers)
		lastPlayerSync = time.Now()
	}
	if i.shouldSync(ctx, "leaderboards", i.getLeaderboardInterval(ctx)) {
		i.runSync(ctx, "sync_leaderboards", i.syncLeaderboards)
		lastLeaderboardSync = time.Now()
	} else {
		lastLeaderboardSync = time.Now()
	}
	if i.isAnyTournamentInPlayHours(ctx) {
		if i.shouldSync(ctx, "scorecards", ScorecardSyncInterval) {
			i.runSyncAsync(ctx, "sync_scorecards", &i.scorecardSyncRunning, i.syncScorecards)
			lastScorecardSync = time.Now()
		}
	}
	if i.shouldSync(ctx, "earnings", EarningsCheckInterval) {
		i.runSync(ctx, "sync_earnings", i.syncEarnings)
		lastEarningsSync = time.Now()
	} else {
		lastEarningsSync = time.Now()
	}
	if i.shouldSyncFieldsNow() {
		i.runSync(ctx, "sync_fields", i.syncFields)
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
				i.runSync(ctx, "sync_tournaments", i.syncTournaments)
				lastTournamentSync = now
			}

			leaderboardInterval := i.getLeaderboardInterval(ctx)
			if now.Sub(lastLeaderboardSync) >= leaderboardInterval {
				i.runSync(ctx, "sync_leaderboards", i.syncLeaderboards)
				lastLeaderboardSync = now
			}

			if i.isAnyTournamentInPlayHours(ctx) && now.Sub(lastScorecardSync) >= ScorecardSyncInterval {
				i.runSyncAsync(ctx, "sync_scorecards", &i.scorecardSyncRunning, i.syncScorecards)
				lastScorecardSync = now
			}

			if i.shouldSyncFieldsNow() && now.Sub(lastFieldSync) >= 23*time.Hour {
				i.runSync(ctx, "sync_fields", i.syncFields)
				lastFieldSync = now
			}

			if now.Sub(lastEarningsSync) >= EarningsCheckInterval {
				i.runSync(ctx, "sync_earnings", i.syncEarnings)
				lastEarningsSync = now
			}

			if i.shouldSyncPlayersNow() && now.Sub(lastPlayerSync) >= 23*time.Hour {
				i.runSync(ctx, "sync_players", i.syncPlayers)
				lastPlayerSync = now
			}

			if now.Sub(lastReminderSync) >= ReminderCheckInterval {
				i.sendPickReminders(ctx)
				lastReminderSync = now
			}

			if i.shouldSyncCoursesNow() && now.Sub(lastCourseSync) >= 6*24*time.Hour {
				i.runSync(ctx, "sync_courses", i.syncCourses)
				lastCourseSync = now
			}

			if i.shouldSyncGolferStatsNow() && now.Sub(lastGolferStatsSync) >= 6*24*time.Hour {
				i.runSync(ctx, "sync_golfer_stats", i.syncGolferSeasonStats)
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

	i.runSync(ctx, "sync_players", i.syncPlayers)
	i.runSync(ctx, "sync_tournaments", i.syncTournaments)
	i.runSync(ctx, "sync_courses", i.syncCourses)
	i.runSync(ctx, "sync_fields_init", i.syncAllFields)
	i.runSync(ctx, "sync_leaderboards_init", i.syncAllLeaderboards)
	i.runSync(ctx, "sync_rounds_init", i.syncAllRounds)
	i.runSync(ctx, "sync_earnings_init", i.syncAllEarnings)
	i.runSync(ctx, "sync_golfer_stats", i.syncGolferSeasonStats)

	i.logger.Info("initialization complete", "duration", time.Since(start))
}

func (i *Ingester) syncAllFields(ctx context.Context) error {
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
		return fmt.Errorf("query tournaments: %w", err)
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no tournaments with opened pick windows")
		return nil
	}

	synced := 0
	for _, t := range tournaments {
		if t.BdlID == nil {
			continue
		}

		hasField, err := i.tournamentHasField(ctx, t)
		if err != nil {
			if isContextError(err) {
				return fmt.Errorf("check field for %s: %w", t.Name, err)
			}
			continue
		}

		if hasField {
			continue
		}

		if err := i.syncTournamentField(ctx, t); err != nil {
			if isContextError(err) {
				return fmt.Errorf("sync field for %s: %w", t.Name, err)
			}
			continue
		}
		synced++
	}

	i.logger.Info("sync completed", "type", "fields_init", "duration", time.Since(start), "tournaments_synced", synced)
	return nil
}

func (i *Ingester) syncAllRounds(ctx context.Context) error {
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
		return fmt.Errorf("query tournaments: %w", err)
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no tournaments found for rounds initialization")
		return nil
	}

	synced := 0
	for _, t := range tournaments {
		if err := i.syncTournamentScorecards(ctx, t); err != nil {
			if isContextError(err) {
				return fmt.Errorf("sync scorecards for %s: %w", t.Name, err)
			}
			continue
		}
		synced++
	}

	i.logger.Info("sync completed", "type", "rounds_init", "duration", time.Since(start), "tournaments_synced", synced)
	return nil
}

func (i *Ingester) syncAllLeaderboards(ctx context.Context) error {
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
		return fmt.Errorf("query active tournaments: %w", err)
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no active tournaments found for leaderboard initialization")
		return nil
	}

	synced := 0
	for _, t := range tournaments {
		if err := i.syncTournamentLeaderboard(ctx, t); err != nil {
			if isContextError(err) {
				return fmt.Errorf("sync leaderboard for %s: %w", t.Name, err)
			}
			continue
		}
		synced++
	}

	i.logger.Info("sync completed", "type", "leaderboards_init", "duration", time.Since(start), "tournaments_synced", synced)
	return nil
}

func (i *Ingester) syncAllEarnings(ctx context.Context) error {
	start := time.Now()
	i.logger.Info("sync started", "type", "earnings_init")

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.HasChampion(),
			tournament.BdlIDNotNil(),
		).
		All(ctx)
	if err != nil {
		return fmt.Errorf("query completed tournaments: %w", err)
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no completed tournaments found for earnings initialization")
		return nil
	}

	synced := 0
	for _, t := range tournaments {
		hasEarnings, err := i.tournamentHasEarnings(ctx, t)
		if err != nil {
			if isContextError(err) {
				return fmt.Errorf("check earnings for %s: %w", t.Name, err)
			}
			continue
		}

		if hasEarnings {
			continue
		}

		if err := i.syncTournamentLeaderboard(ctx, t); err != nil {
			if isContextError(err) {
				return fmt.Errorf("sync leaderboard for %s: %w", t.Name, err)
			}
			continue
		}
		synced++
	}

	i.logger.Info("sync completed", "type", "earnings_init", "duration", time.Since(start), "tournaments_synced", synced)
	return nil
}
