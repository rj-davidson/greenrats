package timezone

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculatePickWindow(t *testing.T) {
	tests := []struct {
		name         string
		startDate    time.Time
		tzName       string
		wantOpensAt  time.Time
		wantClosesAt time.Time
		wantErr      bool
	}{
		{
			name:         "New York timezone - standard time",
			startDate:    time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC),
			tzName:       "America/New_York",
			wantClosesAt: time.Date(2025, 1, 16, 10, 0, 0, 0, time.UTC),
			wantOpensAt:  time.Date(2025, 1, 12, 10, 0, 0, 0, time.UTC),
		},
		{
			name:         "Los Angeles timezone - standard time",
			startDate:    time.Date(2025, 2, 13, 0, 0, 0, 0, time.UTC),
			tzName:       "America/Los_Angeles",
			wantClosesAt: time.Date(2025, 2, 13, 13, 0, 0, 0, time.UTC),
			wantOpensAt:  time.Date(2025, 2, 9, 13, 0, 0, 0, time.UTC),
		},
		{
			name:         "Phoenix timezone - no DST",
			startDate:    time.Date(2025, 2, 6, 0, 0, 0, 0, time.UTC),
			tzName:       "America/Phoenix",
			wantClosesAt: time.Date(2025, 2, 6, 12, 0, 0, 0, time.UTC),
			wantOpensAt:  time.Date(2025, 2, 2, 12, 0, 0, 0, time.UTC),
		},
		{
			name:         "Hawaii timezone",
			startDate:    time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC),
			tzName:       "Pacific/Honolulu",
			wantClosesAt: time.Date(2025, 1, 9, 15, 0, 0, 0, time.UTC),
			wantOpensAt:  time.Date(2025, 1, 5, 15, 0, 0, 0, time.UTC),
		},
		{
			name:         "empty timezone defaults to New York",
			startDate:    time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC),
			tzName:       "",
			wantClosesAt: time.Date(2025, 1, 16, 10, 0, 0, 0, time.UTC),
			wantOpensAt:  time.Date(2025, 1, 12, 10, 0, 0, 0, time.UTC),
		},
		{
			name:      "invalid timezone",
			startDate: time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC),
			tzName:    "Invalid/Timezone",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculatePickWindow(tt.startDate, tt.tzName)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantOpensAt, result.OpensAt, "opensAt mismatch")
			assert.Equal(t, tt.wantClosesAt, result.ClosesAt, "closesAt mismatch")
		})
	}
}

func TestCalculatePickWindow_DST(t *testing.T) {
	t.Run("New York during DST", func(t *testing.T) {
		startDate := time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC)
		result, err := CalculatePickWindow(startDate, "America/New_York")

		require.NoError(t, err)
		assert.Equal(t, time.Date(2025, 4, 10, 9, 0, 0, 0, time.UTC), result.ClosesAt)
		assert.Equal(t, time.Date(2025, 4, 6, 9, 0, 0, 0, time.UTC), result.OpensAt)
	})

	t.Run("Los Angeles during DST", func(t *testing.T) {
		startDate := time.Date(2025, 6, 19, 0, 0, 0, 0, time.UTC)
		result, err := CalculatePickWindow(startDate, "America/Los_Angeles")

		require.NoError(t, err)
		assert.Equal(t, time.Date(2025, 6, 19, 12, 0, 0, 0, time.UTC), result.ClosesAt)
		assert.Equal(t, time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC), result.OpensAt)
	})
}

func TestCalculatePickWindow_WindowDuration(t *testing.T) {
	startDate := time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC)
	result, err := CalculatePickWindow(startDate, "America/New_York")

	require.NoError(t, err)
	duration := result.ClosesAt.Sub(result.OpensAt)
	assert.Equal(t, 4*24*time.Hour, duration, "window should be exactly 4 days")
}
