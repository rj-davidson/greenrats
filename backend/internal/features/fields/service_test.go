package fields

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/fieldentry"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestGetTournamentField_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()
	golfer1 := factory.CreateGolfer(testutil.WithGolferName("Scottie Scheffler"), testutil.WithCountryCode("USA"))
	golfer2 := factory.CreateGolfer(testutil.WithGolferName("Rory McIlroy"), testutil.WithCountryCode("NIR"))

	factory.CreateFieldEntry(tournament, golfer1,
		testutil.WithEntryStatusEnum(fieldentry.EntryStatusConfirmed),
		testutil.WithOwgrAtEntry(1),
	)
	factory.CreateFieldEntry(tournament, golfer2,
		testutil.WithEntryStatusEnum(fieldentry.EntryStatusAlternate),
		testutil.WithQualifier("Sponsor Exemption"),
	)

	service := NewService(db, testLogger())

	entries, err := service.GetTournamentField(ctx, tournament.ID)

	require.NoError(t, err)
	assert.Len(t, entries, 2)

	entryMap := make(map[string]FieldEntry)
	for _, e := range entries {
		entryMap[e.GolferName] = e
	}

	assert.Equal(t, "confirmed", entryMap["Scottie Scheffler"].EntryStatus)
	assert.Equal(t, "USA", entryMap["Scottie Scheffler"].CountryCode)
	assert.NotNil(t, entryMap["Scottie Scheffler"].OWGRAtEntry)
	assert.Equal(t, 1, *entryMap["Scottie Scheffler"].OWGRAtEntry)

	assert.Equal(t, "alternate", entryMap["Rory McIlroy"].EntryStatus)
	assert.Equal(t, "NIR", entryMap["Rory McIlroy"].CountryCode)
	assert.NotNil(t, entryMap["Rory McIlroy"].Qualifier)
	assert.Equal(t, "Sponsor Exemption", *entryMap["Rory McIlroy"].Qualifier)
}

func TestGetTournamentField_Empty(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()

	service := NewService(db, testLogger())

	entries, err := service.GetTournamentField(ctx, tournament.ID)

	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestAddFieldEntry_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()
	golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"), testutil.WithOWGR(15))

	service := NewService(db, testLogger())

	qualifier := "Past Champion"
	entry, err := service.AddFieldEntry(ctx, tournament.ID, golfer.ID, "confirmed", &qualifier)

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, fieldentry.EntryStatusConfirmed, entry.EntryStatus)
	assert.NotNil(t, entry.Qualifier)
	assert.Equal(t, "Past Champion", *entry.Qualifier)
	assert.NotNil(t, entry.OwgrAtEntry)
	assert.Equal(t, 15, *entry.OwgrAtEntry)
}

func TestAddFieldEntry_AlreadyExists(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()
	golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"))
	existingEntry := factory.CreateFieldEntry(tournament, golfer,
		testutil.WithEntryStatusEnum(fieldentry.EntryStatusConfirmed),
	)

	service := NewService(db, testLogger())

	entry, err := service.AddFieldEntry(ctx, tournament.ID, golfer.ID, "alternate", nil)

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, existingEntry.ID, entry.ID)
}

func TestAddFieldEntry_TournamentNotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"))
	randomID := factory.RandomUUID()

	service := NewService(db, testLogger())

	entry, err := service.AddFieldEntry(ctx, randomID, golfer.ID, "confirmed", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get tournament")
	assert.Nil(t, entry)
}

func TestAddFieldEntry_GolferNotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()
	randomID := factory.RandomUUID()

	service := NewService(db, testLogger())

	entry, err := service.AddFieldEntry(ctx, tournament.ID, randomID, "confirmed", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get golfer")
	assert.Nil(t, entry)
}

func TestUpdateFieldEntry_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()
	golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"))
	entry := factory.CreateFieldEntry(tournament, golfer,
		testutil.WithEntryStatusEnum(fieldentry.EntryStatusPending),
	)

	service := NewService(db, testLogger())

	newStatus := "confirmed"
	newQualifier := "Past Champion"
	newOwgr := 10
	isAmateur := true

	updated, err := service.UpdateFieldEntry(ctx, entry.ID, &newStatus, &newQualifier, &newOwgr, &isAmateur)

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, fieldentry.EntryStatusConfirmed, updated.EntryStatus)
	assert.Equal(t, "Past Champion", *updated.Qualifier)
	assert.Equal(t, 10, *updated.OwgrAtEntry)
	assert.True(t, updated.IsAmateur)
}

func TestUpdateFieldEntry_PartialUpdate(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()
	golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"))
	entry := factory.CreateFieldEntry(tournament, golfer,
		testutil.WithEntryStatusEnum(fieldentry.EntryStatusPending),
		testutil.WithQualifier("Original"),
	)

	service := NewService(db, testLogger())

	newStatus := "confirmed"
	updated, err := service.UpdateFieldEntry(ctx, entry.ID, &newStatus, nil, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, fieldentry.EntryStatusConfirmed, updated.EntryStatus)
	assert.Equal(t, "Original", *updated.Qualifier)
}

func TestUpdateFieldEntry_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	randomID := factory.RandomUUID()

	service := NewService(db, testLogger())

	newStatus := "confirmed"
	updated, err := service.UpdateFieldEntry(ctx, randomID, &newStatus, nil, nil, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get entry")
	assert.Nil(t, updated)
}

func TestDeleteFieldEntry_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()
	golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"))
	entry := factory.CreateFieldEntry(tournament, golfer)

	service := NewService(db, testLogger())

	err := service.DeleteFieldEntry(ctx, entry.ID)

	require.NoError(t, err)

	exists, _ := db.FieldEntry.Query().Where().Exist(ctx)
	assert.False(t, exists)
}

func TestDeleteFieldEntry_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	randomID := factory.RandomUUID()

	service := NewService(db, testLogger())

	err := service.DeleteFieldEntry(ctx, randomID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete entry")
}
