package leaderboards

import (
	"context"
	"fmt"
	"sort"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/leaderboardentry"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/pick"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

type Service struct {
	db *ent.Client
}

func NewService(db *ent.Client) *Service {
	return &Service{db: db}
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
	leaderboardKeys := make([]pickKey, 0, len(picks))
	for _, p := range picks {
		if p.Edges.Golfer != nil && p.Edges.Tournament != nil {
			leaderboardKeys = append(leaderboardKeys, pickKey{
				TournamentID: p.Edges.Tournament.ID,
				GolferID:     p.Edges.Golfer.ID,
			})
		}
	}

	leaderboardMap := make(map[pickKey]*ent.LeaderboardEntry)
	if len(leaderboardKeys) > 0 {
		tournamentIDs := make([]uuid.UUID, 0)
		golferIDs := make([]uuid.UUID, 0)
		seen := make(map[uuid.UUID]bool)
		for _, k := range leaderboardKeys {
			if !seen[k.TournamentID] {
				tournamentIDs = append(tournamentIDs, k.TournamentID)
				seen[k.TournamentID] = true
			}
		}
		seen = make(map[uuid.UUID]bool)
		for _, k := range leaderboardKeys {
			if !seen[k.GolferID] {
				golferIDs = append(golferIDs, k.GolferID)
				seen[k.GolferID] = true
			}
		}

		entries, err := s.db.LeaderboardEntry.
			Query().
			Where(
				leaderboardentry.HasTournamentWith(tournament.IDIn(tournamentIDs...)),
				leaderboardentry.HasGolferWith(golfer.IDIn(golferIDs...)),
			).
			WithTournament().
			WithGolfer().
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get leaderboard entries: %w", err)
		}

		for _, e := range entries {
			if e.Edges.Tournament != nil && e.Edges.Golfer != nil {
				key := pickKey{TournamentID: e.Edges.Tournament.ID, GolferID: e.Edges.Golfer.ID}
				leaderboardMap[key] = e
			}
		}
	}

	positionCounts := make(map[uuid.UUID]map[int]int)
	for _, e := range leaderboardMap {
		if e.Edges.Tournament == nil {
			continue
		}
		tid := e.Edges.Tournament.ID
		if positionCounts[tid] == nil {
			positionCounts[tid] = make(map[int]int)
		}
		if !e.Cut && e.Position > 0 {
			positionCounts[tid][e.Position]++
		}
	}

	type userPickData struct {
		DisplayName string
		Earnings    int
		Picks       []PickHistory
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
			if lbEntry, ok := leaderboardMap[key]; ok {
				pickEarnings = lbEntry.Earnings
				data.Earnings += pickEarnings

				if lbEntry.Cut {
					posDisplay = "CUT"
				} else if lbEntry.Position > 0 {
					tid := p.Edges.Tournament.ID
					if positionCounts[tid] != nil && positionCounts[tid][lbEntry.Position] > 1 {
						posDisplay = fmt.Sprintf("T%d", lbEntry.Position)
					} else {
						posDisplay = fmt.Sprintf("%d", lbEntry.Position)
					}
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

	return &LeagueStandingsResponse{
		Entries:    entries,
		Total:      len(entries),
		SeasonYear: seasonYear,
	}, nil
}
