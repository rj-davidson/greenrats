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

func DeriveStatus(t *ent.Tournament) DerivedStatus {
	return DeriveStatusAt(t, time.Now().UTC())
}

func DeriveStatusAt(t *ent.Tournament, now time.Time) DerivedStatus {
	referenceTime := t.StartDate
	if t.PickWindowOpensAt != nil {
		referenceTime = *t.PickWindowOpensAt
	}

	endDateMidnight := time.Date(t.EndDate.Year(), t.EndDate.Month(), t.EndDate.Day(), 0, 0, 0, 0, time.UTC)
	completedThreshold := endDateMidnight.AddDate(0, 0, 2)

	if !now.Before(completedThreshold) {
		return StatusCompleted
	}

	upcomingThreshold := referenceTime.Add(-24 * time.Hour)
	if now.Before(upcomingThreshold) {
		return StatusUpcoming
	}

	return StatusActive
}
