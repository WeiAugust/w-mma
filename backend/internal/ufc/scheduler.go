package ufc

import (
	"context"
	"log"
	"time"
)

func StartScheduler(ctx context.Context, svc *Service, interval time.Duration) {
	if svc == nil {
		return
	}
	if interval <= 0 {
		interval = 6 * time.Hour
	}
	if _, err := svc.SyncEnabledSources(ctx); err != nil {
		log.Printf("ufc initial sync failed: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := svc.SyncEnabledSources(ctx); err != nil {
				log.Printf("ufc sync failed: %v", err)
			}
		}
	}
}
