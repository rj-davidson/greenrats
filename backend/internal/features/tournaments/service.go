package tournaments

import (
	"context"
	"fmt"
	"sort"
	"time"

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

	now := time.Now().UTC()
	if req.Status != "" {
		switch DerivedStatus(req.Status) {
		case StatusCompleted:
			query = query.Where(tournament.HasChampion())
		case StatusActive:
			query = query.Where(
				tournament.Not(tournament.HasChampion()),
				tournament.PickWindowClosesAtLT(now),
			)
		case StatusUpcoming:
			query = query.Where(
				tournament.Not(tournament.HasChampion()),
				tournament.Or(
					tournament.PickWindowClosesAtGTE(now),
					tournament.PickWindowClosesAtIsNil(),
				),
			)
		}
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

	results, err := query.WithChampion().All(ctx)
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

	t, err := s.db.Tournament.Query().
		Where(tournament.IDEQ(uid)).
		WithChampion().
		Only(ctx)
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
	now := time.Now().UTC()
	t, err := s.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.PickWindowClosesAtLT(now),
		).
		Order(ent.Asc(tournament.FieldStartDate)).
		WithChampion().
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
func (s *Service) GetLeaderboard(ctx context.Context, id string, includeHoles bool) (*GetLeaderboardResponse, error) {
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

	query := t.QueryLeaderboardEntries().
		WithGolfer().
		WithRounds(func(q *ent.RoundQuery) {
			q.Order(ent.Asc("round_number"))
			if includeHoles {
				q.WithHoleScores(func(hq *ent.HoleScoreQuery) {
					hq.Order(ent.Asc("hole_number"))
				})
			}
		})

	entries, err := query.All(ctx)
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

	// Track max current round for tournament metadata
	maxRound := 0

	result := make([]LeaderboardEntry, len(entries))
	for i, e := range entries {
		golfer := e.Edges.Golfer

		if e.CurrentRound > maxRound {
			maxRound = e.CurrentRound
		}

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

		entry := LeaderboardEntry{
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
			Rounds:          make([]RoundScore, 0, 4),
		}

		if golfer.Country != nil {
			entry.Country = *golfer.Country
		}
		if golfer.ImageURL != nil {
			entry.ImageURL = *golfer.ImageURL
		}

		for _, r := range e.Edges.Rounds {
			round := RoundScore{
				RoundNumber: r.RoundNumber,
				Score:       r.Score,
			}
			if r.ParRelativeScore != nil {
				round.ParRelativeScore = r.ParRelativeScore
			}
			if r.TeeTime != nil {
				round.TeeTime = r.TeeTime
			}

			if includeHoles && len(r.Edges.HoleScores) > 0 {
				round.Holes = make([]HoleScore, 0, len(r.Edges.HoleScores))
				for _, h := range r.Edges.HoleScores {
					round.Holes = append(round.Holes, HoleScore{
						HoleNumber: h.HoleNumber,
						Par:        h.Par,
						Score:      h.Score,
					})
				}
			}

			entry.Rounds = append(entry.Rounds, round)
		}

		result[i] = entry
	}

	return &GetLeaderboardResponse{
		TournamentID:   t.ID.String(),
		TournamentName: t.Name,
		CurrentRound:   maxRound,
		Entries:        result,
		Total:          len(result),
	}, nil
}

// GetField returns the field entries for a tournament.
func (s *Service) GetField(ctx context.Context, id string) (*GetFieldResponse, error) {
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

	entries, err := t.QueryFieldEntries().
		WithGolfer().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get field entries: %w", err)
	}

	result := make([]FieldEntry, 0, len(entries))
	for _, e := range entries {
		golfer := e.Edges.Golfer
		if golfer == nil {
			continue
		}

		entry := FieldEntry{
			GolferID:    golfer.ID.String(),
			GolferName:  golfer.Name,
			CountryCode: golfer.CountryCode,
			EntryStatus: string(e.EntryStatus),
			IsAmateur:   e.IsAmateur,
		}

		if golfer.Country != nil {
			entry.Country = *golfer.Country
		}
		if golfer.Owgr != nil {
			entry.OWGR = golfer.Owgr
		}
		if e.OwgrAtEntry != nil {
			entry.OWGRAtEntry = e.OwgrAtEntry
		}
		if e.Qualifier != nil {
			entry.Qualifier = *e.Qualifier
		}
		if golfer.ImageURL != nil {
			entry.ImageURL = *golfer.ImageURL
		}

		result = append(result, entry)
	}

	return &GetFieldResponse{
		TournamentID:   t.ID.String(),
		TournamentName: t.Name,
		Entries:        result,
		Total:          len(result),
	}, nil
}

func toTournament(t *ent.Tournament) Tournament {
	hasChampion := t.Edges.Champion != nil
	status := DeriveStatus(t, hasChampion)

	result := Tournament{
		ID:                 t.ID.String(),
		Name:               t.Name,
		StartDate:          t.StartDate,
		EndDate:            t.EndDate,
		Status:             string(status),
		PickWindowOpensAt:  t.PickWindowOpensAt,
		PickWindowClosesAt: t.PickWindowClosesAt,
	}

	if t.Course != nil {
		result.Course = *t.Course
	}
	if t.Purse != nil {
		result.Purse = float64(*t.Purse)
	}
	if t.City != nil {
		result.City = *t.City
	}
	if t.State != nil {
		result.State = *t.State
	}
	if t.Country != nil {
		result.Country = *t.Country
	}
	if t.Timezone != nil {
		result.Timezone = *t.Timezone
	}
	if hasChampion {
		result.ChampionID = t.Edges.Champion.ID.String()
		result.ChampionName = t.Edges.Champion.Name
	}

	return result
}
