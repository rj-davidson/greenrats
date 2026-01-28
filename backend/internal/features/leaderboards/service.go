package leaderboards

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/pick"
	"github.com/rj-davidson/greenrats/ent/placement"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/internal/features/tournaments"
)

type Service struct {
	db                *ent.Client
	tournamentService *tournaments.Service
}

func NewService(db *ent.Client, tournamentService *tournaments.Service) *Service {
	return &Service{db: db, tournamentService: tournamentService}
}

func (s *Service) GetLeagueLeaderboard(ctx context.Context, leagueID uuid.UUID, seasonYear int) (*LeagueLeaderboardResponse, error) {
	resp, err := s.GetLeagueStandings(ctx, leagueID, seasonYear, false, uuid.Nil)
	if err != nil {
		return nil, err
	}

	entries := make([]LeaderboardEntry, len(resp.Entries))
	for i, e := range resp.Entries {
		entries[i] = LeaderboardEntry{
			UserID:      e.UserID,
			DisplayName: e.UserDisplayName,
			Earnings:    e.TotalEarnings,
			PickCount:   e.PickCount,
		}
	}

	return &LeagueLeaderboardResponse{
		Entries:    entries,
		Total:      resp.Total,
		SeasonYear: resp.SeasonYear,
	}, nil
}

func (s *Service) GetLeagueStandings(ctx context.Context, leagueID uuid.UUID, seasonYear int, includePicks bool, requestingUserID uuid.UUID) (*LeagueStandingsResponse, error) {
	entLeague, err := s.db.League.
		Query().
		Where(league.IDEQ(leagueID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("league not found")
		}
		return nil, fmt.Errorf("failed to get league: %w", err)
	}

	if seasonYear == 0 {
		seasonYear = entLeague.SeasonYear
	}

	var activeTournament *tournaments.Tournament
	if s.tournamentService != nil {
		activeTournament, err = s.tournamentService.GetActive(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get active tournament: %w", err)
		}
	}

	var isPickWindowClosed bool
	if activeTournament != nil && activeTournament.PickWindowClosesAt != nil {
		isPickWindowClosed = time.Now().UTC().After(*activeTournament.PickWindowClosesAt)
	}

	// If no active tournament, check for an upcoming tournament with open pick window
	var upcomingTournament *ent.Tournament
	if activeTournament == nil {
		now := time.Now().UTC()
		upcomingTournament, err = s.db.Tournament.Query().
			Where(
				tournament.Not(tournament.HasChampion()),
				tournament.PickWindowOpensAtLTE(now),
				tournament.PickWindowClosesAtGT(now),
				tournament.SeasonYearEQ(seasonYear),
			).
			Order(ent.Asc(tournament.FieldStartDate)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get upcoming tournament: %w", err)
		}
	}

	picks, err := s.db.Pick.
		Query().
		Where(
			pick.HasLeagueWith(league.IDEQ(leagueID)),
			pick.SeasonYearEQ(seasonYear),
		).
		WithUser().
		WithGolfer().
		WithTournament().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get picks: %w", err)
	}

	type pickKey struct {
		TournamentID uuid.UUID
		GolferID     uuid.UUID
	}
	placementKeys := make([]pickKey, 0, len(picks))
	for _, p := range picks {
		if p.Edges.Golfer != nil && p.Edges.Tournament != nil {
			placementKeys = append(placementKeys, pickKey{
				TournamentID: p.Edges.Tournament.ID,
				GolferID:     p.Edges.Golfer.ID,
			})
		}
	}

	placementMap := make(map[pickKey]*ent.Placement)
	if len(placementKeys) > 0 {
		tournamentIDs := make([]uuid.UUID, 0)
		golferIDs := make([]uuid.UUID, 0)
		seen := make(map[uuid.UUID]bool)
		for _, k := range placementKeys {
			if !seen[k.TournamentID] {
				tournamentIDs = append(tournamentIDs, k.TournamentID)
				seen[k.TournamentID] = true
			}
		}
		seen = make(map[uuid.UUID]bool)
		for _, k := range placementKeys {
			if !seen[k.GolferID] {
				golferIDs = append(golferIDs, k.GolferID)
				seen[k.GolferID] = true
			}
		}

		placements, err := s.db.Placement.
			Query().
			Where(
				placement.HasTournamentWith(tournament.IDIn(tournamentIDs...)),
				placement.HasGolferWith(golfer.IDIn(golferIDs...)),
			).
			WithTournament().
			WithGolfer().
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get placements: %w", err)
		}

		for _, pl := range placements {
			if pl.Edges.Tournament != nil && pl.Edges.Golfer != nil {
				key := pickKey{TournamentID: pl.Edges.Tournament.ID, GolferID: pl.Edges.Golfer.ID}
				placementMap[key] = pl
			}
		}
	}

	type userPickData struct {
		DisplayName string
		Earnings    int
		Picks       []PickHistory
		CurrentPick *CurrentPick
	}

	userData := make(map[uuid.UUID]*userPickData)

	for _, p := range picks {
		if p.Edges.User == nil {
			continue
		}

		userID := p.Edges.User.ID
		data, exists := userData[userID]
		if !exists {
			displayName := "Unknown"
			if p.Edges.User.DisplayName != nil {
				displayName = *p.Edges.User.DisplayName
			}
			data = &userPickData{
				DisplayName: displayName,
				Earnings:    0,
				Picks:       make([]PickHistory, 0),
			}
			userData[userID] = data
		}

		var pickEarnings int
		var position int
		var status string

		if p.Edges.Golfer != nil && p.Edges.Tournament != nil {
			key := pickKey{TournamentID: p.Edges.Tournament.ID, GolferID: p.Edges.Golfer.ID}
			if pl, ok := placementMap[key]; ok {
				pickEarnings = pl.Earnings
				data.Earnings += pickEarnings
				status = string(pl.Status)
				if pl.PositionNumeric != nil {
					position = *pl.PositionNumeric
				}
			}

			isActiveMatch := activeTournament != nil && p.Edges.Tournament.ID.String() == activeTournament.ID
			isUpcomingMatch := upcomingTournament != nil && p.Edges.Tournament.ID == upcomingTournament.ID
			if isActiveMatch || isUpcomingMatch {
				data.CurrentPick = &CurrentPick{
					TournamentID:   p.Edges.Tournament.ID,
					TournamentName: p.Edges.Tournament.Name,
					GolferID:       p.Edges.Golfer.ID,
					GolferName:     p.Edges.Golfer.Name,
				}
			}

			if includePicks {
				data.Picks = append(data.Picks, PickHistory{
					TournamentID:   p.Edges.Tournament.ID,
					TournamentName: p.Edges.Tournament.Name,
					GolferID:       p.Edges.Golfer.ID,
					GolferName:     p.Edges.Golfer.Name,
					Position:       position,
					Status:         status,
					Earnings:       pickEarnings,
				})
			}
		}
	}

	pickCountByUser := make(map[uuid.UUID]int)
	for _, p := range picks {
		if p.Edges.User != nil {
			pickCountByUser[p.Edges.User.ID]++
		}
	}

	entries := make([]StandingsEntry, 0, len(userData))
	for userID, data := range userData {
		entry := StandingsEntry{
			UserID:          userID,
			UserDisplayName: data.DisplayName,
			TotalEarnings:   data.Earnings,
			PickCount:       pickCountByUser[userID],
			HasCurrentPick:  data.CurrentPick != nil,
		}
		if data.CurrentPick != nil && (isPickWindowClosed || userID == requestingUserID) {
			entry.CurrentPick = data.CurrentPick
		}
		if includePicks {
			entry.Picks = data.Picks
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].TotalEarnings != entries[j].TotalEarnings {
			return entries[i].TotalEarnings > entries[j].TotalEarnings
		}
		return entries[i].UserDisplayName < entries[j].UserDisplayName
	})

	var activeTournamentResponse *ActiveTournament
	if activeTournament != nil {
		now := time.Now().UTC()
		isWindowClosed := activeTournament.PickWindowClosesAt != nil && now.After(*activeTournament.PickWindowClosesAt)
		tournamentID, _ := uuid.FromString(activeTournament.ID)
		activeTournamentResponse = &ActiveTournament{
			ID:                 tournamentID,
			Name:               activeTournament.Name,
			IsPickWindowClosed: isWindowClosed,
			StartDate:          activeTournament.StartDate,
		}
	} else if upcomingTournament != nil {
		activeTournamentResponse = &ActiveTournament{
			ID:                 upcomingTournament.ID,
			Name:               upcomingTournament.Name,
			IsPickWindowClosed: false,
			StartDate:          upcomingTournament.StartDate,
		}
	}

	return &LeagueStandingsResponse{
		Entries:          entries,
		Total:            len(entries),
		SeasonYear:       seasonYear,
		ActiveTournament: activeTournamentResponse,
	}, nil
}
