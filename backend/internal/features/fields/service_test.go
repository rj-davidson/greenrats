package fields

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/internal/external/pgatour"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestSyncTournamentField_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament(testutil.WithPgaTourID("R2025001"))
	golfer1 := factory.CreateGolfer(testutil.WithGolferName("Scottie Scheffler"))
	golfer2 := factory.CreateGolfer(testutil.WithGolferName("Rory McIlroy"))

	mockClient := pgatour.NewMockClient()
	mockClient.On("GetTournamentField", mock.Anything, "R2025001").Return([]pgatour.FieldEntry{
		{DisplayName: "Scottie Scheffler", IsAmateur: false},
		{DisplayName: "Rory McIlroy", IsAmateur: false},
	}, nil)

	service := NewService(db, mockClient, testLogger())

	result, err := service.SyncTournamentField(ctx, tournament.ID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, tournament.ID, result.TournamentID)
	assert.Equal(t, 2, result.TotalPlayers)
	assert.Equal(t, 2, result.MatchedPlayers)
	assert.Equal(t, 2, result.NewEntries)
	assert.Equal(t, 0, result.UpdatedEntries)
	assert.Empty(t, result.Errors)

	entries, err := db.TournamentEntry.Query().All(ctx)
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	golferIDs := make(map[string]bool)
	for _, e := range entries {
		g, _ := e.QueryGolfer().Only(ctx)
		golferIDs[g.Name] = true
	}
	assert.True(t, golferIDs["Scottie Scheffler"])
	assert.True(t, golferIDs["Rory McIlroy"])

	_ = golfer1
	_ = golfer2
	mockClient.AssertExpectations(t)
}

func TestSyncTournamentField_UpdatesExisting(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament(testutil.WithPgaTourID("R2025001"))
	golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"))
	factory.CreateTournamentEntry(tournament, golfer, testutil.WithEntryStatusEnum(tournamententry.EntryStatusPending))

	mockClient := pgatour.NewMockClient()
	mockClient.On("GetTournamentField", mock.Anything, "R2025001").Return([]pgatour.FieldEntry{
		{DisplayName: "Tiger Woods", IsAmateur: false},
	}, nil)

	service := NewService(db, mockClient, testLogger())

	result, err := service.SyncTournamentField(ctx, tournament.ID)

	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalPlayers)
	assert.Equal(t, 1, result.MatchedPlayers)
	assert.Equal(t, 0, result.NewEntries)
	assert.Equal(t, 1, result.UpdatedEntries)

	mockClient.AssertExpectations(t)
}

func TestSyncTournamentField_NoPgaTourID(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()

	mockClient := pgatour.NewMockClient()
	service := NewService(db, mockClient, testLogger())

	result, err := service.SyncTournamentField(ctx, tournament.ID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "has no PGA Tour ID")
	assert.Nil(t, result)
}

func TestSyncTournamentField_EmptyField(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament(testutil.WithPgaTourID("R2025001"))

	mockClient := pgatour.NewMockClient()
	mockClient.On("GetTournamentField", mock.Anything, "R2025001").Return([]pgatour.FieldEntry{}, nil)

	service := NewService(db, mockClient, testLogger())

	result, err := service.SyncTournamentField(ctx, tournament.ID)

	require.NoError(t, err)
	assert.Equal(t, 0, result.TotalPlayers)
	assert.Equal(t, 0, result.MatchedPlayers)

	mockClient.AssertExpectations(t)
}

func TestSyncTournamentField_PartialMatches(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament(testutil.WithPgaTourID("R2025001"))
	factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"))

	mockClient := pgatour.NewMockClient()
	mockClient.On("GetTournamentField", mock.Anything, "R2025001").Return([]pgatour.FieldEntry{
		{DisplayName: "Tiger Woods", IsAmateur: false},
		{DisplayName: "Unknown Player", IsAmateur: false},
	}, nil)

	service := NewService(db, mockClient, testLogger())

	result, err := service.SyncTournamentField(ctx, tournament.ID)

	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalPlayers)
	assert.Equal(t, 1, result.MatchedPlayers)
	assert.Equal(t, 1, result.NewEntries)

	mockClient.AssertExpectations(t)
}

func TestSyncTournamentField_ClientError(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament(testutil.WithPgaTourID("R2025001"))

	mockClient := pgatour.NewMockClient()
	mockClient.On("GetTournamentField", mock.Anything, "R2025001").Return(nil, errors.New("API error"))

	service := NewService(db, mockClient, testLogger())

	result, err := service.SyncTournamentField(ctx, tournament.ID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch field from PGA Tour")
	assert.Nil(t, result)

	mockClient.AssertExpectations(t)
}

func TestGetTournamentField_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()
	golfer1 := factory.CreateGolfer(testutil.WithGolferName("Scottie Scheffler"), testutil.WithCountryCode("USA"))
	golfer2 := factory.CreateGolfer(testutil.WithGolferName("Rory McIlroy"), testutil.WithCountryCode("NIR"))

	factory.CreateTournamentEntry(tournament, golfer1,
		testutil.WithEntryStatusEnum(tournamententry.EntryStatusConfirmed),
		testutil.WithOwgrAtEntry(1),
	)
	factory.CreateTournamentEntry(tournament, golfer2,
		testutil.WithEntryStatusEnum(tournamententry.EntryStatusAlternate),
		testutil.WithQualifier("Sponsor Exemption"),
	)

	mockClient := pgatour.NewMockClient()
	service := NewService(db, mockClient, testLogger())

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

	mockClient := pgatour.NewMockClient()
	service := NewService(db, mockClient, testLogger())

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

	mockClient := pgatour.NewMockClient()
	service := NewService(db, mockClient, testLogger())

	qualifier := "Past Champion"
	entry, err := service.AddFieldEntry(ctx, tournament.ID, golfer.ID, "confirmed", &qualifier)

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, tournamententry.EntryStatusConfirmed, entry.EntryStatus)
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
	existingEntry := factory.CreateTournamentEntry(tournament, golfer,
		testutil.WithEntryStatusEnum(tournamententry.EntryStatusConfirmed),
	)

	mockClient := pgatour.NewMockClient()
	service := NewService(db, mockClient, testLogger())

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

	mockClient := pgatour.NewMockClient()
	service := NewService(db, mockClient, testLogger())

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

	mockClient := pgatour.NewMockClient()
	service := NewService(db, mockClient, testLogger())

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
	entry := factory.CreateTournamentEntry(tournament, golfer,
		testutil.WithEntryStatusEnum(tournamententry.EntryStatusPending),
	)

	mockClient := pgatour.NewMockClient()
	service := NewService(db, mockClient, testLogger())

	newStatus := "confirmed"
	newQualifier := "Past Champion"
	newOwgr := 10
	isAmateur := true

	updated, err := service.UpdateFieldEntry(ctx, entry.ID, &newStatus, &newQualifier, &newOwgr, &isAmateur)

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, tournamententry.EntryStatusConfirmed, updated.EntryStatus)
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
	entry := factory.CreateTournamentEntry(tournament, golfer,
		testutil.WithEntryStatusEnum(tournamententry.EntryStatusPending),
		testutil.WithQualifier("Original"),
	)

	mockClient := pgatour.NewMockClient()
	service := NewService(db, mockClient, testLogger())

	newStatus := "confirmed"
	updated, err := service.UpdateFieldEntry(ctx, entry.ID, &newStatus, nil, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, tournamententry.EntryStatusConfirmed, updated.EntryStatus)
	assert.Equal(t, "Original", *updated.Qualifier)
}

func TestUpdateFieldEntry_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	randomID := factory.RandomUUID()

	mockClient := pgatour.NewMockClient()
	service := NewService(db, mockClient, testLogger())

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
	entry := factory.CreateTournamentEntry(tournament, golfer)

	mockClient := pgatour.NewMockClient()
	service := NewService(db, mockClient, testLogger())

	err := service.DeleteFieldEntry(ctx, entry.ID)

	require.NoError(t, err)

	exists, _ := db.TournamentEntry.Query().Where().Exist(ctx)
	assert.False(t, exists)
}

func TestDeleteFieldEntry_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	randomID := factory.RandomUUID()

	mockClient := pgatour.NewMockClient()
	service := NewService(db, mockClient, testLogger())

	err := service.DeleteFieldEntry(ctx, randomID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete entry")
}
