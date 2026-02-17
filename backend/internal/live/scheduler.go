package live

import (
	"context"
	"time"
)

// RunScheduler updates live event results every 30 seconds until context is canceled.
func RunScheduler(ctx context.Context, updater *Updater, eventID int64) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		if err := updater.UpdateEvent(ctx, eventID); err != nil {
			// swallow error to keep scheduler alive; next tick retries.
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
