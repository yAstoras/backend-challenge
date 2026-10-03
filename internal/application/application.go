package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func Start(
	lifecycle fx.Lifecycle,
	config Config,
	pool *pgxpool.Pool,
	processor *PostgresWagerTransactionProcessor,
	outboxRepository *OutboxRepository,
	sqsConsumer *SQSConsumer,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("Application started:", config.AppName)
			fmt.Println("PostgreSQL connection established")

			go processor.StartReferenceWorker(ctx, 5*time.Second)
			go StartOutboxPublisher(ctx, outboxRepository, 5*time.Second)
			go sqsConsumer.Start(context.Background(), 1*time.Second)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Closing PostgreSQL connection")
			pool.Close()
			return nil
		},
	})
}
