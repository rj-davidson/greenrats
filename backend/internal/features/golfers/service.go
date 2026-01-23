package golfers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
)

type Service struct {
	db *ent.Client
}

func NewService(db *ent.Client) *Service {
	return &Service{db: db}
}

func (s *Service) GetByID(ctx context.Context, id string) (*GolferDetail, error) {
	uid, err := uuid.FromString(id)
	if err != nil {
		return nil, ErrInvalidGolferID
	}

	g, err := s.db.Golfer.Query().
		Where(golfer.IDEQ(uid)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrGolferNotFound
		}
		return nil, fmt.Errorf("failed to get golfer: %w", err)
	}

	detail := &GolferDetail{
		ID:          g.ID.String(),
		Name:        g.Name,
		CountryCode: g.CountryCode,
		Active:      g.Active,
	}

	if g.FirstName != nil {
		detail.FirstName = *g.FirstName
	}
	if g.LastName != nil {
		detail.LastName = *g.LastName
	}
	if g.Country != nil {
		detail.Country = *g.Country
	}
	if g.Owgr != nil {
		detail.OWGR = g.Owgr
	}
	if g.ImageURL != nil {
		detail.ImageURL = *g.ImageURL
	}
	if g.Height != nil {
		detail.Height = *g.Height
	}
	if g.Weight != nil {
		detail.Weight = *g.Weight
	}
	if g.BirthDate != nil {
		detail.BirthDate = g.BirthDate
	}
	if g.BirthplaceCity != nil {
		detail.BirthplaceCity = *g.BirthplaceCity
	}
	if g.BirthplaceState != nil {
		detail.BirthplaceState = *g.BirthplaceState
	}
	if g.BirthplaceCountry != nil {
		detail.BirthplaceCountry = *g.BirthplaceCountry
	}
	if g.TurnedPro != nil {
		detail.TurnedPro = g.TurnedPro
	}
	if g.School != nil {
		detail.School = *g.School
	}
	if g.ResidenceCity != nil {
		detail.ResidenceCity = *g.ResidenceCity
	}
	if g.ResidenceState != nil {
		detail.ResidenceState = *g.ResidenceState
	}
	if g.ResidenceCountry != nil {
		detail.ResidenceCountry = *g.ResidenceCountry
	}

	return detail, nil
}

func NormalizeName(name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}

	candidates := []string{name}

	if strings.Contains(name, ",") {
		parts := strings.SplitN(name, ",", 2)
		last := strings.TrimSpace(parts[0])
		first := strings.TrimSpace(parts[1])
		if last != "" && first != "" {
			flipped := first + " " + last
			candidates = append([]string{flipped}, candidates...)
		}
	}

	noPeriods := strings.ReplaceAll(name, ".", "")
	noPeriods = strings.ReplaceAll(noPeriods, "  ", " ")
	if noPeriods != name {
		candidates = append(candidates, noPeriods)
	}

	noHyphens := strings.ReplaceAll(name, "-", " ")
	if noHyphens != name {
		candidates = append(candidates, noHyphens)
	}

	if strings.Contains(name, ".") && strings.Contains(name, "-") {
		combined := strings.ReplaceAll(name, ".", "")
		combined = strings.ReplaceAll(combined, "-", " ")
		combined = strings.ReplaceAll(combined, "  ", " ")
		candidates = append(candidates, combined)
	}

	return candidates
}

func (s *Service) LookupByName(ctx context.Context, name string) (*ent.Golfer, error) {
	candidates := NormalizeName(name)
	for _, candidate := range candidates {
		g, err := s.db.Golfer.Query().
			Where(golfer.NameEQ(candidate)).
			Only(ctx)
		if err == nil {
			return g, nil
		}
		if !ent.IsNotFound(err) {
			return nil, err
		}
	}
	return nil, nil
}

func (s *Service) BulkLookupByName(ctx context.Context, names []string) (map[string]*ent.Golfer, error) {
	result := make(map[string]*ent.Golfer)
	for _, name := range names {
		g, err := s.LookupByName(ctx, name)
		if err != nil {
			return nil, err
		}
		if g != nil {
			result[name] = g
		}
	}
	return result, nil
}
