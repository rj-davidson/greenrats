package sync

import (
	"io"
	"log/slog"
	"testing"

	"github.com/rj-davidson/greenrats/internal/testutil"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	db := testutil.NewTestDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(db, nil, logger)
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
