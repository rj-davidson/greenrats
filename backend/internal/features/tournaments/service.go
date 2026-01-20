package tournaments

import (
	"context"
	"fmt"
	"sort"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

// Service handles tournament business logic.
type Service struct {
	db *ent.Client
}

// NewService creates a new tournament service.
func NewService(db *ent.Client) *Service {
	return &Service{db: db}
}

// List returns a list of tournaments with optional filtering.
func (s *Service) List(ctx context.Context, req ListTournamentsRequest) (*ListTournamentsResponse, error) {
	query := s.db.Tournament.Query()

	if req.Season > 0 {
		query = query.Where(tournament.SeasonYear(req.Season))
	}

	if req.Status != "" {
		query = query.Where(tournament.StatusEQ(tournament.Status(req.Status)))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count tournaments: %w", err)
	}

	if req.Status == "upcoming" {
		query = query.Order(ent.Asc(tournament.FieldStartDate))
	} else {
		query = query.Order(ent.Desc(tournament.FieldStartDate))
	}

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	results, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tournaments: %w", err)
	}

	tournaments := make([]Tournament, len(results))
	for i, t := range results {
		tournaments[i] = toTournament(t)
	}

	return &ListTournamentsResponse{
		Tournaments: tournaments,
		Total:       total,
	}, nil
}

// GetByID returns a tournament by its ID.
func (s *Service) GetByID(ctx context.Context, id string) (*Tournament, error) {
	uid, err := uuid.FromString(id)
	if err != nil {
		return nil, ErrInvalidTournamentID
	}

	t, err := s.db.Tournament.Get(ctx, uid)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrTournamentNotFound
		}
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	result := toTournament(t)
	return &result, nil
}

// GetActive returns the currently active tournament.
func (s *Service) GetActive(ctx context.Context) (*Tournament, error) {
	t, err := s.db.Tournament.Query().
		Where(tournament.StatusEQ(tournament.StatusActive)).
		Order(ent.Asc(tournament.FieldStartDate)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active tournament: %w", err)
	}

	result := toTournament(t)
	return &result, nil
}

// GetLeaderboard returns the leaderboard entries for a tournament.
func (s *Service) GetLeaderboard(ctx context.Context, id string) (*GetLeaderboardResponse, error) {
	uid, err := uuid.FromString(id)
	if err != nil {
		return nil, ErrInvalidTournamentID
	}

	t, err := s.db.Tournament.Get(ctx, uid)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrTournamentNotFound
		}
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	entries, err := t.QueryEntries().
		WithGolfer().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard entries: %w", err)
	}

	// Sort: players with position first, then cut players, then position 0 last
	// Within each group, sort by position/score then alphabetically by name
	sort.Slice(entries, func(i, j int) bool {
		iCut, jCut := entries[i].Cut, entries[j].Cut
		iPos, jPos := entries[i].Position, entries[j].Position

		// Determine sorting group: 0 = has position, 1 = cut, 2 = no position (0)
		getGroup := func(cut bool, pos int) int {
			if !cut && pos > 0 {
				return 0 // has position
			}
			if cut {
				return 1 // cut
			}
			return 2 // no position
		}

		iGroup, jGroup := getGroup(iCut, iPos), getGroup(jCut, jPos)
		if iGroup != jGroup {
			return iGroup < jGroup
		}

		// Within same group
		if iGroup == 0 {
			// Both have positions: sort by position, then by name for ties
			if iPos != jPos {
				return iPos < jPos
			}
		} else if iGroup == 1 {
			// Both cut: sort by score (ascending), then by name
			if entries[i].Score != entries[j].Score {
				return entries[i].Score < entries[j].Score
			}
		}
		// Sort alphabetically by name
		return entries[i].Edges.Golfer.Name < entries[j].Edges.Golfer.Name
	})

	// Count occurrences of each position to determine ties
	positionCounts := make(map[int]int)
	for _, e := range entries {
		if !e.Cut && e.Position > 0 {
			positionCounts[e.Position]++
		}
	}

	result := make([]LeaderboardEntry, len(entries))
	for i, e := range entries {
		golfer := e.Edges.Golfer

		// Format position display
		posDisplay := "-"
		if e.Cut {
			posDisplay = "CUT"
		} else if e.Position > 0 {
			if positionCounts[e.Position] > 1 {
				posDisplay = fmt.Sprintf("T%d", e.Position)
			} else {
				posDisplay = fmt.Sprintf("%d", e.Position)
			}
		}

		result[i] = LeaderboardEntry{
			Position:        e.Position,
			PositionDisplay: posDisplay,
			GolferID:        golfer.ID.String(),
			GolferName:      golfer.Name,
			CountryCode:     golfer.CountryCode,
			Score:           e.Score,
			TotalStrokes:    e.TotalStrokes,
			Thru:            e.Thru,
			CurrentRound:    e.CurrentRound,
			Cut:             e.Cut,
			Status:          string(e.Status),
			Earnings:        e.Earnings,
		}
	}

	return &GetLeaderboardResponse{
		Entries: result,
		Total:   len(result),
	}, nil
}

func toTournament(t *ent.Tournament) Tournament {
	result := Tournament{
		ID:        t.ID.String(),
		Name:      t.Name,
		StartDate: t.StartDate,
		EndDate:   t.EndDate,
		Status:    string(t.Status),
	}

	if t.Location != nil {
		result.Venue = *t.Location
	}
	if t.Course != nil {
		result.Course = *t.Course
	}
	if t.Purse != nil {
		result.Purse = float64(*t.Purse)
	}

	return result
}
