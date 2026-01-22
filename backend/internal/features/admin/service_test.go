package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestListTournamentField_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament(testutil.WithTournamentName("The Masters"))
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

	service := NewService(db)

	resp, err := service.ListTournamentField(ctx, tournament.ID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Entries, 2)

	entryMap := make(map[string]FieldEntryResponse)
	for _, e := range resp.Entries {
		entryMap[e.GolferName] = e
	}

	scheffler := entryMap["Scottie Scheffler"]
	assert.Equal(t, "confirmed", scheffler.EntryStatus)
	assert.Equal(t, "USA", scheffler.CountryCode)
	assert.NotNil(t, scheffler.OWGRAtEntry)
	assert.Equal(t, 1, *scheffler.OWGRAtEntry)

	mcilroy := entryMap["Rory McIlroy"]
	assert.Equal(t, "alternate", mcilroy.EntryStatus)
	assert.Equal(t, "NIR", mcilroy.CountryCode)
	assert.NotNil(t, mcilroy.Qualifier)
	assert.Equal(t, "Sponsor Exemption", *mcilroy.Qualifier)
}

func TestListTournamentField_Empty(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()

	service := NewService(db)

	resp, err := service.ListTournamentField(ctx, tournament.ID)

	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Entries)
}

func TestAddFieldEntry_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()
	golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"), testutil.WithCountryCode("USA"), testutil.WithOWGR(15))

	service := NewService(db)

	qualifier := "Past Champion"
	req := &AddFieldEntryRequest{
		GolferID:    golfer.ID,
		EntryStatus: "confirmed",
		Qualifier:   &qualifier,
	}

	resp, err := service.AddFieldEntry(ctx, tournament.ID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, golfer.ID, resp.GolferID)
	assert.Equal(t, "Tiger Woods", resp.GolferName)
	assert.Equal(t, "USA", resp.CountryCode)
	assert.Equal(t, "confirmed", resp.EntryStatus)
	assert.NotNil(t, resp.Qualifier)
	assert.Equal(t, "Past Champion", *resp.Qualifier)
	assert.NotNil(t, resp.OWGRAtEntry)
	assert.Equal(t, 15, *resp.OWGRAtEntry)
}

func TestAddFieldEntry_TournamentNotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"))
	randomID := factory.RandomUUID()

	service := NewService(db)

	req := &AddFieldEntryRequest{
		GolferID:    golfer.ID,
		EntryStatus: "confirmed",
	}

	resp, err := service.AddFieldEntry(ctx, randomID, req)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTournamentNotFound))
	assert.Nil(t, resp)
}

func TestAddFieldEntry_GolferNotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()
	randomID := factory.RandomUUID()

	service := NewService(db)

	req := &AddFieldEntryRequest{
		GolferID:    randomID,
		EntryStatus: "confirmed",
	}

	resp, err := service.AddFieldEntry(ctx, tournament.ID, req)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrGolferNotFound))
	assert.Nil(t, resp)
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

	service := NewService(db)

	req := &AddFieldEntryRequest{
		GolferID:    golfer.ID,
		EntryStatus: "alternate",
	}

	resp, err := service.AddFieldEntry(ctx, tournament.ID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, existingEntry.ID, resp.ID)
	assert.Equal(t, "confirmed", resp.EntryStatus)
}

func TestUpdateFieldEntry_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()
	golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"), testutil.WithCountryCode("USA"))
	entry := factory.CreateTournamentEntry(tournament, golfer,
		testutil.WithEntryStatusEnum(tournamententry.EntryStatusPending),
	)

	service := NewService(db)

	newStatus := "confirmed"
	newQualifier := "Past Champion"
	newOwgr := 10
	isAmateur := true

	req := &UpdateFieldEntryRequest{
		EntryStatus: &newStatus,
		Qualifier:   &newQualifier,
		OWGRAtEntry: &newOwgr,
		IsAmateur:   &isAmateur,
	}

	resp, err := service.UpdateFieldEntry(ctx, entry.ID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Tiger Woods", resp.GolferName)
	assert.Equal(t, "USA", resp.CountryCode)
	assert.Equal(t, "confirmed", resp.EntryStatus)
	assert.NotNil(t, resp.Qualifier)
	assert.Equal(t, "Past Champion", *resp.Qualifier)
	assert.NotNil(t, resp.OWGRAtEntry)
	assert.Equal(t, 10, *resp.OWGRAtEntry)
	assert.True(t, resp.IsAmateur)
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

	service := NewService(db)

	newStatus := "confirmed"
	req := &UpdateFieldEntryRequest{
		EntryStatus: &newStatus,
	}

	resp, err := service.UpdateFieldEntry(ctx, entry.ID, req)

	require.NoError(t, err)
	assert.Equal(t, "confirmed", resp.EntryStatus)
	assert.NotNil(t, resp.Qualifier)
	assert.Equal(t, "Original", *resp.Qualifier)
}

func TestUpdateFieldEntry_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	randomID := factory.RandomUUID()

	service := NewService(db)

	newStatus := "confirmed"
	req := &UpdateFieldEntryRequest{
		EntryStatus: &newStatus,
	}

	resp, err := service.UpdateFieldEntry(ctx, randomID, req)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrEntryNotFound))
	assert.Nil(t, resp)
}

func TestDeleteFieldEntry_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	factory := testutil.NewFactory(t, db)
	ctx := context.Background()

	tournament := factory.CreateTournament()
	golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"))
	entry := factory.CreateTournamentEntry(tournament, golfer)

	service := NewService(db)

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

	service := NewService(db)

	err := service.DeleteFieldEntry(ctx, randomID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrEntryNotFound))
}

func TestMapEntryStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected tournamententry.EntryStatus
	}{
		{"confirmed", tournamententry.EntryStatusConfirmed},
		{"alternate", tournamententry.EntryStatusAlternate},
		{"withdrawn", tournamententry.EntryStatusWithdrawn},
		{"pending", tournamententry.EntryStatusPending},
		{"unknown", tournamententry.EntryStatusConfirmed},
		{"", tournamententry.EntryStatusConfirmed},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapEntryStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
