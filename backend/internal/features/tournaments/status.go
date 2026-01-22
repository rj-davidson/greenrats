package tournaments

import (
	"time"

	"github.com/rj-davidson/greenrats/ent"
)

type DerivedStatus string

const (
	StatusUpcoming  DerivedStatus = "upcoming"
	StatusActive    DerivedStatus = "active"
	StatusCompleted DerivedStatus = "completed"
)

func DeriveStatus(t *ent.Tournament, hasChampion bool) DerivedStatus {
	if hasChampion {
		return StatusCompleted
	}

	now := time.Now().UTC()
	if t.PickWindowClosesAt != nil && now.After(*t.PickWindowClosesAt) {
		return StatusActive
	}
	return StatusUpcoming
}
