package tournaments

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
	"github.com/rj-davidson/greenrats/ent/round"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

type Service struct {
	db *ent.Client
}

func NewService(db *ent.Client) *Service {
	return &Service{db: db}
}

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

func (s *Service) GetLeaderboard(ctx context.Context, id string, includeHoles bool, leagueID string) (*GetLeaderboardResponse, error) {
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

	golferPicksMap := s.buildGolferPicksMap(ctx, uid, leagueID, t)

	isCompleted := t.Edges.Champion != nil

	if isCompleted {
		return s.getCompletedLeaderboard(ctx, t, includeHoles, golferPicksMap)
	}

	return s.getLiveLeaderboard(ctx, t, includeHoles, golferPicksMap)
}

func (s *Service) buildGolferPicksMap(ctx context.Context, tournamentID uuid.UUID, leagueID string, t *ent.Tournament) map[uuid.UUID][]string {
	golferPicksMap := make(map[uuid.UUID][]string)
	if leagueID == "" {
		return golferPicksMap
	}

	leagueUUID, err := uuid.FromString(leagueID)
	if err != nil {
		return golferPicksMap
	}

	if t.PickWindowClosesAt == nil || time.Now().UTC().Before(*t.PickWindowClosesAt) {
		return golferPicksMap
	}

	picks, err := s.db.Pick.Query().
		Where(
			pick.HasLeagueWith(league.IDEQ(leagueUUID)),
			pick.HasTournamentWith(tournament.IDEQ(tournamentID)),
		).
		WithUser().
		WithGolfer().
		All(ctx)
	if err != nil {
		return golferPicksMap
	}

	for _, p := range picks {
		if p.Edges.Golfer != nil && p.Edges.User != nil {
			golferID := p.Edges.Golfer.ID
			displayName := p.Edges.User.Email
			if p.Edges.User.DisplayName != nil && *p.Edges.User.DisplayName != "" {
				displayName = *p.Edges.User.DisplayName
			}
			golferPicksMap[golferID] = append(golferPicksMap[golferID], displayName)
		}
	}

	return golferPicksMap
}

func (s *Service) getCompletedLeaderboard(ctx context.Context, t *ent.Tournament, includeHoles bool, golferPicksMap map[uuid.UUID][]string) (*GetLeaderboardResponse, error) {
	placements, err := s.db.Placement.Query().
		Where(placement.HasTournamentWith(tournament.IDEQ(t.ID))).
		WithGolfer().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get placements: %w", err)
	}

	golferIDs := make([]uuid.UUID, 0, len(placements))
	for _, p := range placements {
		if p.Edges.Golfer != nil {
			golferIDs = append(golferIDs, p.Edges.Golfer.ID)
		}
	}

	roundsMap := make(map[uuid.UUID][]*ent.Round)
	if len(golferIDs) > 0 {
		roundQuery := s.db.Round.Query().
			Where(
				round.HasTournamentWith(tournament.IDEQ(t.ID)),
				round.HasGolferWith(golfer.IDIn(golferIDs...)),
			).
			WithGolfer().
			Order(ent.Asc("round_number"))

		if includeHoles {
			roundQuery = roundQuery.WithHoleScores(func(q *ent.HoleScoreQuery) {
				q.Order(ent.Asc("hole_number"))
			})
		}

		rounds, err := roundQuery.All(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get rounds: %w", err)
		}

		for _, r := range rounds {
			if r.Edges.Golfer != nil {
				roundsMap[r.Edges.Golfer.ID] = append(roundsMap[r.Edges.Golfer.ID], r)
			}
		}
	}

	sort.Slice(placements, func(i, j int) bool {
		iStatus, jStatus := placements[i].Status, placements[j].Status
		iPos, jPos := placements[i].PositionNumeric, placements[j].PositionNumeric

		getGroup := func(status placement.Status, posNumeric *int) int {
			if status == placement.StatusFinished && posNumeric != nil && *posNumeric > 0 {
				return 0
			}
			if status == placement.StatusCut {
				return 1
			}
			if status == placement.StatusWithdrawn {
				return 2
			}
			return 3
		}

		iGroup, jGroup := getGroup(iStatus, iPos), getGroup(jStatus, jPos)
		if iGroup != jGroup {
			return iGroup < jGroup
		}

		if iGroup == 0 && iPos != nil && jPos != nil {
			if *iPos != *jPos {
				return *iPos < *jPos
			}
		}

		iGolfer, jGolfer := placements[i].Edges.Golfer, placements[j].Edges.Golfer
		if iGolfer != nil && jGolfer != nil {
			return iGolfer.Name < jGolfer.Name
		}
		return false
	})

	maxRound := 0
	result := make([]LeaderboardEntry, 0, len(placements))
	for _, p := range placements {
		g := p.Edges.Golfer
		if g == nil {
			continue
		}

		position := 0
		if p.PositionNumeric != nil {
			position = *p.PositionNumeric
		}

		score := 0
		if p.ParRelativeScore != nil {
			score = *p.ParRelativeScore
		}

		totalStrokes := 0
		if p.TotalScore != nil {
			totalStrokes = *p.TotalScore
		}

		currentRound := 0
		thru := 0
		rounds := roundsMap[g.ID]
		if len(rounds) > 0 {
			currentRound = len(rounds)
			if currentRound > maxRound {
				maxRound = currentRound
			}
			thru = 18
		}

		status := string(p.Status)

		entry := LeaderboardEntry{
			Position:     position,
			GolferID:     g.ID.String(),
			GolferName:   g.Name,
			CountryCode:  g.CountryCode,
			Score:        score,
			TotalStrokes: totalStrokes,
			Thru:         thru,
			CurrentRound: currentRound,
			Status:       status,
			Earnings:     p.Earnings,
			Rounds:       make([]RoundScore, 0, 4),
		}

		if g.Country != nil {
			entry.Country = *g.Country
		}
		if g.ImageURL != nil {
			entry.ImageURL = *g.ImageURL
		}
		if pickedBy, ok := golferPicksMap[g.ID]; ok {
			entry.PickedBy = pickedBy
		}

		for _, r := range rounds {
			roundScore := RoundScore{
				RoundNumber: r.RoundNumber,
				Score:       r.Score,
			}
			if r.ParRelativeScore != nil {
				roundScore.ParRelativeScore = r.ParRelativeScore
			}
			if r.TeeTime != nil {
				roundScore.TeeTime = r.TeeTime
			}

			if includeHoles && len(r.Edges.HoleScores) > 0 {
				roundScore.Holes = make([]HoleScore, 0, len(r.Edges.HoleScores))
				for _, h := range r.Edges.HoleScores {
					roundScore.Holes = append(roundScore.Holes, HoleScore{
						HoleNumber: h.HoleNumber,
						Par:        h.Par,
						Score:      h.Score,
					})
				}
			}

			entry.Rounds = append(entry.Rounds, roundScore)
		}

		result = append(result, entry)
	}

	return &GetLeaderboardResponse{
		TournamentID:   t.ID.String(),
		TournamentName: t.Name,
		CurrentRound:   maxRound,
		Entries:        result,
		Total:          len(result),
	}, nil
}

type golferRoundData struct {
	golfer       *ent.Golfer
	rounds       []*ent.Round
	totalScore   int
	totalPar     int
	currentRound int
	thru         int
	status       placement.Status
}

func (s *Service) getLiveLeaderboard(ctx context.Context, t *ent.Tournament, includeHoles bool, golferPicksMap map[uuid.UUID][]string) (*GetLeaderboardResponse, error) {
	placements, err := s.db.Placement.Query().
		Where(placement.HasTournamentWith(tournament.IDEQ(t.ID))).
		WithGolfer().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get placements: %w", err)
	}

	statusMap := make(map[uuid.UUID]placement.Status)
	for _, p := range placements {
		if p.Edges.Golfer != nil {
			statusMap[p.Edges.Golfer.ID] = p.Status
		}
	}

	roundQuery := s.db.Round.Query().
		Where(round.HasTournamentWith(tournament.IDEQ(t.ID))).
		WithGolfer().
		Order(ent.Asc("round_number"))

	if includeHoles {
		roundQuery = roundQuery.WithHoleScores(func(q *ent.HoleScoreQuery) {
			q.Order(ent.Asc("hole_number"))
		})
	}

	rounds, err := roundQuery.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rounds: %w", err)
	}

	golferData := make(map[uuid.UUID]*golferRoundData)
	maxRound := 0

	for _, r := range rounds {
		if r.Edges.Golfer == nil {
			continue
		}

		golferID := r.Edges.Golfer.ID
		data, ok := golferData[golferID]
		if !ok {
			data = &golferRoundData{
				golfer: r.Edges.Golfer,
				rounds: make([]*ent.Round, 0, 4),
				status: statusMap[golferID],
			}
			golferData[golferID] = data
		}

		data.rounds = append(data.rounds, r)

		if r.ParRelativeScore != nil {
			data.totalPar += *r.ParRelativeScore
		}
		if r.Score != nil {
			data.totalScore += *r.Score
		}

		if r.RoundNumber > data.currentRound {
			data.currentRound = r.RoundNumber
		}

		if r.RoundNumber > maxRound {
			maxRound = r.RoundNumber
		}

		data.thru = 18
	}

	sortedGolfers := make([]*golferRoundData, 0, len(golferData))
	for _, data := range golferData {
		sortedGolfers = append(sortedGolfers, data)
	}

	getStatusGroup := func(status placement.Status) int {
		switch status {
		case placement.StatusCut:
			return 1
		case placement.StatusWithdrawn:
			return 2
		default:
			return 0
		}
	}

	sort.Slice(sortedGolfers, func(i, j int) bool {
		iGroup := getStatusGroup(sortedGolfers[i].status)
		jGroup := getStatusGroup(sortedGolfers[j].status)
		if iGroup != jGroup {
			return iGroup < jGroup
		}
		if iGroup == 0 {
			if sortedGolfers[i].totalPar != sortedGolfers[j].totalPar {
				return sortedGolfers[i].totalPar < sortedGolfers[j].totalPar
			}
		}
		return sortedGolfers[i].golfer.Name < sortedGolfers[j].golfer.Name
	})

	activeGolfers := make([]*golferRoundData, 0, len(sortedGolfers))
	for _, data := range sortedGolfers {
		if getStatusGroup(data.status) == 0 {
			activeGolfers = append(activeGolfers, data)
		}
	}

	previousPositions := s.calculatePreviousPositions(activeGolfers, maxRound)

	result := make([]LeaderboardEntry, 0, len(sortedGolfers))
	currentPos := 1
	activeIdx := 0

	for _, data := range sortedGolfers {
		statusGroup := getStatusGroup(data.status)
		var position int
		var status string

		switch statusGroup {
		case 0:
			if activeIdx > 0 && data.totalPar != activeGolfers[activeIdx-1].totalPar {
				currentPos = activeIdx + 1
			}
			position = currentPos
			status = "active"
			activeIdx++
		case 1:
			position = 0
			status = "cut"
		default:
			position = 0
			status = "withdrawn"
		}

		entry := LeaderboardEntry{
			Position:     position,
			GolferID:     data.golfer.ID.String(),
			GolferName:   data.golfer.Name,
			CountryCode:  data.golfer.CountryCode,
			Score:        data.totalPar,
			TotalStrokes: data.totalScore,
			Thru:         data.thru,
			CurrentRound: data.currentRound,
			Status:       status,
			Earnings:     0,
			Rounds:       make([]RoundScore, 0, len(data.rounds)),
		}

		if statusGroup == 0 {
			if prevPos, ok := previousPositions[data.golfer.ID.String()]; ok {
				entry.PreviousPosition = &prevPos
				change := prevPos - position
				entry.PositionChange = &change
			}
		}

		if data.golfer.Country != nil {
			entry.Country = *data.golfer.Country
		}
		if data.golfer.ImageURL != nil {
			entry.ImageURL = *data.golfer.ImageURL
		}
		if pickedBy, ok := golferPicksMap[data.golfer.ID]; ok {
			entry.PickedBy = pickedBy
		}

		for _, r := range data.rounds {
			roundScore := RoundScore{
				RoundNumber: r.RoundNumber,
				Score:       r.Score,
			}
			if r.ParRelativeScore != nil {
				roundScore.ParRelativeScore = r.ParRelativeScore
			}
			if r.TeeTime != nil {
				roundScore.TeeTime = r.TeeTime
			}

			if includeHoles && len(r.Edges.HoleScores) > 0 {
				roundScore.Holes = make([]HoleScore, 0, len(r.Edges.HoleScores))
				for _, h := range r.Edges.HoleScores {
					roundScore.Holes = append(roundScore.Holes, HoleScore{
						HoleNumber: h.HoleNumber,
						Par:        h.Par,
						Score:      h.Score,
					})
				}
			}

			entry.Rounds = append(entry.Rounds, roundScore)
		}

		result = append(result, entry)
	}

	return &GetLeaderboardResponse{
		TournamentID:   t.ID.String(),
		TournamentName: t.Name,
		CurrentRound:   maxRound,
		Entries:        result,
		Total:          len(result),
	}, nil
}

func (s *Service) calculatePreviousPositions(golfers []*golferRoundData, currentRound int) map[string]int {
	if currentRound < 2 {
		return nil
	}

	type prevScore struct {
		golferID string
		score    int
	}

	prevScores := make([]prevScore, 0, len(golfers))
	for _, data := range golfers {
		cumulative := 0
		for _, r := range data.rounds {
			if r.RoundNumber < currentRound && r.ParRelativeScore != nil {
				cumulative += *r.ParRelativeScore
			}
		}
		prevScores = append(prevScores, prevScore{
			golferID: data.golfer.ID.String(),
			score:    cumulative,
		})
	}

	sort.Slice(prevScores, func(i, j int) bool {
		return prevScores[i].score < prevScores[j].score
	})

	positions := make(map[string]int)
	currentPos := 1
	for i, ps := range prevScores {
		if i > 0 && ps.score != prevScores[i-1].score {
			currentPos = i + 1
		}
		positions[ps.golferID] = currentPos
	}

	return positions
}

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
		g := e.Edges.Golfer
		if g == nil {
			continue
		}

		entry := FieldEntry{
			GolferID:    g.ID.String(),
			GolferName:  g.Name,
			CountryCode: g.CountryCode,
			EntryStatus: string(e.EntryStatus),
			IsAmateur:   e.IsAmateur,
		}

		if g.Country != nil {
			entry.Country = *g.Country
		}
		if g.Owgr != nil {
			entry.OWGR = g.Owgr
		}
		if e.OwgrAtEntry != nil {
			entry.OWGRAtEntry = e.OwgrAtEntry
		}
		if e.Qualifier != nil {
			entry.Qualifier = *e.Qualifier
		}
		if g.ImageURL != nil {
			entry.ImageURL = *g.ImageURL
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
	status := DeriveStatus(t)

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
