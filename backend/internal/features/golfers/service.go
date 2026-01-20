package golfers

import (
	"context"
	"strings"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
)

type Service struct {
	db *ent.Client
}

func NewService(db *ent.Client) *Service {
	return &Service{db: db}
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
