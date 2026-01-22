package timezone

import (
	"fmt"
	"time"
)

const (
	DefaultTimezone     = "America/New_York"
	PickWindowOpenDays  = 4
	PickWindowCloseHour = 5
)

type PickWindow struct {
	OpensAt  time.Time
	ClosesAt time.Time
}

func CalculatePickWindow(startDate time.Time, tzName string) (*PickWindow, error) {
	if tzName == "" {
		tzName = DefaultTimezone
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", tzName, err)
	}

	year, month, day := startDate.UTC().Date()
	localClosesAt := time.Date(year, month, day, PickWindowCloseHour, 0, 0, 0, loc)
	closesAt := localClosesAt.UTC()
	opensAt := closesAt.AddDate(0, 0, -PickWindowOpenDays)

	return &PickWindow{
		OpensAt:  opensAt,
		ClosesAt: closesAt,
	}, nil
}
