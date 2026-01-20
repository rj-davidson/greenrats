package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/leaguemembership"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
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

func WithTournamentStatus(status tournament.Status) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetStatus(status)
	}
}

func WithSeasonYear(year int) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetSeasonYear(year)
	}
}

func WithCourse(course string) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetCourse(course)
	}
}

func WithLocation(location string) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetLocation(location)
	}
}

func WithPurse(purse int) TournamentOption {
	return func(tc *ent.TournamentCreate) {
		tc.SetPurse(purse)
	}
}

func (f *Factory) CreateTournament(opts ...TournamentOption) *ent.Tournament {
	f.t.Helper()

	startDate := time.Now().AddDate(0, 0, 7)
	endDate := startDate.AddDate(0, 0, 4)

	create := f.db.Tournament.Create().
		SetName(gofakeit.Company() + " Championship").
		SetStartDate(startDate).
		SetEndDate(endDate).
		SetStatus(tournament.StatusUpcoming).
		SetSeasonYear(time.Now().Year()).
		SetCourse(gofakeit.City() + " Golf Club").
		SetLocation(gofakeit.City() + ", " + gofakeit.StateAbr())

	for _, opt := range opts {
		opt(create)
	}

	t, err := create.Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to create tournament: %v", err)
	}
	return t
}

func (f *Factory) CreateUpcomingTournament(daysFromNow int, opts ...TournamentOption) *ent.Tournament {
	f.t.Helper()

	startDate := time.Now().AddDate(0, 0, daysFromNow)
	allOpts := append([]TournamentOption{
		WithStartDate(startDate),
		WithEndDate(startDate.AddDate(0, 0, 4)),
		WithTournamentStatus(tournament.StatusUpcoming),
	}, opts...)

	return f.CreateTournament(allOpts...)
}

func (f *Factory) CreateActiveTournament(opts ...TournamentOption) *ent.Tournament {
	f.t.Helper()

	startDate := time.Now().AddDate(0, 0, -1)
	allOpts := append([]TournamentOption{
		WithStartDate(startDate),
		WithEndDate(startDate.AddDate(0, 0, 4)),
		WithTournamentStatus(tournament.StatusActive),
	}, opts...)

	return f.CreateTournament(allOpts...)
}

func (f *Factory) CreateCompletedTournament(opts ...TournamentOption) *ent.Tournament {
	f.t.Helper()

	startDate := time.Now().AddDate(0, 0, -10)
	allOpts := append([]TournamentOption{
		WithStartDate(startDate),
		WithEndDate(startDate.AddDate(0, 0, 4)),
		WithTournamentStatus(tournament.StatusCompleted),
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

	create := f.db.League.Create().
		SetName(gofakeit.Company() + " League").
		SetCode(code).
		SetSeasonYear(seasonYear).
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

type TournamentEntryOption func(*ent.TournamentEntryCreate)

func WithPosition(pos int) TournamentEntryOption {
	return func(tec *ent.TournamentEntryCreate) {
		tec.SetPosition(pos)
	}
}

func WithCut(cut bool) TournamentEntryOption {
	return func(tec *ent.TournamentEntryCreate) {
		tec.SetCut(cut)
	}
}

func WithScore(score int) TournamentEntryOption {
	return func(tec *ent.TournamentEntryCreate) {
		tec.SetScore(score)
	}
}

func WithEarnings(earnings int) TournamentEntryOption {
	return func(tec *ent.TournamentEntryCreate) {
		tec.SetEarnings(earnings)
	}
}

func WithEntryStatus(status tournamententry.Status) TournamentEntryOption {
	return func(tec *ent.TournamentEntryCreate) {
		tec.SetStatus(status)
	}
}

func (f *Factory) CreateTournamentEntry(t *ent.Tournament, g *ent.Golfer, opts ...TournamentEntryOption) *ent.TournamentEntry {
	f.t.Helper()

	create := f.db.TournamentEntry.Create().
		SetTournament(t).
		SetGolfer(g).
		SetStatus(tournamententry.StatusPending)

	for _, opt := range opts {
		opt(create)
	}

	entry, err := create.Save(f.ctx)
	if err != nil {
		f.t.Fatalf("failed to create tournament entry: %v", err)
	}
	return entry
}

type PickOption func(*ent.PickCreate)

func (f *Factory) CreatePick(user *ent.User, t *ent.Tournament, g *ent.Golfer, league *ent.League, opts ...PickOption) *ent.Pick {
	f.t.Helper()

	create := f.db.Pick.Create().
		SetUser(user).
		SetTournament(t).
		SetGolfer(g).
		SetLeague(league).
		SetSeasonYear(t.SeasonYear)

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

func (f *Factory) CreateTournamentField(t *ent.Tournament, golfers []*ent.Golfer, opts ...TournamentEntryOption) []*ent.TournamentEntry {
	f.t.Helper()

	entries := make([]*ent.TournamentEntry, len(golfers))
	for i, g := range golfers {
		entries[i] = f.CreateTournamentEntry(t, g, opts...)
	}
	return entries
}

func (f *Factory) RandomUUID() uuid.UUID {
	return uuid.Must(uuid.NewV4())
}
