package sync

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/course"
	"github.com/rj-davidson/greenrats/ent/coursehole"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/golferseason"
	"github.com/rj-davidson/greenrats/ent/holescore"
	"github.com/rj-davidson/greenrats/ent/round"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	db := testutil.NewTestDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(db, nil, logger)
}

func TestUpsertCourse_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	bdlCourse := &balldontlie.Course{
		ID:      123,
		Name:    "Augusta National Golf Club",
		City:    strPtr("Augusta"),
		State:   strPtr("GA"),
		Country: strPtr("USA"),
		Par:     intPtr(72),
		Yardage: strPtr("7475"),
	}

	created, err := svc.UpsertCourse(ctx, bdlCourse)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "Augusta National Golf Club", created.Name)
	assert.Equal(t, 123, *created.BdlID)
	assert.Equal(t, 72, *created.Par)
	assert.Equal(t, 7475, *created.Yardage)
	assert.Equal(t, "Augusta", *created.City)
	assert.Equal(t, "GA", *created.State)
	assert.Equal(t, "USA", *created.Country)
}

func TestUpsertCourse_Update(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	bdlCourse := &balldontlie.Course{
		ID:   123,
		Name: "Augusta National",
		Par:  intPtr(72),
	}

	_, err := svc.UpsertCourse(ctx, bdlCourse)
	require.NoError(t, err)

	bdlCourse.Name = "Augusta National Golf Club"
	bdlCourse.Yardage = strPtr("7510")

	updated, err := svc.UpsertCourse(ctx, bdlCourse)

	require.NoError(t, err)
	assert.Equal(t, "Augusta National Golf Club", updated.Name)
	assert.Equal(t, 7510, *updated.Yardage)
}

func TestUpsertCourseHole_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	bdlCourse := &balldontlie.Course{ID: 1, Name: "Test Course"}
	courseEntity, err := svc.UpsertCourse(ctx, bdlCourse)
	require.NoError(t, err)

	bdlHole := &balldontlie.CourseHole{
		Course:     balldontlie.CourseRef{ID: 1},
		HoleNumber: 12,
		Par:        3,
		Yardage:    intPtr(155),
	}

	err = svc.UpsertCourseHole(ctx, courseEntity.ID, bdlHole)
	require.NoError(t, err)

	holes, err := svc.db.CourseHole.Query().
		Where(coursehole.HasCourseWith(course.IDEQ(courseEntity.ID))).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, holes, 1)
	assert.Equal(t, 12, holes[0].HoleNumber)
	assert.Equal(t, 3, holes[0].Par)
	assert.Equal(t, 155, *holes[0].Yardage)
}

func TestUpsertCourseHole_Update(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	bdlCourse := &balldontlie.Course{ID: 1, Name: "Test Course"}
	courseEntity, err := svc.UpsertCourse(ctx, bdlCourse)
	require.NoError(t, err)

	bdlHole := &balldontlie.CourseHole{
		Course:     balldontlie.CourseRef{ID: 1},
		HoleNumber: 1,
		Par:        4,
		Yardage:    intPtr(445),
	}

	err = svc.UpsertCourseHole(ctx, courseEntity.ID, bdlHole)
	require.NoError(t, err)

	bdlHole.Par = 5
	bdlHole.Yardage = intPtr(520)

	err = svc.UpsertCourseHole(ctx, courseEntity.ID, bdlHole)
	require.NoError(t, err)

	holes, err := svc.db.CourseHole.Query().
		Where(coursehole.HasCourseWith(course.IDEQ(courseEntity.ID))).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, holes, 1)
	assert.Equal(t, 5, holes[0].Par)
	assert.Equal(t, 520, *holes[0].Yardage)
}

func TestUpsertRound_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournament := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	entry := testutil.CreateTournamentEntry(t, svc.db, tournament.ID, golferEntity.ID)

	bdlRound := &balldontlie.PlayerRoundResult{
		Tournament:       balldontlie.TournamentRef{ID: 1},
		Player:           balldontlie.Player{ID: 1, DisplayName: "Scottie Scheffler"},
		RoundNumber:      1,
		Score:            intPtr(68),
		ParRelativeScore: intPtr(-4),
	}

	created, err := svc.UpsertRound(ctx, entry.ID, bdlRound)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, 1, created.RoundNumber)
	assert.Equal(t, 68, *created.Score)
	assert.Equal(t, -4, *created.ParRelativeScore)
}

func TestUpsertRound_Update(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournament := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	entry := testutil.CreateTournamentEntry(t, svc.db, tournament.ID, golferEntity.ID)

	bdlRound := &balldontlie.PlayerRoundResult{
		RoundNumber:      1,
		Score:            intPtr(68),
		ParRelativeScore: intPtr(-4),
	}

	_, err := svc.UpsertRound(ctx, entry.ID, bdlRound)
	require.NoError(t, err)

	bdlRound.Score = intPtr(70)
	bdlRound.ParRelativeScore = intPtr(-2)

	updated, err := svc.UpsertRound(ctx, entry.ID, bdlRound)

	require.NoError(t, err)
	assert.Equal(t, 70, *updated.Score)
	assert.Equal(t, -2, *updated.ParRelativeScore)
}

func TestUpsertHoleScore_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournament := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	entry := testutil.CreateTournamentEntry(t, svc.db, tournament.ID, golferEntity.ID)

	bdlRoundResult := &balldontlie.PlayerRoundResult{RoundNumber: 1}
	roundRecord, err := svc.UpsertRound(ctx, entry.ID, bdlRoundResult)
	require.NoError(t, err)

	bdlScorecard := &balldontlie.PlayerScorecard{
		RoundNumber: 1,
		HoleNumber:  12,
		Par:         3,
		Score:       intPtr(2),
	}

	err = svc.UpsertHoleScore(ctx, roundRecord.ID, bdlScorecard)
	require.NoError(t, err)

	scores, err := svc.db.HoleScore.Query().
		Where(holescore.HasRoundWith(round.IDEQ(roundRecord.ID))).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, scores, 1)
	assert.Equal(t, 12, scores[0].HoleNumber)
	assert.Equal(t, 3, scores[0].Par)
	assert.Equal(t, 2, *scores[0].Score)
}

func TestUpsertHoleScore_Update(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournament := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	entry := testutil.CreateTournamentEntry(t, svc.db, tournament.ID, golferEntity.ID)

	bdlRoundResult := &balldontlie.PlayerRoundResult{RoundNumber: 1}
	roundRecord, err := svc.UpsertRound(ctx, entry.ID, bdlRoundResult)
	require.NoError(t, err)

	bdlScorecard := &balldontlie.PlayerScorecard{
		HoleNumber: 1,
		Par:        4,
		Score:      intPtr(4),
	}

	err = svc.UpsertHoleScore(ctx, roundRecord.ID, bdlScorecard)
	require.NoError(t, err)

	bdlScorecard.Score = intPtr(3)

	err = svc.UpsertHoleScore(ctx, roundRecord.ID, bdlScorecard)
	require.NoError(t, err)

	scores, err := svc.db.HoleScore.Query().
		Where(holescore.HasRoundWith(round.IDEQ(roundRecord.ID))).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, scores, 1)
	assert.Equal(t, 3, *scores[0].Score)
}

func TestUpsertGolferSeasonStat_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	season := testutil.CreateSeason(t, svc.db, 2026)

	stat := &balldontlie.PlayerSeasonStat{
		StatName:  "Scoring Average",
		StatValue: 68.5,
	}

	err := svc.UpsertGolferSeasonStat(ctx, golferEntity.ID, season.ID, stat)
	require.NoError(t, err)

	gs, err := svc.db.GolferSeason.Query().
		Where(golferseason.HasGolferWith(golfer.IDEQ(golferEntity.ID))).
		Only(ctx)

	require.NoError(t, err)
	assert.Equal(t, 68.5, *gs.ScoringAvg)
}

func TestUpsertGolferSeasonStat_Update(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	season := testutil.CreateSeason(t, svc.db, 2026)

	stat := &balldontlie.PlayerSeasonStat{
		StatName:  "Scoring Average",
		StatValue: 68.5,
	}

	err := svc.UpsertGolferSeasonStat(ctx, golferEntity.ID, season.ID, stat)
	require.NoError(t, err)

	stat.StatName = "Driving Distance"
	stat.StatValue = 310.5

	err = svc.UpsertGolferSeasonStat(ctx, golferEntity.ID, season.ID, stat)
	require.NoError(t, err)

	gs, err := svc.db.GolferSeason.Query().
		Where(golferseason.HasGolferWith(golfer.IDEQ(golferEntity.ID))).
		Only(ctx)

	require.NoError(t, err)
	assert.Equal(t, 68.5, *gs.ScoringAvg)
	assert.Equal(t, 310.5, *gs.DrivingDistance)
}

func TestUpsertGolferSeasonStat_AllStats(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	season := testutil.CreateSeason(t, svc.db, 2026)

	stats := []balldontlie.PlayerSeasonStat{
		{StatName: "Scoring Average", StatValue: 68.5},
		{StatName: "Top 10 Finishes", StatValue: float64(10)},
		{StatName: "Cuts Made", StatValue: float64(18)},
		{StatName: "Events Played", StatValue: float64(20)},
		{StatName: "Wins", StatValue: float64(4)},
		{StatName: "Official Money", StatValue: float64(15000000)},
		{StatName: "Driving Distance", StatValue: 310.5},
		{StatName: "Driving Accuracy Percentage", StatValue: 62.5},
		{StatName: "Greens in Regulation Percentage", StatValue: 70.2},
		{StatName: "Putting Average", StatValue: 1.72},
		{StatName: "Scrambling", StatValue: 65.8},
	}

	for _, stat := range stats {
		err := svc.UpsertGolferSeasonStat(ctx, golferEntity.ID, season.ID, &stat)
		require.NoError(t, err)
	}

	gs, err := svc.db.GolferSeason.Query().
		Where(golferseason.HasGolferWith(golfer.IDEQ(golferEntity.ID))).
		Only(ctx)

	require.NoError(t, err)
	assert.Equal(t, 68.5, *gs.ScoringAvg)
	assert.Equal(t, 10, *gs.Top10s)
	assert.Equal(t, 18, *gs.CutsMade)
	assert.Equal(t, 20, *gs.EventsPlayed)
	assert.Equal(t, 4, *gs.Wins)
	assert.Equal(t, 15000000, *gs.Earnings)
	assert.Equal(t, 310.5, *gs.DrivingDistance)
	assert.Equal(t, 62.5, *gs.DrivingAccuracy)
	assert.Equal(t, 70.2, *gs.GirPct)
	assert.Equal(t, 1.72, *gs.PuttingAvg)
	assert.Equal(t, 65.8, *gs.ScramblingPct)
	assert.NotNil(t, gs.LastSyncedAt)
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
