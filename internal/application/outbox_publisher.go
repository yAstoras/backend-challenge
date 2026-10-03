package application

import (
	"context"
	"time"
)

func StartOutboxPublisher(
	ctx context.Context,
	repository *OutboxRepository,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			_ = repository.PublishPending(ctx, 100)
		}
	}
}
