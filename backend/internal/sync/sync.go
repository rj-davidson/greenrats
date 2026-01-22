package sync

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/course"
	"github.com/rj-davidson/greenrats/ent/coursehole"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/golferseason"
	"github.com/rj-davidson/greenrats/ent/holescore"
	"github.com/rj-davidson/greenrats/ent/round"
	"github.com/rj-davidson/greenrats/ent/season"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/googlemaps"
	"github.com/rj-davidson/greenrats/internal/external/pgatour"
	"github.com/rj-davidson/greenrats/internal/timezone"
)

type Service struct {
	db         *ent.Client
	googlemaps *googlemaps.Client
	logger     *slog.Logger
}

func NewService(db *ent.Client, googlemaps *googlemaps.Client, logger *slog.Logger) *Service {
	return &Service{
		db:         db,
		googlemaps: googlemaps,
		logger:     logger,
	}
}

type TournamentUpsertResult struct {
	Created         bool
	BecameCompleted bool
	Tournament      *ent.Tournament
}

func (s *Service) UpsertTournament(ctx context.Context, t *balldontlie.Tournament) (*TournamentUpsertResult, error) {
	startDate, err := time.Parse(balldontlie.DateFormat, t.StartDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start date: %w", err)
	}

	endDate := ParseEndDate(t.EndDate, startDate)

	city := StringPtrValue(t.City)
	state := StringPtrValue(t.State)
	country := StringPtrValue(t.Country)

	tz := s.lookupTimezone(ctx, city, state, country, startDate)
	pickWindow, _ := timezone.CalculatePickWindow(startDate, tz)
	s.logger.Info("tournament timezone resolved",
		"tournament", t.Name,
		"location", fmt.Sprintf("%s, %s, %s", city, state, country),
		"timezone", tz,
		"pick_window_closes", pickWindow.ClosesAt.Format(time.RFC3339))

	var championID *uuid.UUID
	if t.Champion != nil {
		if g, err := s.db.Golfer.Query().Where(golfer.BdlID(t.Champion.ID)).Only(ctx); err == nil {
			championID = &g.ID
		}
	}

	existing, err := s.db.Tournament.Query().
		Where(tournament.BdlID(t.ID)).
		Only(ctx)

	result := &TournamentUpsertResult{}

	switch {
	case ent.IsNotFound(err):
		builder := s.db.Tournament.Create().
			SetBdlID(t.ID).
			SetName(t.Name).
			SetStartDate(startDate).
			SetEndDate(endDate).
			SetSeasonYear(t.Season)

		if pgaTourID := pgatour.GetPGATourID(t.ID); pgaTourID != "" {
			builder.SetPgaTourID(pgaTourID)
		}
		if t.CourseName != nil && *t.CourseName != "" {
			builder.SetCourse(*t.CourseName)
		}
		if t.Purse != nil && *t.Purse != "" {
			if purse, err := strconv.Atoi(*t.Purse); err == nil && purse > 0 {
				builder.SetPurse(purse)
			}
		}
		if city != "" {
			builder.SetCity(city)
		}
		if state != "" {
			builder.SetState(state)
		}
		if country != "" {
			builder.SetCountry(country)
		}
		if tz != "" {
			builder.SetTimezone(tz)
		}
		if pickWindow != nil {
			builder.SetPickWindowOpensAt(pickWindow.OpensAt)
			builder.SetPickWindowClosesAt(pickWindow.ClosesAt)
		}
		if championID != nil {
			builder.SetChampionID(*championID)
		}

		created, err := builder.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create tournament: %w", err)
		}
		result.Created = true
		result.Tournament = created
		s.logger.Debug("created tournament", "name", t.Name)

	case err != nil:
		return nil, fmt.Errorf("failed to query tournament: %w", err)

	default:
		hadChampion := existing.Edges.Champion != nil
		if !hadChampion {
			existingWithChampion, err := s.db.Tournament.Query().
				Where(tournament.IDEQ(existing.ID)).
				WithChampion().
				Only(ctx)
			if err == nil {
				hadChampion = existingWithChampion.Edges.Champion != nil
			}
		}

		updater := existing.Update().
			SetName(t.Name).
			SetStartDate(startDate).
			SetEndDate(endDate)

		if t.CourseName != nil && *t.CourseName != "" {
			updater.SetCourse(*t.CourseName)
		}
		if t.Purse != nil && *t.Purse != "" {
			if purse, err := strconv.Atoi(*t.Purse); err == nil && purse > 0 {
				updater.SetPurse(purse)
			}
		}

		if existing.PgaTourID == nil || *existing.PgaTourID == "" {
			if pgaTourID := pgatour.GetPGATourID(t.ID); pgaTourID != "" {
				updater.SetPgaTourID(pgaTourID)
			}
		}

		if city != "" {
			updater.SetCity(city)
		}
		if state != "" {
			updater.SetState(state)
		}
		if country != "" {
			updater.SetCountry(country)
		}
		if existing.Timezone == nil || *existing.Timezone == "" {
			if tz != "" {
				updater.SetTimezone(tz)
			}
		}
		if existing.PickWindowOpensAt == nil && pickWindow != nil {
			updater.SetPickWindowOpensAt(pickWindow.OpensAt)
			updater.SetPickWindowClosesAt(pickWindow.ClosesAt)
		}
		if championID != nil {
			updater.SetChampionID(*championID)
		}

		updated, err := updater.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to update tournament: %w", err)
		}
		result.Tournament = updated

		if championID != nil && !hadChampion {
			result.BecameCompleted = true
		}

		s.logger.Debug("updated tournament", "name", t.Name)
	}

	return result, nil
}

func (s *Service) UpsertPlayer(ctx context.Context, p *balldontlie.Player) error {
	name := p.DisplayName
	if name == "" && p.FirstName != nil && p.LastName != nil {
		name = fmt.Sprintf("%s %s", *p.FirstName, *p.LastName)
	}

	countryCode := "UNK"
	if p.CountryCode != nil && *p.CountryCode != "" {
		countryCode = *p.CountryCode
	}

	existing, err := s.db.Golfer.Query().
		Where(golfer.BdlID(p.ID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		existing, err = s.db.Golfer.Query().
			Where(golfer.Name(name)).
			Only(ctx)
	}

	switch {
	case ent.IsNotFound(err):
		builder := s.db.Golfer.Create().
			SetBdlID(p.ID).
			SetName(name).
			SetCountryCode(countryCode).
			SetActive(p.Active)

		if p.FirstName != nil {
			builder.SetFirstName(*p.FirstName)
		}
		if p.LastName != nil {
			builder.SetLastName(*p.LastName)
		}
		if p.Country != nil && *p.Country != "" {
			builder.SetCountry(*p.Country)
		}
		if p.OWGR != nil && *p.OWGR > 0 {
			builder.SetOwgr(*p.OWGR)
		}

		_, err = builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create golfer: %w", err)
		}
		s.logger.Debug("created golfer", "name", name)

	case err != nil:
		return fmt.Errorf("failed to query golfer: %w", err)

	default:
		updater := existing.Update().
			SetName(name).
			SetCountryCode(countryCode).
			SetActive(p.Active).
			SetBdlID(p.ID)

		if p.FirstName != nil {
			updater.SetFirstName(*p.FirstName)
		}
		if p.LastName != nil {
			updater.SetLastName(*p.LastName)
		}
		if p.Country != nil && *p.Country != "" {
			updater.SetCountry(*p.Country)
		}
		if p.OWGR != nil && *p.OWGR > 0 {
			updater.SetOwgr(*p.OWGR)
		}

		_, err = updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update golfer: %w", err)
		}
	}

	return nil
}

func (s *Service) UpsertTournamentEntry(ctx context.Context, t *ent.Tournament, r *balldontlie.TournamentResult) error {
	g, err := s.db.Golfer.Query().
		Where(golfer.BdlID(r.Player.ID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to query golfer: %w", err)
	}

	status := tournamententry.StatusActive
	cut := false
	if r.Tournament.Status != nil {
		switch *r.Tournament.Status {
		case "COMPLETED":
			status = tournamententry.StatusFinished
		case "IN_PROGRESS":
			status = tournamententry.StatusActive
		}
	}

	position := 0
	if r.PositionNumeric != nil {
		position = *r.PositionNumeric
	}
	if r.Position != nil && *r.Position == "CUT" {
		cut = true
		status = tournamententry.StatusFinished
	}

	score := 0
	if r.TotalScore != nil {
		score = *r.TotalScore
	}
	totalStrokes := 0

	earnings := 0
	if r.Earnings != nil {
		earnings = *r.Earnings
	}

	existing, err := s.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.ID(t.ID)),
			tournamententry.HasGolferWith(golfer.ID(g.ID)),
		).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		_, err = s.db.TournamentEntry.Create().
			SetTournament(t).
			SetGolfer(g).
			SetPosition(position).
			SetCut(cut).
			SetScore(score).
			SetTotalStrokes(totalStrokes).
			SetStatus(status).
			SetEarnings(earnings).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create tournament entry: %w", err)
		}

	case err != nil:
		return fmt.Errorf("failed to query tournament entry: %w", err)

	default:
		_, err = existing.Update().
			SetPosition(position).
			SetCut(cut).
			SetScore(score).
			SetTotalStrokes(totalStrokes).
			SetStatus(status).
			SetEarnings(earnings).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update tournament entry: %w", err)
		}
	}

	return nil
}

func (s *Service) lookupTimezone(ctx context.Context, city, state, country string, startDate time.Time) string {
	if city == "" && state == "" && country == "" {
		return timezone.DefaultTimezone
	}

	if s.googlemaps == nil {
		s.logger.Warn("Google Maps client not configured, using default timezone")
		return timezone.DefaultTimezone
	}

	tz, err := s.googlemaps.GetTimezone(ctx, city, state, country, startDate)
	if err != nil {
		s.logger.Warn("timezone lookup failed, using default",
			"city", city, "state", state, "country", country, "error", err)
		return timezone.DefaultTimezone
	}

	return tz
}

func StringPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ParseEndDate(endDateStr *string, startDate time.Time) time.Time {
	if endDateStr == nil || *endDateStr == "" {
		return startDate.AddDate(0, 0, 4).Add(6 * time.Hour)
	}

	if endDate, err := time.Parse(balldontlie.DateFormat, *endDateStr); err == nil {
		return endDate.AddDate(0, 0, 1).Add(6 * time.Hour)
	}

	str := strings.TrimSpace(*endDateStr)
	if idx := strings.LastIndex(str, " - "); idx != -1 {
		endPart := strings.TrimSpace(str[idx+3:])

		if day, err := strconv.Atoi(endPart); err == nil && day >= 1 && day <= 31 {
			endDate := time.Date(startDate.Year(), startDate.Month(), day, 0, 0, 0, 0, time.UTC)
			if endDate.Before(startDate) {
				endDate = endDate.AddDate(0, 1, 0)
			}
			return endDate.AddDate(0, 0, 1).Add(6 * time.Hour)
		}

		formats := []string{"Jan 2", "January 2"}
		for _, format := range formats {
			if parsed, err := time.Parse(format, endPart); err == nil {
				endDate := time.Date(startDate.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
				if endDate.Before(startDate) {
					endDate = endDate.AddDate(1, 0, 0)
				}
				return endDate.AddDate(0, 0, 1).Add(6 * time.Hour)
			}
		}
	}

	return startDate.AddDate(0, 0, 4).Add(6 * time.Hour)
}

func (s *Service) UpsertCourse(ctx context.Context, c *balldontlie.Course) (*ent.Course, error) {
	existing, err := s.db.Course.Query().
		Where(course.BdlID(c.ID)).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		builder := s.db.Course.Create().
			SetBdlID(c.ID).
			SetName(c.Name)

		if c.Par != nil {
			builder.SetPar(*c.Par)
		}
		if c.Yardage != nil && *c.Yardage != "" {
			if yardage, err := strconv.Atoi(*c.Yardage); err == nil {
				builder.SetYardage(yardage)
			}
		}
		if c.City != nil {
			builder.SetCity(*c.City)
		}
		if c.State != nil {
			builder.SetState(*c.State)
		}
		if c.Country != nil {
			builder.SetCountry(*c.Country)
		}

		created, err := builder.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create course: %w", err)
		}
		s.logger.Debug("created course", "name", c.Name)
		return created, nil

	case err != nil:
		return nil, fmt.Errorf("failed to query course: %w", err)

	default:
		updater := existing.Update().
			SetName(c.Name)

		if c.Par != nil {
			updater.SetPar(*c.Par)
		}
		if c.Yardage != nil && *c.Yardage != "" {
			if yardage, err := strconv.Atoi(*c.Yardage); err == nil {
				updater.SetYardage(yardage)
			}
		}
		if c.City != nil {
			updater.SetCity(*c.City)
		}
		if c.State != nil {
			updater.SetState(*c.State)
		}
		if c.Country != nil {
			updater.SetCountry(*c.Country)
		}

		updated, err := updater.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to update course: %w", err)
		}
		return updated, nil
	}
}

func (s *Service) UpsertCourseHole(ctx context.Context, courseID uuid.UUID, h *balldontlie.CourseHole) error {
	existing, err := s.db.CourseHole.Query().
		Where(
			coursehole.HasCourseWith(course.IDEQ(courseID)),
			coursehole.HoleNumberEQ(h.HoleNumber),
		).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		builder := s.db.CourseHole.Create().
			SetCourseID(courseID).
			SetHoleNumber(h.HoleNumber).
			SetPar(h.Par)

		if h.Yardage != nil {
			builder.SetYardage(*h.Yardage)
		}

		_, err := builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create course hole: %w", err)
		}

	case err != nil:
		return fmt.Errorf("failed to query course hole: %w", err)

	default:
		updater := existing.Update().
			SetPar(h.Par)

		if h.Yardage != nil {
			updater.SetYardage(*h.Yardage)
		}

		_, err := updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update course hole: %w", err)
		}
	}

	return nil
}

func (s *Service) UpsertRound(ctx context.Context, entryID uuid.UUID, r *balldontlie.PlayerRoundResult) (*ent.Round, error) {
	existing, err := s.db.Round.Query().
		Where(
			round.HasTournamentEntryWith(tournamententry.IDEQ(entryID)),
			round.RoundNumberEQ(r.RoundNumber),
		).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		builder := s.db.Round.Create().
			SetTournamentEntryID(entryID).
			SetRoundNumber(r.RoundNumber)

		if r.Score != nil {
			builder.SetScore(*r.Score)
		}
		if r.ParRelativeScore != nil {
			builder.SetParRelativeScore(*r.ParRelativeScore)
		}

		created, err := builder.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create round: %w", err)
		}
		return created, nil

	case err != nil:
		return nil, fmt.Errorf("failed to query round: %w", err)

	default:
		updater := existing.Update()

		if r.Score != nil {
			updater.SetScore(*r.Score)
		}
		if r.ParRelativeScore != nil {
			updater.SetParRelativeScore(*r.ParRelativeScore)
		}

		updated, err := updater.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to update round: %w", err)
		}
		return updated, nil
	}
}

func (s *Service) UpsertHoleScore(ctx context.Context, roundID uuid.UUID, h *balldontlie.PlayerScorecard) error {
	existing, err := s.db.HoleScore.Query().
		Where(
			holescore.HasRoundWith(round.IDEQ(roundID)),
			holescore.HoleNumberEQ(h.HoleNumber),
		).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		builder := s.db.HoleScore.Create().
			SetRoundID(roundID).
			SetHoleNumber(h.HoleNumber).
			SetPar(h.Par)

		if h.Score != nil {
			builder.SetScore(*h.Score)
		}

		_, err := builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create hole score: %w", err)
		}

	case err != nil:
		return fmt.Errorf("failed to query hole score: %w", err)

	default:
		updater := existing.Update().
			SetPar(h.Par)

		if h.Score != nil {
			updater.SetScore(*h.Score)
		}

		_, err := updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update hole score: %w", err)
		}
	}

	return nil
}

func (s *Service) UpsertGolferSeasonStat(ctx context.Context, golferID, seasonID uuid.UUID, stat *balldontlie.PlayerSeasonStat) error {
	existing, err := s.db.GolferSeason.Query().
		Where(
			golferseason.HasGolferWith(golfer.IDEQ(golferID)),
			golferseason.HasSeasonWith(season.IDEQ(seasonID)),
		).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		builder := s.db.GolferSeason.Create().
			SetGolferID(golferID).
			SetSeasonID(seasonID).
			SetLastSyncedAt(time.Now())

		applyStatToGolferSeason(builder, stat)

		_, err := builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create golfer season: %w", err)
		}

	case err != nil:
		return fmt.Errorf("failed to query golfer season: %w", err)

	default:
		updater := existing.Update().
			SetLastSyncedAt(time.Now())

		applyStatToGolferSeasonUpdate(updater, stat)

		_, err := updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update golfer season: %w", err)
		}
	}

	return nil
}

func applyStatToGolferSeason(builder *ent.GolferSeasonCreate, stat *balldontlie.PlayerSeasonStat) {
	if stat.StatValue == nil {
		return
	}

	switch stat.StatName {
	case "Scoring Average":
		if v, ok := stat.StatValue.(float64); ok {
			builder.SetScoringAvg(v)
		}
	case "Top 10 Finishes":
		if v, ok := stat.StatValue.(float64); ok {
			builder.SetTop10s(int(v))
		}
	case "Cuts Made":
		if v, ok := stat.StatValue.(float64); ok {
			builder.SetCutsMade(int(v))
		}
	case "Events Played":
		if v, ok := stat.StatValue.(float64); ok {
			builder.SetEventsPlayed(int(v))
		}
	case "Wins":
		if v, ok := stat.StatValue.(float64); ok {
			builder.SetWins(int(v))
		}
	case "Official Money":
		if v, ok := stat.StatValue.(float64); ok {
			builder.SetEarnings(int(v))
		}
	case "Driving Distance":
		if v, ok := stat.StatValue.(float64); ok {
			builder.SetDrivingDistance(v)
		}
	case "Driving Accuracy Percentage":
		if v, ok := stat.StatValue.(float64); ok {
			builder.SetDrivingAccuracy(v)
		}
	case "Greens in Regulation Percentage":
		if v, ok := stat.StatValue.(float64); ok {
			builder.SetGirPct(v)
		}
	case "Putting Average":
		if v, ok := stat.StatValue.(float64); ok {
			builder.SetPuttingAvg(v)
		}
	case "Scrambling":
		if v, ok := stat.StatValue.(float64); ok {
			builder.SetScramblingPct(v)
		}
	}
}

func applyStatToGolferSeasonUpdate(updater *ent.GolferSeasonUpdateOne, stat *balldontlie.PlayerSeasonStat) {
	if stat.StatValue == nil {
		return
	}

	switch stat.StatName {
	case "Scoring Average":
		if v, ok := stat.StatValue.(float64); ok {
			updater.SetScoringAvg(v)
		}
	case "Top 10 Finishes":
		if v, ok := stat.StatValue.(float64); ok {
			updater.SetTop10s(int(v))
		}
	case "Cuts Made":
		if v, ok := stat.StatValue.(float64); ok {
			updater.SetCutsMade(int(v))
		}
	case "Events Played":
		if v, ok := stat.StatValue.(float64); ok {
			updater.SetEventsPlayed(int(v))
		}
	case "Wins":
		if v, ok := stat.StatValue.(float64); ok {
			updater.SetWins(int(v))
		}
	case "Official Money":
		if v, ok := stat.StatValue.(float64); ok {
			updater.SetEarnings(int(v))
		}
	case "Driving Distance":
		if v, ok := stat.StatValue.(float64); ok {
			updater.SetDrivingDistance(v)
		}
	case "Driving Accuracy Percentage":
		if v, ok := stat.StatValue.(float64); ok {
			updater.SetDrivingAccuracy(v)
		}
	case "Greens in Regulation Percentage":
		if v, ok := stat.StatValue.(float64); ok {
			updater.SetGirPct(v)
		}
	case "Putting Average":
		if v, ok := stat.StatValue.(float64); ok {
			updater.SetPuttingAvg(v)
		}
	case "Scrambling":
		if v, ok := stat.StatValue.(float64); ok {
			updater.SetScramblingPct(v)
		}
	}
}
