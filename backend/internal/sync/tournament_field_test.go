package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestUpsertTournamentFieldEntry_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

	bdlField := &balldontlie.TournamentField{
		ID:          1,
		Tournament:  balldontlie.Tournament{ID: 1, Name: "Masters"},
		Player:      balldontlie.Player{ID: 1, DisplayName: "Scottie Scheffler"},
		EntryStatus: "Committed",
		Qualifier:   "Winner",
		OWGR:        intPtr(1),
		IsAmateur:   false,
	}

	err := svc.UpsertTournamentFieldEntry(ctx, tournamentEntity, bdlField)
	require.NoError(t, err)

	entries, err := svc.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			tournamententry.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
		).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, tournamententry.EntryStatusConfirmed, entries[0].EntryStatus)
	assert.Equal(t, "Winner", *entries[0].Qualifier)
	assert.Equal(t, 1, *entries[0].OwgrAtEntry)
	assert.False(t, entries[0].IsAmateur)
}

func TestUpsertTournamentFieldEntry_Update(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

	bdlField := &balldontlie.TournamentField{
		Player:      balldontlie.Player{ID: 1, DisplayName: "Scottie Scheffler"},
		EntryStatus: "Committed",
		Qualifier:   "Exemption",
		OWGR:        intPtr(2),
		IsAmateur:   false,
	}

	err := svc.UpsertTournamentFieldEntry(ctx, tournamentEntity, bdlField)
	require.NoError(t, err)

	bdlField.EntryStatus = "Withdrawn"
	bdlField.OWGR = intPtr(1)

	err = svc.UpsertTournamentFieldEntry(ctx, tournamentEntity, bdlField)
	require.NoError(t, err)

	entries, err := svc.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			tournamententry.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
		).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, tournamententry.EntryStatusWithdrawn, entries[0].EntryStatus)
	assert.Equal(t, 1, *entries[0].OwgrAtEntry)
}

func TestUpsertTournamentFieldEntry_Amateur(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	testutil.CreateGolfer(t, svc.db, "Neal Shipley", 999)

	bdlField := &balldontlie.TournamentField{
		Player:      balldontlie.Player{ID: 999, DisplayName: "Neal Shipley"},
		EntryStatus: "Confirmed",
		Qualifier:   "US Amateur Champion",
		OWGR:        intPtr(250),
		IsAmateur:   true,
	}

	err := svc.UpsertTournamentFieldEntry(ctx, tournamentEntity, bdlField)
	require.NoError(t, err)

	entries, err := svc.db.TournamentEntry.Query().
		Where(tournamententry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID))).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.True(t, entries[0].IsAmateur)
	assert.Equal(t, "US Amateur Champion", *entries[0].Qualifier)
}

func TestUpsertTournamentFieldEntry_GolferNotFound(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)

	bdlField := &balldontlie.TournamentField{
		Player:      balldontlie.Player{ID: 9999, DisplayName: "Unknown Player"},
		EntryStatus: "Committed",
	}

	err := svc.UpsertTournamentFieldEntry(ctx, tournamentEntity, bdlField)
	require.NoError(t, err)

	entries, err := svc.db.TournamentEntry.Query().
		Where(tournamententry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID))).
		All(ctx)

	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestMapEntryStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected tournamententry.EntryStatus
	}{
		{"Committed", tournamententry.EntryStatusConfirmed},
		{"committed", tournamententry.EntryStatusConfirmed},
		{"Confirmed", tournamententry.EntryStatusConfirmed},
		{"Alternate", tournamententry.EntryStatusAlternate},
		{"alternate", tournamententry.EntryStatusAlternate},
		{"Withdrawn", tournamententry.EntryStatusWithdrawn},
		{"WD", tournamententry.EntryStatusWithdrawn},
		{"wd", tournamententry.EntryStatusWithdrawn},
		{"Unknown", tournamententry.EntryStatusPending},
		{"", tournamententry.EntryStatusPending},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapEntryStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
