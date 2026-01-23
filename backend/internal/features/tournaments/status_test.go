package tournaments

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/rj-davidson/greenrats/ent"
)

func TestDeriveStatusAt(t *testing.T) {
	tests := []struct {
		name              string
		startDate         time.Time
		endDate           time.Time
		pickWindowOpensAt *time.Time
		now               time.Time
		expectedStatus    DerivedStatus
	}{
		{
			name:           "upcoming - more than 1 day before start_date",
			startDate:      time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			now:            time.Date(2024, 1, 23, 12, 0, 0, 0, time.UTC),
			expectedStatus: StatusUpcoming,
		},
		{
			name:              "upcoming - more than 1 day before pick_window_opens_at",
			startDate:         time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			endDate:           time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			pickWindowOpensAt: ptrTime(time.Date(2024, 1, 22, 8, 0, 0, 0, time.UTC)),
			now:               time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC),
			expectedStatus:    StatusUpcoming,
		},
		{
			name:           "active - within 1 day of start_date",
			startDate:      time.Date(2024, 1, 25, 8, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			now:            time.Date(2024, 1, 24, 10, 0, 0, 0, time.UTC),
			expectedStatus: StatusActive,
		},
		{
			name:              "active - within 1 day of pick_window_opens_at",
			startDate:         time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			endDate:           time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			pickWindowOpensAt: ptrTime(time.Date(2024, 1, 22, 8, 0, 0, 0, time.UTC)),
			now:               time.Date(2024, 1, 21, 10, 0, 0, 0, time.UTC),
			expectedStatus:    StatusActive,
		},
		{
			name:           "active - during tournament",
			startDate:      time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			now:            time.Date(2024, 1, 26, 14, 0, 0, 0, time.UTC),
			expectedStatus: StatusActive,
		},
		{
			name:           "active - same day as end_date",
			startDate:      time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			now:            time.Date(2024, 1, 28, 18, 0, 0, 0, time.UTC),
			expectedStatus: StatusActive,
		},
		{
			name:           "active - 1 day after end_date (grace period)",
			startDate:      time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			now:            time.Date(2024, 1, 29, 12, 0, 0, 0, time.UTC),
			expectedStatus: StatusActive,
		},
		{
			name:           "active - just before midnight on day after end_date",
			startDate:      time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			now:            time.Date(2024, 1, 29, 23, 59, 59, 0, time.UTC),
			expectedStatus: StatusActive,
		},
		{
			name:           "completed - at midnight 2 days after end_date",
			startDate:      time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			now:            time.Date(2024, 1, 30, 0, 0, 0, 0, time.UTC),
			expectedStatus: StatusCompleted,
		},
		{
			name:           "completed - well after end_date",
			startDate:      time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			now:            time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC),
			expectedStatus: StatusCompleted,
		},
		{
			name:           "active - exactly at boundary (1 day before start)",
			startDate:      time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			now:            time.Date(2024, 1, 24, 12, 0, 0, 0, time.UTC),
			expectedStatus: StatusActive,
		},
		{
			name:           "upcoming - exactly 1 nanosecond before boundary",
			startDate:      time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC),
			now:            time.Date(2024, 1, 24, 11, 59, 59, 999999999, time.UTC),
			expectedStatus: StatusUpcoming,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tournament := &ent.Tournament{
				StartDate:         tt.startDate,
				EndDate:           tt.endDate,
				PickWindowOpensAt: tt.pickWindowOpensAt,
			}

			status := DeriveStatusAt(tournament, tt.now)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
