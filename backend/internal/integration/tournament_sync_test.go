package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/internal/external/pgatour"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func upcomingStartDate() time.Time {
	return time.Now().AddDate(0, 0, 7)
}

func TestTournamentSync_SetsPGATourID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := testutil.NewTestDB(t)

	t.Run("creates tournament with PGA Tour ID from mapping", func(t *testing.T) {
		bdlID := 8
		expectedPGAID := "R2026002"
		startDate := upcomingStartDate()

		created, err := db.Tournament.Create().
			SetBdlID(bdlID).
			SetName("The American Express").
			SetStartDate(startDate).
			SetEndDate(startDate.Add(4 * 24 * time.Hour)).
			SetSeasonYear(2026).
			SetNillablePgaTourID(getPGATourIDPtr(bdlID)).
			Save(ctx)

		require.NoError(t, err)
		require.NotNil(t, created.PgaTourID)
		assert.Equal(t, expectedPGAID, *created.PgaTourID)
	})

	t.Run("creates tournament without PGA Tour ID when not in mapping", func(t *testing.T) {
		bdlID := 9999
		startDate := upcomingStartDate()

		created, err := db.Tournament.Create().
			SetBdlID(bdlID).
			SetName("Unknown Tournament").
			SetStartDate(startDate).
			SetEndDate(startDate.Add(4 * 24 * time.Hour)).
			SetSeasonYear(2026).
			SetNillablePgaTourID(getPGATourIDPtr(bdlID)).
			Save(ctx)

		require.NoError(t, err)
		assert.Nil(t, created.PgaTourID)
	})

	t.Run("updates tournament with PGA Tour ID when missing", func(t *testing.T) {
		bdlID := 10
		expectedPGAID := "R2026003"
		startDate := upcomingStartDate()

		existing, err := db.Tournament.Create().
			SetBdlID(bdlID).
			SetName("WM Phoenix Open").
			SetStartDate(startDate).
			SetEndDate(startDate.Add(4 * 24 * time.Hour)).
			SetSeasonYear(2026).
			Save(ctx)
		require.NoError(t, err)
		assert.Nil(t, existing.PgaTourID)

		updated, err := existing.Update().
			SetNillablePgaTourID(getPGATourIDPtr(bdlID)).
			Save(ctx)

		require.NoError(t, err)
		require.NotNil(t, updated.PgaTourID)
		assert.Equal(t, expectedPGAID, *updated.PgaTourID)
	})

	t.Run("does not overwrite existing PGA Tour ID", func(t *testing.T) {
		bdlID := 11
		manualPGAID := "MANUAL123"
		startDate := upcomingStartDate()

		existing, err := db.Tournament.Create().
			SetBdlID(bdlID).
			SetName("AT&T Pebble Beach Pro-Am").
			SetStartDate(startDate).
			SetEndDate(startDate.Add(4 * 24 * time.Hour)).
			SetSeasonYear(2026).
			SetPgaTourID(manualPGAID).
			Save(ctx)
		require.NoError(t, err)

		shouldUpdate := existing.PgaTourID == nil || *existing.PgaTourID == ""
		assert.False(t, shouldUpdate, "should not update when PGA Tour ID already set")
		assert.Equal(t, manualPGAID, *existing.PgaTourID)
	})

	t.Run("all mapped BDL IDs have PGA Tour IDs", func(t *testing.T) {
		mappedIDs := []int{7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42}

		for _, bdlID := range mappedIDs {
			pgaID := pgatour.GetPGATourID(bdlID)
			assert.NotEmpty(t, pgaID, "BDL ID %d should have a PGA Tour ID mapping", bdlID)
		}
	})
}

func getPGATourIDPtr(bdlID int) *string {
	pgaID := pgatour.GetPGATourID(bdlID)
	if pgaID == "" {
		return nil
	}
	return &pgaID
}

func upsertTournament(ctx context.Context, db *ent.Client, bdlID int, name string) (*ent.Tournament, error) {
	existing, err := db.Tournament.Query().
		Where(tournament.BdlID(bdlID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		startDate := upcomingStartDate()
		builder := db.Tournament.Create().
			SetBdlID(bdlID).
			SetName(name).
			SetStartDate(startDate).
			SetEndDate(startDate.Add(4 * 24 * time.Hour)).
			SetSeasonYear(2026)

		if pgaTourID := pgatour.GetPGATourID(bdlID); pgaTourID != "" {
			builder.SetPgaTourID(pgaTourID)
		}

		return builder.Save(ctx)
	}

	if err != nil {
		return nil, err
	}

	updater := existing.Update().SetName(name)

	if existing.PgaTourID == nil || *existing.PgaTourID == "" {
		if pgaTourID := pgatour.GetPGATourID(bdlID); pgaTourID != "" {
			updater.SetPgaTourID(pgaTourID)
		}
	}

	return updater.Save(ctx)
}

func TestUpsertTournament(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := testutil.NewTestDB(t)

	t.Run("upsert creates with PGA Tour ID", func(t *testing.T) {
		bdlID := 7
		tournament, err := upsertTournament(ctx, db, bdlID, "Sony Open in Hawaii")

		require.NoError(t, err)
		require.NotNil(t, tournament.PgaTourID)
		assert.Equal(t, "R2026006", *tournament.PgaTourID)
	})

	t.Run("upsert updates existing without PGA Tour ID", func(t *testing.T) {
		bdlID := 20
		startDate := upcomingStartDate()

		_, err := db.Tournament.Create().
			SetBdlID(bdlID).
			SetName("Masters Tournament").
			SetStartDate(startDate).
			SetEndDate(startDate.Add(4 * 24 * time.Hour)).
			SetSeasonYear(2026).
			Save(ctx)
		require.NoError(t, err)

		updated, err := upsertTournament(ctx, db, bdlID, "Masters Tournament")

		require.NoError(t, err)
		require.NotNil(t, updated.PgaTourID)
		assert.Equal(t, "R2026014", *updated.PgaTourID)
	})

	t.Run("upsert preserves manually set PGA Tour ID", func(t *testing.T) {
		bdlID := 31
		manualID := "CUSTOM_US_OPEN"
		startDate := upcomingStartDate()

		_, err := db.Tournament.Create().
			SetBdlID(bdlID).
			SetName("U.S. Open").
			SetStartDate(startDate).
			SetEndDate(startDate.Add(4 * 24 * time.Hour)).
			SetSeasonYear(2026).
			SetPgaTourID(manualID).
			Save(ctx)
		require.NoError(t, err)

		updated, err := upsertTournament(ctx, db, bdlID, "U.S. Open")

		require.NoError(t, err)
		require.NotNil(t, updated.PgaTourID)
		assert.Equal(t, manualID, *updated.PgaTourID, "should preserve manually set ID")
	})
}
