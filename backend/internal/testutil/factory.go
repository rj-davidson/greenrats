package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/fieldentry"
	"github.com/rj-davidson/greenrats/ent/leaguemembership"
	"github.com/rj-davidson/greenrats/ent/placement"
	"github.com/rj-davidson/greenrats/ent/season"
)

type Factory struct {
	t   *testing.T
	db  *ent.Client
	ctx context.Context
}

func NewFactory(t *testing.T, db *ent.Client) *Factory {
	t.Helper()
	return &Factory{
		t:   t,
		db:  db,
		ctx: context.Background(),
	}
}

func (f *Factory) WithContext(ctx context.Context) *Factory {
	f.ctx = ctx
	return f
}

func (f *Factory) getOrCreateSeason(year int) *ent.Season {
	f.t.Helper()

	existing, err := f.db.Season.Query().
		Where(season.YearEQ(year)).
		Only(f.ctx)
	if err == nil {
		return existing
	}
	if !ent.IsNotFound(err) {
		f.t.Fatalf("failed to query season: %v", err)
	}

	startDate := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, time.December, 31, 23, 59, 59, 0, time.UTC)

	created, err := f.db.Season.Create().
		SetYear(year).
		SetStartDate(startDate).
		SetEndDate(endDate).
		SetIsCurrent(true).
		Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to create season: %v", err)
	}
	return created
}

func (f *Factory) EnsureSeason(year int) *ent.Season {
	f.t.Helper()
	return f.getOrCreateSeason(year)
}

type UserOption func(*ent.UserCreate)

func WithWorkOSID(id string) UserOption {
	return func(uc *ent.UserCreate) {
		uc.SetWorkosID(id)
	}
}

func WithEmail(email string) UserOption {
	return func(uc *ent.UserCreate) {
		uc.SetEmail(email)
	}
}

func WithDisplayName(name string) UserOption {
	return func(uc *ent.UserCreate) {
		uc.SetDisplayName(name)
	}
}

func (f *Factory) CreateUser(opts ...UserOption) *ent.User {
	f.t.Helper()

	create := f.db.User.Create().
		SetWorkosID(fmt.Sprintf("user_%s", gofakeit.UUID())).
		SetEmail(gofakeit.Email()).
		SetDisplayName(gofakeit.Name())

	for _, opt := range opts {
		opt(create)
	}

	user, err := create.Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to create user: %v", err)
	}
	return user
}

func (f *Factory) CreateUserWithoutDisplayName(opts ...UserOption) *ent.User {
	f.t.Helper()

	create := f.db.User.Create().
		SetWorkosID(fmt.Sprintf("user_%s", gofakeit.UUID())).
		SetEmail(gofakeit.Email())

	for _, opt := range opts {
		opt(create)
	}

	user, err := create.Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to create user: %v", err)
	}
	return user
}

type TournamentOption func(*ent.TournamentCreate)

func WithTournamentName(name string) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetName(name)
	}
}

func WithStartDate(t time.Time) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetStartDate(t)
	}
}

func WithEndDate(t time.Time) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetEndDate(t)
	}
}

func WithChampion(golferID uuid.UUID) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetChampionID(golferID)
	}
}

func WithCourse(course string) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetCourse(course)
	}
}

func WithPurse(purse int) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetPurse(purse)
	}
}

func WithPgaTourID(id string) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetPgaTourID(id)
	}
}

var trackedSeasonYear *int

func WithSeasonYear(year int) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetSeasonYear(year)
		if trackedSeasonYear != nil {
			*trackedSeasonYear = year
		}
	}
}

func (f *Factory) CreateTournament(opts ...TournamentOption) *ent.Tournament {
	f.t.Helper()

	seasonYear := time.Now().Year()
	trackedSeasonYear = &seasonYear

	for _, opt := range opts {
		opt(f.db.Tournament.Create())
	}

	trackedSeasonYear = nil
	seasonEnt := f.getOrCreateSeason(seasonYear)

	startDate := time.Now().AddDate(0, 0, 7)
	endDate := startDate.AddDate(0, 0, 4)
	pickWindowOpens := startDate.AddDate(0, 0, -3)
	pickWindowCloses := startDate.Add(-1 * time.Hour)

	create := f.db.Tournament.Create().
		SetName(gofakeit.Company() + " Championship").
		SetStartDate(startDate).
		SetEndDate(endDate).
		SetSeasonYear(seasonYear).
		SetSeason(seasonEnt).
		SetCourse(gofakeit.City() + " Golf Club").
		SetPickWindowOpensAt(pickWindowOpens).
		SetPickWindowClosesAt(pickWindowCloses)

	for _, opt := range opts {
		opt(create)
	}

	t, err := create.Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to create tournament: %v", err)
	}
	return t
}

func WithPickWindow(opens, closes time.Time) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetPickWindowOpensAt(opens)
		tc.SetPickWindowClosesAt(closes)
	}
}

func (f *Factory) CreateUpcomingTournament(daysFromNow int, opts ...TournamentOption) *ent.Tournament {
	f.t.Helper()

	startDate := time.Now().AddDate(0, 0, daysFromNow)
	pickWindowCloses := startDate.Add(-1 * time.Hour)
	allOpts := append([]TournamentOption{
		WithStartDate(startDate),
		WithEndDate(startDate.AddDate(0, 0, 4)),
		WithPickWindow(startDate.AddDate(0, 0, -3), pickWindowCloses),
	}, opts...)

	return f.CreateTournament(allOpts...)
}

func (f *Factory) CreateActiveTournament(opts ...TournamentOption) *ent.Tournament {
	f.t.Helper()

	startDate := time.Now().AddDate(0, 0, -1)
	pickWindowCloses := time.Now().AddDate(0, 0, -2)
	allOpts := append([]TournamentOption{
		WithStartDate(startDate),
		WithEndDate(startDate.AddDate(0, 0, 4)),
		WithPickWindow(startDate.AddDate(0, 0, -3), pickWindowCloses),
	}, opts...)

	return f.CreateTournament(allOpts...)
}

func (f *Factory) CreateCompletedTournament(opts ...TournamentOption) *ent.Tournament {
	f.t.Helper()

	champion := f.CreateGolfer()

	startDate := time.Now().AddDate(0, 0, -10)
	pickWindowCloses := time.Now().AddDate(0, 0, -11)
	allOpts := append([]TournamentOption{
		WithStartDate(startDate),
		WithEndDate(startDate.AddDate(0, 0, 4)),
		WithPickWindow(startDate.AddDate(0, 0, -3), pickWindowCloses),
		WithChampion(champion.ID),
	}, opts...)

	return f.CreateTournament(allOpts...)
}

type GolferOption func(*ent.GolferCreate)

func WithGolferName(name string) GolferOption {
	return func(gc *ent.GolferCreate) {
		gc.SetName(name)
	}
}

func WithFirstName(name string) GolferOption {
	return func(gc *ent.GolferCreate) {
		gc.SetFirstName(name)
	}
}

func WithLastName(name string) GolferOption {
	return func(gc *ent.GolferCreate) {
		gc.SetLastName(name)
	}
}

func WithCountry(country string) GolferOption {
	return func(gc *ent.GolferCreate) {
		gc.SetCountry(country)
	}
}

func WithCountryCode(code string) GolferOption {
	return func(gc *ent.GolferCreate) {
		gc.SetCountryCode(code)
	}
}

func WithOWGR(rank int) GolferOption {
	return func(gc *ent.GolferCreate) {
		gc.SetOwgr(rank)
	}
}

func WithActive(active bool) GolferOption {
	return func(gc *ent.GolferCreate) {
		gc.SetActive(active)
	}
}

func (f *Factory) CreateGolfer(opts ...GolferOption) *ent.Golfer {
	f.t.Helper()

	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()

	create := f.db.Golfer.Create().
		SetFirstName(firstName).
		SetLastName(lastName).
		SetName(firstName + " " + lastName).
		SetCountry(gofakeit.Country()).
		SetCountryCode("USA").
		SetOwgr(gofakeit.Number(1, 500)).
		SetActive(true)

	for _, opt := range opts {
		opt(create)
	}

	g, err := create.Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to create golfer: %v", err)
	}
	return g
}

type LeagueOption func(*ent.LeagueCreate)

func WithLeagueName(name string) LeagueOption {
	return func(lc *ent.LeagueCreate) {
		lc.SetName(name)
	}
}

func WithLeagueCode(code string) LeagueOption {
	return func(lc *ent.LeagueCreate) {
		lc.SetCode(code)
	}
}

func WithLeagueSeasonYear(year int) LeagueOption {
	return func(lc *ent.LeagueCreate) {
		lc.SetSeasonYear(year)
	}
}

func WithJoiningEnabled(enabled bool) LeagueOption {
	return func(lc *ent.LeagueCreate) {
		lc.SetJoiningEnabled(enabled)
	}
}

func (f *Factory) CreateLeague(owner *ent.User, seasonYear int, opts ...LeagueOption) *ent.League {
	f.t.Helper()

	code := generateJoinCode()
	seasonEnt := f.getOrCreateSeason(seasonYear)

	create := f.db.League.Create().
		SetName(gofakeit.Company() + " League").
		SetCode(code).
		SetSeasonYear(seasonYear).
		SetSeason(seasonEnt).
		SetJoiningEnabled(true).
		SetCreatedBy(owner)

	for _, opt := range opts {
		opt(create)
	}

	league, err := create.Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to create league: %v", err)
	}

	_, err = f.db.LeagueMembership.Create().
		SetUser(owner).
		SetLeague(league).
		SetRole(leaguemembership.RoleOwner).
		Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to create owner membership: %v", err)
	}

	return league
}

func generateJoinCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 6)
	for i := range result {
		result[i] = charset[gofakeit.Number(0, len(charset)-1)]
	}
	return string(result)
}

func (f *Factory) AddUserToLeague(user *ent.User, league *ent.League) *ent.LeagueMembership {
	f.t.Helper()

	membership, err := f.db.LeagueMembership.Create().
		SetUser(user).
		SetLeague(league).
		SetRole(leaguemembership.RoleMember).
		Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to add user to league: %v", err)
	}
	return membership
}

type FieldEntryOption func(*ent.FieldEntryCreate)

func WithEntryStatusEnum(status fieldentry.EntryStatus) FieldEntryOption {
	return func(fec *ent.FieldEntryCreate) {
		fec.SetEntryStatus(status)
	}
}

func WithQualifier(q string) FieldEntryOption {
	return func(fec *ent.FieldEntryCreate) {
		fec.SetQualifier(q)
	}
}

func WithOwgrAtEntry(owgr int) FieldEntryOption {
	return func(fec *ent.FieldEntryCreate) {
		fec.SetOwgrAtEntry(owgr)
	}
}

func WithIsAmateur(amateur bool) FieldEntryOption {
	return func(fec *ent.FieldEntryCreate) {
		fec.SetIsAmateur(amateur)
	}
}

func (f *Factory) CreateFieldEntry(t *ent.Tournament, g *ent.Golfer, opts ...FieldEntryOption) *ent.FieldEntry {
	f.t.Helper()

	create := f.db.FieldEntry.Create().
		SetTournament(t).
		SetGolfer(g).
		SetEntryStatus(fieldentry.EntryStatusConfirmed)

	for _, opt := range opts {
		opt(create)
	}

	entry, err := create.Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to create field entry: %v", err)
	}
	return entry
}

type PlacementOption func(*ent.PlacementCreate)

func WithPosition(pos int) PlacementOption {
	return func(pc *ent.PlacementCreate) {
		pc.SetPositionNumeric(pos)
		pc.SetPosition(fmt.Sprintf("%d", pos))
	}
}

func WithPlacementStatus(status placement.Status) PlacementOption {
	return func(pc *ent.PlacementCreate) {
		pc.SetStatus(status)
	}
}

func WithParRelativeScore(score int) PlacementOption {
	return func(pc *ent.PlacementCreate) {
		pc.SetParRelativeScore(score)
	}
}

func WithEarnings(earnings int) PlacementOption {
	return func(pc *ent.PlacementCreate) {
		pc.SetEarnings(earnings)
	}
}

func WithCut(cut bool) PlacementOption {
	return func(pc *ent.PlacementCreate) {
		if cut {
			pc.SetStatus(placement.StatusCut)
			pc.SetPosition("CUT")
		}
	}
}

func (f *Factory) CreatePlacement(t *ent.Tournament, g *ent.Golfer, opts ...PlacementOption) *ent.Placement {
	f.t.Helper()

	create := f.db.Placement.Create().
		SetTournament(t).
		SetGolfer(g).
		SetStatus(placement.StatusFinished)

	for _, opt := range opts {
		opt(create)
	}

	p, err := create.Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to create placement: %v", err)
	}
	return p
}

type PickOption func(*ent.PickCreate)

func (f *Factory) CreatePick(user *ent.User, t *ent.Tournament, g *ent.Golfer, league *ent.League, opts ...PickOption) *ent.Pick {
	f.t.Helper()

	seasonEnt := f.getOrCreateSeason(t.SeasonYear)

	create := f.db.Pick.Create().
		SetUser(user).
		SetTournament(t).
		SetGolfer(g).
		SetLeague(league).
		SetSeasonYear(t.SeasonYear).
		SetSeason(seasonEnt)

	for _, opt := range opts {
		opt(create)
	}

	pick, err := create.Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to create pick: %v", err)
	}
	return pick
}

func (f *Factory) CreateGolfers(count int, opts ...GolferOption) []*ent.Golfer {
	f.t.Helper()

	golfers := make([]*ent.Golfer, count)
	for i := range count {
		golfers[i] = f.CreateGolfer(opts...)
	}
	return golfers
}

func (f *Factory) CreateTournamentField(t *ent.Tournament, golfers []*ent.Golfer, opts ...FieldEntryOption) []*ent.FieldEntry {
	f.t.Helper()

	entries := make([]*ent.FieldEntry, len(golfers))
	for i, g := range golfers {
		entries[i] = f.CreateFieldEntry(t, g, opts...)
	}
	return entries
}

func (f *Factory) RandomUUID() uuid.UUID {
	return uuid.Must(uuid.NewV4())
}

func CreateTournament(t *testing.T, db *ent.Client, name string, seasonYear int) *ent.Tournament {
	t.Helper()
	ctx := context.Background()

	seasonEnt := CreateSeason(t, db, seasonYear)

	startDate := time.Now().AddDate(0, 0, 7)
	endDate := startDate.AddDate(0, 0, 4)

	tournament, err := db.Tournament.Create().
		SetName(name).
		SetStartDate(startDate).
		SetEndDate(endDate).
		SetSeasonYear(seasonYear).
		SetSeason(seasonEnt).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create tournament: %v", err)
	}
	return tournament
}

func CreateGolfer(t *testing.T, db *ent.Client, name string, bdlID int) *ent.Golfer {
	t.Helper()
	ctx := context.Background()

	golfer, err := db.Golfer.Create().
		SetName(name).
		SetBdlID(bdlID).
		SetCountryCode("USA").
		SetActive(true).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create golfer: %v", err)
	}
	return golfer
}

func CreateFieldEntry(t *testing.T, db *ent.Client, tournamentID, golferID uuid.UUID) *ent.FieldEntry {
	t.Helper()
	ctx := context.Background()

	entry, err := db.FieldEntry.Create().
		SetTournamentID(tournamentID).
		SetGolferID(golferID).
		SetEntryStatus(fieldentry.EntryStatusConfirmed).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create field entry: %v", err)
	}
	return entry
}

func CreatePlacement(t *testing.T, db *ent.Client, tournamentID, golferID uuid.UUID) *ent.Placement {
	t.Helper()
	ctx := context.Background()

	p, err := db.Placement.Create().
		SetTournamentID(tournamentID).
		SetGolferID(golferID).
		SetStatus(placement.StatusFinished).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create placement: %v", err)
	}
	return p
}

func CreateSeason(t *testing.T, db *ent.Client, year int) *ent.Season {
	t.Helper()
	ctx := context.Background()

	existing, err := db.Season.Query().
		Where(season.YearEQ(year)).
		Only(ctx)
	if err == nil {
		return existing
	}
	if !ent.IsNotFound(err) {
		t.Fatalf("failed to query season: %v", err)
	}

	startDate := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, time.December, 31, 23, 59, 59, 0, time.UTC)

	created, err := db.Season.Create().
		SetYear(year).
		SetStartDate(startDate).
		SetEndDate(endDate).
		SetIsCurrent(true).
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create season: %v", err)
	}
	return created
}
