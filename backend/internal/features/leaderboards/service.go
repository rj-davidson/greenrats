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
	resp, err := s.GetLeagueStandings(ctx, leagueID, seasonYear, false)
	if err != nil {
		return nil, err
	}

	entries := make([]LeaderboardEntry, len(resp.Entries))
	for i, e := range resp.Entries {
		entries[i] = LeaderboardEntry{
			Rank:        e.Rank,
			RankDisplay: e.RankDisplay,
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

func (s *Service) GetLeagueStandings(ctx context.Context, leagueID uuid.UUID, seasonYear int, includePicks bool) (*LeagueStandingsResponse, error) {
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

	positionCounts := make(map[uuid.UUID]map[int]int)
	for _, pl := range placementMap {
		if pl.Edges.Tournament == nil {
			continue
		}
		tid := pl.Edges.Tournament.ID
		if positionCounts[tid] == nil {
			positionCounts[tid] = make(map[int]int)
		}
		if pl.Status != placement.StatusCut && pl.PositionNumeric != nil && *pl.PositionNumeric > 0 {
			positionCounts[tid][*pl.PositionNumeric]++
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
		var posDisplay string

		if p.Edges.Golfer != nil && p.Edges.Tournament != nil {
			key := pickKey{TournamentID: p.Edges.Tournament.ID, GolferID: p.Edges.Golfer.ID}
			if pl, ok := placementMap[key]; ok {
				pickEarnings = pl.Earnings
				data.Earnings += pickEarnings

				if pl.Status == placement.StatusCut {
					posDisplay = "CUT"
				} else if pl.Status == placement.StatusWithdrawn {
					posDisplay = "WD"
				} else if pl.PositionNumeric != nil && *pl.PositionNumeric > 0 {
					tid := p.Edges.Tournament.ID
					if positionCounts[tid] != nil && positionCounts[tid][*pl.PositionNumeric] > 1 {
						posDisplay = fmt.Sprintf("T%d", *pl.PositionNumeric)
					} else {
						posDisplay = fmt.Sprintf("%d", *pl.PositionNumeric)
					}
				}
			}

			if activeTournament != nil && p.Edges.Tournament.ID.String() == activeTournament.ID {
				data.CurrentPick = &CurrentPick{
					TournamentID:   p.Edges.Tournament.ID,
					TournamentName: p.Edges.Tournament.Name,
					GolferID:       p.Edges.Golfer.ID,
					GolferName:     p.Edges.Golfer.Name,
				}
			}

			if includePicks {
				data.Picks = append(data.Picks, PickHistory{
					TournamentID:    p.Edges.Tournament.ID,
					TournamentName:  p.Edges.Tournament.Name,
					GolferID:        p.Edges.Golfer.ID,
					GolferName:      p.Edges.Golfer.Name,
					PositionDisplay: posDisplay,
					Earnings:        pickEarnings,
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
			CurrentPick:     data.CurrentPick,
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

	currentRank := 1
	for i := range entries {
		if i > 0 && entries[i].TotalEarnings < entries[i-1].TotalEarnings {
			currentRank = i + 1
		}
		entries[i].Rank = currentRank
	}

	rankCounts := make(map[int]int)
	for _, e := range entries {
		rankCounts[e.Rank]++
	}
	for i := range entries {
		if rankCounts[entries[i].Rank] > 1 {
			entries[i].RankDisplay = fmt.Sprintf("T%d", entries[i].Rank)
		} else {
			entries[i].RankDisplay = fmt.Sprintf("%d", entries[i].Rank)
		}
	}

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
	}

	return &LeagueStandingsResponse{
		Entries:          entries,
		Total:            len(entries),
		SeasonYear:       seasonYear,
		ActiveTournament: activeTournamentResponse,
	}, nil
}
