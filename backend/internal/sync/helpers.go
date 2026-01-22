package sync

import (
	"fmt"

	"github.com/getsentry/sentry-go"
)

func (i *Ingester) captureJobError(job string, err error) {
	if err == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("job", job)
		sentry.CaptureException(err)
	})
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
