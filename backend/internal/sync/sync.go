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
	"github.com/rj-davidson/greenrats/ent/fieldentry"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/golferseason"
	"github.com/rj-davidson/greenrats/ent/holescore"
	"github.com/rj-davidson/greenrats/ent/placement"
	"github.com/rj-davidson/greenrats/ent/round"
	"github.com/rj-davidson/greenrats/ent/season"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamentcourse"
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

func NewService(db *ent.Client, gmaps *googlemaps.Client, logger *slog.Logger) *Service {
	return &Service{
		db:         db,
		googlemaps: gmaps,
		logger:     logger,
	}
}

func (s *Service) UpsertSeason(ctx context.Context, year int) (*ent.Season, error) {
	existing, err := s.db.Season.Query().
		Where(season.YearEQ(year)).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		startDate := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(year, time.December, 31, 23, 59, 59, 0, time.UTC)

		created, err := s.db.Season.Create().
			SetYear(year).
			SetStartDate(startDate).
			SetEndDate(endDate).
			SetIsCurrent(true).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create season: %w", err)
		}
		s.logger.Info("created season", "year", year)
		return created, nil

	case err != nil:
		return nil, fmt.Errorf("failed to query season: %w", err)

	default:
		return existing, nil
	}
}

type TournamentUpsertResult struct {
	Created         bool
	BecameCompleted bool
	Tournament      *ent.Tournament
}

func (s *Service) UpsertTournament(ctx context.Context, t *balldontlie.Tournament, seasonEnt *ent.Season) (*TournamentUpsertResult, error) {
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
			SetSeasonYear(t.Season).
			SetSeason(seasonEnt)

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
			SetEndDate(endDate).
			SetSeason(seasonEnt)

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

	if len(t.Courses) > 0 {
		for _, tc := range t.Courses {
			courseEnt, err := s.UpsertCourseFromRef(ctx, &tc.Course)
			if err != nil {
				s.logger.Warn("failed to upsert course from tournament",
					"tournament", t.Name,
					"course", tc.Course.Name,
					"error", err)
				continue
			}

			if err := s.UpsertTournamentCourse(ctx, result.Tournament.ID, courseEnt.ID, tc.Rounds); err != nil {
				s.logger.Warn("failed to upsert tournament course",
					"tournament", t.Name,
					"course", tc.Course.Name,
					"error", err)
			}
		}

		if result.Tournament.Course == nil || *result.Tournament.Course == "" {
			courseName := t.Courses[0].Course.Name
			_, err := result.Tournament.Update().SetCourse(courseName).Save(ctx)
			if err != nil {
				s.logger.Warn("failed to set primary course name", "error", err)
			}
		}
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

		applyPlayerFieldsToCreate(builder, p)

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

		applyPlayerFieldsToUpdate(updater, p)

		_, err = updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update golfer: %w", err)
		}
	}

	return nil
}

func applyPlayerFieldsToCreate(builder *ent.GolferCreate, p *balldontlie.Player) {
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
	if p.Height != nil && *p.Height != "" {
		builder.SetHeight(*p.Height)
	}
	if p.Weight != nil && *p.Weight != "" {
		builder.SetWeight(*p.Weight)
	}
	if p.BirthDate != nil && *p.BirthDate != "" {
		if bd, err := time.Parse("2006-01-02", *p.BirthDate); err == nil {
			builder.SetBirthDate(bd)
		}
	}
	if p.BirthplaceCity != nil && *p.BirthplaceCity != "" {
		builder.SetBirthplaceCity(*p.BirthplaceCity)
	}
	if p.BirthplaceState != nil && *p.BirthplaceState != "" {
		builder.SetBirthplaceState(*p.BirthplaceState)
	}
	if p.BirthplaceCountry != nil && *p.BirthplaceCountry != "" {
		builder.SetBirthplaceCountry(*p.BirthplaceCountry)
	}
	if p.TurnedPro != nil && *p.TurnedPro != "" {
		if year, err := strconv.Atoi(*p.TurnedPro); err == nil {
			builder.SetTurnedPro(year)
		}
	}
	if p.School != nil && *p.School != "" {
		builder.SetSchool(*p.School)
	}
	if p.ResidenceCity != nil && *p.ResidenceCity != "" {
		builder.SetResidenceCity(*p.ResidenceCity)
	}
	if p.ResidenceState != nil && *p.ResidenceState != "" {
		builder.SetResidenceState(*p.ResidenceState)
	}
	if p.ResidenceCountry != nil && *p.ResidenceCountry != "" {
		builder.SetResidenceCountry(*p.ResidenceCountry)
	}
}

func applyPlayerFieldsToUpdate(updater *ent.GolferUpdateOne, p *balldontlie.Player) {
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
	if p.Height != nil && *p.Height != "" {
		updater.SetHeight(*p.Height)
	}
	if p.Weight != nil && *p.Weight != "" {
		updater.SetWeight(*p.Weight)
	}
	if p.BirthDate != nil && *p.BirthDate != "" {
		if bd, err := time.Parse("2006-01-02", *p.BirthDate); err == nil {
			updater.SetBirthDate(bd)
		}
	}
	if p.BirthplaceCity != nil && *p.BirthplaceCity != "" {
		updater.SetBirthplaceCity(*p.BirthplaceCity)
	}
	if p.BirthplaceState != nil && *p.BirthplaceState != "" {
		updater.SetBirthplaceState(*p.BirthplaceState)
	}
	if p.BirthplaceCountry != nil && *p.BirthplaceCountry != "" {
		updater.SetBirthplaceCountry(*p.BirthplaceCountry)
	}
	if p.TurnedPro != nil && *p.TurnedPro != "" {
		if year, err := strconv.Atoi(*p.TurnedPro); err == nil {
			updater.SetTurnedPro(year)
		}
	}
	if p.School != nil && *p.School != "" {
		updater.SetSchool(*p.School)
	}
	if p.ResidenceCity != nil && *p.ResidenceCity != "" {
		updater.SetResidenceCity(*p.ResidenceCity)
	}
	if p.ResidenceState != nil && *p.ResidenceState != "" {
		updater.SetResidenceState(*p.ResidenceState)
	}
	if p.ResidenceCountry != nil && *p.ResidenceCountry != "" {
		updater.SetResidenceCountry(*p.ResidenceCountry)
	}
}

func (s *Service) UpsertFieldEntry(ctx context.Context, t *ent.Tournament, f *balldontlie.TournamentField) error {
	g, err := s.db.Golfer.Query().
		Where(golfer.BdlID(f.Player.ID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to query golfer: %w", err)
	}

	entryStatus := mapFieldEntryStatus(f.EntryStatus)

	existing, err := s.db.FieldEntry.Query().
		Where(
			fieldentry.HasTournamentWith(tournament.ID(t.ID)),
			fieldentry.HasGolferWith(golfer.ID(g.ID)),
		).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		builder := s.db.FieldEntry.Create().
			SetTournament(t).
			SetGolfer(g).
			SetEntryStatus(entryStatus).
			SetIsAmateur(f.IsAmateur)

		if f.Qualifier != "" {
			builder.SetQualifier(f.Qualifier)
		}
		if f.OWGR != nil {
			builder.SetOwgrAtEntry(*f.OWGR)
		}

		_, err = builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create field entry: %w", err)
		}

	case err != nil:
		return fmt.Errorf("failed to query field entry: %w", err)

	default:
		updater := existing.Update().
			SetEntryStatus(entryStatus).
			SetIsAmateur(f.IsAmateur)

		if f.Qualifier != "" {
			updater.SetQualifier(f.Qualifier)
		}
		if f.OWGR != nil {
			updater.SetOwgrAtEntry(*f.OWGR)
		}

		_, err = updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update field entry: %w", err)
		}
	}

	return nil
}

func mapFieldEntryStatus(status string) fieldentry.EntryStatus {
	switch strings.ToLower(status) {
	case "committed", "confirmed":
		return fieldentry.EntryStatusConfirmed
	case "alternate":
		return fieldentry.EntryStatusAlternate
	case "withdrawn", "wd":
		return fieldentry.EntryStatusWithdrawn
	default:
		return fieldentry.EntryStatusPending
	}
}

func (s *Service) UpsertPlacement(ctx context.Context, t *ent.Tournament, r *balldontlie.TournamentResult) error {
	g, err := s.db.Golfer.Query().
		Where(golfer.BdlID(r.Player.ID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to query golfer: %w", err)
	}

	status := placement.StatusFinished
	position := ""
	var positionNumeric *int

	if r.Position != nil {
		position = *r.Position
		switch position {
		case "CUT":
			status = placement.StatusCut
		case "WD":
			status = placement.StatusWithdrawn
		}
	}

	if r.PositionNumeric != nil && *r.PositionNumeric > 0 {
		positionNumeric = r.PositionNumeric
	}

	var totalScore *int
	if r.TotalScore != nil {
		totalScore = r.TotalScore
	}

	var parRelativeScore *int
	if r.ParRelativeScore != nil {
		parRelativeScore = r.ParRelativeScore
	}

	earnings := 0
	if r.Earnings != nil {
		earnings = int(*r.Earnings)
	}

	existing, err := s.db.Placement.Query().
		Where(
			placement.HasTournamentWith(tournament.ID(t.ID)),
			placement.HasGolferWith(golfer.ID(g.ID)),
		).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		builder := s.db.Placement.Create().
			SetTournament(t).
			SetGolfer(g).
			SetPosition(position).
			SetStatus(status).
			SetEarnings(earnings)

		if positionNumeric != nil {
			builder.SetPositionNumeric(*positionNumeric)
		}
		if totalScore != nil {
			builder.SetTotalScore(*totalScore)
		}
		if parRelativeScore != nil {
			builder.SetParRelativeScore(*parRelativeScore)
		}

		_, err = builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create placement: %w", err)
		}

	case err != nil:
		return fmt.Errorf("failed to query placement: %w", err)

	default:
		updater := existing.Update().
			SetPosition(position).
			SetStatus(status).
			SetEarnings(earnings)

		if positionNumeric != nil {
			updater.SetPositionNumeric(*positionNumeric)
		} else {
			updater.ClearPositionNumeric()
		}
		if totalScore != nil {
			updater.SetTotalScore(*totalScore)
		}
		if parRelativeScore != nil {
			updater.SetParRelativeScore(*parRelativeScore)
		}

		_, err = updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update placement: %w", err)
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

		applyCourseFieldsToCreate(builder, c)

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

		applyCourseFieldsToUpdate(updater, c)

		updated, err := updater.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to update course: %w", err)
		}
		return updated, nil
	}
}

func applyCourseFieldsToCreate(builder *ent.CourseCreate, c *balldontlie.Course) {
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
	if c.Established != nil && *c.Established != "" {
		if year, err := strconv.Atoi(*c.Established); err == nil {
			builder.SetEstablished(year)
		}
	}
	if c.Architect != nil && *c.Architect != "" {
		builder.SetArchitect(*c.Architect)
	}
	if c.FairwayGrass != nil && *c.FairwayGrass != "" {
		builder.SetFairwayGrass(*c.FairwayGrass)
	}
	if c.RoughGrass != nil && *c.RoughGrass != "" {
		builder.SetRoughGrass(*c.RoughGrass)
	}
	if c.GreenGrass != nil && *c.GreenGrass != "" {
		builder.SetGreenGrass(*c.GreenGrass)
	}
}

func applyCourseFieldsToUpdate(updater *ent.CourseUpdateOne, c *balldontlie.Course) {
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
	if c.Established != nil && *c.Established != "" {
		if year, err := strconv.Atoi(*c.Established); err == nil {
			updater.SetEstablished(year)
		}
	}
	if c.Architect != nil && *c.Architect != "" {
		updater.SetArchitect(*c.Architect)
	}
	if c.FairwayGrass != nil && *c.FairwayGrass != "" {
		updater.SetFairwayGrass(*c.FairwayGrass)
	}
	if c.RoughGrass != nil && *c.RoughGrass != "" {
		updater.SetRoughGrass(*c.RoughGrass)
	}
	if c.GreenGrass != nil && *c.GreenGrass != "" {
		updater.SetGreenGrass(*c.GreenGrass)
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

func (s *Service) UpsertCourseFromRef(ctx context.Context, ref *balldontlie.CourseRef) (*ent.Course, error) {
	existing, err := s.db.Course.Query().
		Where(course.BdlID(ref.ID)).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		builder := s.db.Course.Create().
			SetBdlID(ref.ID).
			SetName(ref.Name)

		if ref.Par != nil {
			builder.SetPar(*ref.Par)
		}
		if ref.City != nil {
			builder.SetCity(*ref.City)
		}
		if ref.State != nil {
			builder.SetState(*ref.State)
		}
		if ref.Country != nil {
			builder.SetCountry(*ref.Country)
		}
		if ref.Yardage != nil && *ref.Yardage != "" {
			if yardage, err := strconv.Atoi(*ref.Yardage); err == nil {
				builder.SetYardage(yardage)
			}
		}

		created, err := builder.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create course: %w", err)
		}
		s.logger.Debug("created course from ref", "name", ref.Name)
		return created, nil

	case err != nil:
		return nil, fmt.Errorf("failed to query course: %w", err)

	default:
		return existing, nil
	}
}

func (s *Service) SetRoundCourse(ctx context.Context, roundID, courseID uuid.UUID) error {
	return s.db.Round.UpdateOneID(roundID).
		SetCourseID(courseID).
		Exec(ctx)
}

func (s *Service) UpsertTournamentCourse(ctx context.Context, tournamentID, courseID uuid.UUID, rounds []int) error {
	existing, err := s.db.TournamentCourse.Query().
		Where(
			tournamentcourse.HasTournamentWith(tournament.IDEQ(tournamentID)),
			tournamentcourse.HasCourseWith(course.IDEQ(courseID)),
		).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		_, err := s.db.TournamentCourse.Create().
			SetTournamentID(tournamentID).
			SetCourseID(courseID).
			SetRounds(rounds).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create tournament course: %w", err)
		}

	case err != nil:
		return fmt.Errorf("failed to query tournament course: %w", err)

	default:
		_, err := existing.Update().
			SetRounds(rounds).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update tournament course: %w", err)
		}
	}

	return nil
}

func (s *Service) getCourseForRound(ctx context.Context, tournamentID uuid.UUID, roundNumber int) *uuid.UUID {
	tcs, err := s.db.TournamentCourse.Query().
		Where(tournamentcourse.HasTournamentWith(tournament.IDEQ(tournamentID))).
		WithCourse().
		All(ctx)
	if err != nil || len(tcs) == 0 {
		return nil
	}

	if len(tcs) == 1 {
		return &tcs[0].Edges.Course.ID
	}

	for _, tc := range tcs {
		for _, r := range tc.Rounds {
			if r == roundNumber {
				return &tc.Edges.Course.ID
			}
		}
	}

	return nil
}

func (s *Service) UpsertRound(ctx context.Context, tournamentID, golferID uuid.UUID, r *balldontlie.PlayerRoundResult) (*ent.Round, error) {
	existing, err := s.db.Round.Query().
		Where(
			round.HasTournamentWith(tournament.IDEQ(tournamentID)),
			round.HasGolferWith(golfer.IDEQ(golferID)),
			round.RoundNumberEQ(r.RoundNumber),
		).
		Only(ctx)

	courseID := s.getCourseForRound(ctx, tournamentID, r.RoundNumber)

	switch {
	case ent.IsNotFound(err):
		builder := s.db.Round.Create().
			SetTournamentID(tournamentID).
			SetGolferID(golferID).
			SetRoundNumber(r.RoundNumber)

		if courseID != nil {
			builder.SetCourseID(*courseID)
		}

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

		if courseID != nil {
			updater.SetCourseID(*courseID)
		}
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

	if err := s.updateRoundThru(ctx, roundID); err != nil {
		s.logger.Warn("failed to update round thru", "round_id", roundID, "error", err)
	}

	if err := s.updateRoundScores(ctx, roundID); err != nil {
		s.logger.Warn("failed to update round scores", "round_id", roundID, "error", err)
	}

	return nil
}

func (s *Service) updateRoundThru(ctx context.Context, roundID uuid.UUID) error {
	holeScores, err := s.db.HoleScore.Query().
		Where(holescore.HasRoundWith(round.IDEQ(roundID))).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query hole scores: %w", err)
	}

	thru := 0
	for _, h := range holeScores {
		if h.Score != nil {
			thru++
		}
	}

	_, err = s.db.Round.UpdateOneID(roundID).SetThru(thru).Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update round thru: %w", err)
	}

	return nil
}

func (s *Service) updateRoundScores(ctx context.Context, roundID uuid.UUID) error {
	holeScores, err := s.db.HoleScore.Query().
		Where(holescore.HasRoundWith(round.IDEQ(roundID))).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query hole scores: %w", err)
	}

	var totalScore int
	var parRelativeScore int
	scoredHoles := 0

	for _, h := range holeScores {
		if h.Score != nil {
			totalScore += *h.Score
			parRelativeScore += *h.Score - h.Par
			scoredHoles++
		}
	}

	if scoredHoles == 0 {
		return nil
	}

	_, err = s.db.Round.UpdateOneID(roundID).
		SetScore(totalScore).
		SetParRelativeScore(parRelativeScore).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update round scores: %w", err)
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

func getStatValue(items []balldontlie.StatValueItem) string {
	if len(items) == 0 {
		return ""
	}
	return items[0].StatValue
}

func applyStatToGolferSeason(builder *ent.GolferSeasonCreate, stat *balldontlie.PlayerSeasonStat) {
	if len(stat.StatValue) == 0 {
		return
	}

	valStr := getStatValue(stat.StatValue)
	if valStr == "" {
		return
	}

	switch stat.StatName {
	case "Scoring Average":
		if v, err := strconv.ParseFloat(valStr, 64); err == nil {
			builder.SetScoringAvg(v)
		}
	case "Top 10 Finishes":
		if v, err := strconv.Atoi(valStr); err == nil {
			builder.SetTop10s(v)
		}
	case "Cuts Made":
		if v, err := strconv.Atoi(valStr); err == nil {
			builder.SetCutsMade(v)
		}
	case "Events Played":
		if v, err := strconv.Atoi(valStr); err == nil {
			builder.SetEventsPlayed(v)
		}
	case "Wins":
		if v, err := strconv.Atoi(valStr); err == nil {
			builder.SetWins(v)
		}
	case "Official Money":
		cleaned := strings.ReplaceAll(strings.ReplaceAll(valStr, "$", ""), ",", "")
		if v, err := strconv.ParseFloat(cleaned, 64); err == nil {
			builder.SetEarnings(int(v))
		}
	case "Driving Distance":
		if v, err := strconv.ParseFloat(valStr, 64); err == nil {
			builder.SetDrivingDistance(v)
		}
	case "Driving Accuracy Percentage":
		if v, err := strconv.ParseFloat(valStr, 64); err == nil {
			builder.SetDrivingAccuracy(v)
		}
	case "Greens in Regulation Percentage":
		if v, err := strconv.ParseFloat(valStr, 64); err == nil {
			builder.SetGirPct(v)
		}
	case "Putting Average":
		if v, err := strconv.ParseFloat(valStr, 64); err == nil {
			builder.SetPuttingAvg(v)
		}
	case "Scrambling":
		if v, err := strconv.ParseFloat(valStr, 64); err == nil {
			builder.SetScramblingPct(v)
		}
	}
}

func applyStatToGolferSeasonUpdate(updater *ent.GolferSeasonUpdateOne, stat *balldontlie.PlayerSeasonStat) {
	if len(stat.StatValue) == 0 {
		return
	}

	valStr := getStatValue(stat.StatValue)
	if valStr == "" {
		return
	}

	switch stat.StatName {
	case "Scoring Average":
		if v, err := strconv.ParseFloat(valStr, 64); err == nil {
			updater.SetScoringAvg(v)
		}
	case "Top 10 Finishes":
		if v, err := strconv.Atoi(valStr); err == nil {
			updater.SetTop10s(v)
		}
	case "Cuts Made":
		if v, err := strconv.Atoi(valStr); err == nil {
			updater.SetCutsMade(v)
		}
	case "Events Played":
		if v, err := strconv.Atoi(valStr); err == nil {
			updater.SetEventsPlayed(v)
		}
	case "Wins":
		if v, err := strconv.Atoi(valStr); err == nil {
			updater.SetWins(v)
		}
	case "Official Money":
		cleaned := strings.ReplaceAll(strings.ReplaceAll(valStr, "$", ""), ",", "")
		if v, err := strconv.ParseFloat(cleaned, 64); err == nil {
			updater.SetEarnings(int(v))
		}
	case "Driving Distance":
		if v, err := strconv.ParseFloat(valStr, 64); err == nil {
			updater.SetDrivingDistance(v)
		}
	case "Driving Accuracy Percentage":
		if v, err := strconv.ParseFloat(valStr, 64); err == nil {
			updater.SetDrivingAccuracy(v)
		}
	case "Greens in Regulation Percentage":
		if v, err := strconv.ParseFloat(valStr, 64); err == nil {
			updater.SetGirPct(v)
		}
	case "Putting Average":
		if v, err := strconv.ParseFloat(valStr, 64); err == nil {
			updater.SetPuttingAvg(v)
		}
	case "Scrambling":
		if v, err := strconv.ParseFloat(valStr, 64); err == nil {
			updater.SetScramblingPct(v)
		}
	}
}
