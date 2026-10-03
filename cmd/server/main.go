package main

import (
	"go.uber.org/fx"

	"github.com/yasto/backend-challenge/internal/application"
)

func main() {
	app := fx.New(
		fx.Provide(
			application.NewConfig,
			application.NewPostgresPool,
			application.NewPostgresWalletRepository,
			application.NewPostgresLedgerRepository,
			application.NewPostgresWagerTransactionRepository,
			application.NewPostgresWagerTransactionProcessor,
			application.NewOutboxRepository,
			application.NewInboxRepository,
			application.NewSQSClient,
			application.NewSQSQueueURL,
			application.NewSQSConsumer,
		),
		fx.Invoke(application.Start),
	)

	app.Run()
}
