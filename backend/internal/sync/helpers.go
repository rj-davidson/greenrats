package sync

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/getsentry/sentry-go"
)

func (i *Ingester) runSync(ctx context.Context, name string, fn func(context.Context) error) {
	if err := fn(ctx); err != nil {
		if isContextError(err) {
			i.logger.Debug("sync interrupted", "type", name, "error", err)
			return
		}
		i.logger.Error("sync failed", "type", name, "error", err)
		i.captureJobError(name, err)
	}
}

func (i *Ingester) runSyncAsync(ctx context.Context, name string, running *atomic.Bool, fn func(context.Context) error) {
	if !running.CompareAndSwap(false, true) {
		i.logger.Debug("sync already running, skipping", "type", name)
		return
	}
	go func() {
		defer running.Store(false)
		i.runSync(ctx, name, fn)
	}()
}

func (i *Ingester) captureJobError(job string, err error) {
	if err == nil {
		return
	}
	if isContextError(err) {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("job", job)
		sentry.CaptureException(err)
	})
}

func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func formatCurrency(amount int) string {
	if amount >= 1000000 {
		return fmt.Sprintf("$%.2fM", float64(amount)/1000000)
	}
	if amount >= 1000 {
		return fmt.Sprintf("$%dK", amount/1000)
	}
	return fmt.Sprintf("$%d", amount)
}
