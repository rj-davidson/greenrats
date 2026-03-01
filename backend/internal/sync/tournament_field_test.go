package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/fieldentry"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/tournament"
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
		Qualifier:   strPtr("Winner"),
		OWGR:        intPtr(1),
		IsAmateur:   false,
	}

	err := svc.UpsertFieldEntry(ctx, tournamentEntity, bdlField)
	require.NoError(t, err)

	entries, err := svc.db.FieldEntry.Query().
		Where(
			fieldentry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			fieldentry.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
		).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, fieldentry.EntryStatusConfirmed, entries[0].EntryStatus)
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
		Qualifier:   strPtr("Exemption"),
		OWGR:        intPtr(2),
		IsAmateur:   false,
	}

	err := svc.UpsertFieldEntry(ctx, tournamentEntity, bdlField)
	require.NoError(t, err)

	bdlField.EntryStatus = "Withdrawn"
	bdlField.OWGR = intPtr(1)

	err = svc.UpsertFieldEntry(ctx, tournamentEntity, bdlField)
	require.NoError(t, err)

	entries, err := svc.db.FieldEntry.Query().
		Where(
			fieldentry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			fieldentry.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
		).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, fieldentry.EntryStatusWithdrawn, entries[0].EntryStatus)
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
		Qualifier:   strPtr("US Amateur Champion"),
		OWGR:        intPtr(250),
		IsAmateur:   true,
	}

	err := svc.UpsertFieldEntry(ctx, tournamentEntity, bdlField)
	require.NoError(t, err)

	entries, err := svc.db.FieldEntry.Query().
		Where(fieldentry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID))).
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

	err := svc.UpsertFieldEntry(ctx, tournamentEntity, bdlField)
	require.NoError(t, err)

	entries, err := svc.db.FieldEntry.Query().
		Where(fieldentry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID))).
		All(ctx)

	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestMapFieldEntryStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected fieldentry.EntryStatus
	}{
		{"Committed", fieldentry.EntryStatusConfirmed},
		{"committed", fieldentry.EntryStatusConfirmed},
		{"Confirmed", fieldentry.EntryStatusConfirmed},
		{"Alternate", fieldentry.EntryStatusAlternate},
		{"alternate", fieldentry.EntryStatusAlternate},
		{"Withdrawn", fieldentry.EntryStatusWithdrawn},
		{"WD", fieldentry.EntryStatusWithdrawn},
		{"wd", fieldentry.EntryStatusWithdrawn},
		{"Unknown", fieldentry.EntryStatusPending},
		{"", fieldentry.EntryStatusPending},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapFieldEntryStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
