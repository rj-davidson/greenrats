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

	userEarnings := make(map[uuid.UUID]*LeaderboardEntry)

	for _, p := range picks {
		if p.Edges.User == nil {
			continue
		}

		userID := p.Edges.User.ID
		entry, exists := userEarnings[userID]
		if !exists {
			displayName := "Unknown"
			if p.Edges.User.DisplayName != nil {
				displayName = *p.Edges.User.DisplayName
			}
			entry = &LeaderboardEntry{
				UserID:      userID,
				DisplayName: displayName,
				Earnings:    0,
				PickCount:   0,
			}
			userEarnings[userID] = entry
		}

		entry.PickCount++

		if p.Edges.Golfer != nil && p.Edges.Tournament != nil {
			lbEntry, err := s.db.LeaderboardEntry.
				Query().
				Where(
					leaderboardentry.HasTournamentWith(tournament.IDEQ(p.Edges.Tournament.ID)),
					leaderboardentry.HasGolferWith(golfer.IDEQ(p.Edges.Golfer.ID)),
				).
				Only(ctx)
			if err == nil {
				entry.Earnings += lbEntry.Earnings
			}
		}
	}

	entries := make([]LeaderboardEntry, 0, len(userEarnings))
	for _, e := range userEarnings {
		entries = append(entries, *e)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Earnings != entries[j].Earnings {
			return entries[i].Earnings > entries[j].Earnings
		}
		return entries[i].DisplayName < entries[j].DisplayName
	})

	currentRank := 1
	for i := range entries {
		if i > 0 && entries[i].Earnings < entries[i-1].Earnings {
			currentRank = i + 1
		}
		entries[i].Rank = currentRank
	}

	return &LeagueLeaderboardResponse{
		Entries:    entries,
		Total:      len(entries),
		SeasonYear: seasonYear,
	}, nil
}
