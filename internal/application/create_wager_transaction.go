package application

import (
	"context"
	"errors"

	"github.com/yasto/backend-challenge/internal/domain"
)

type CreateWagerTransaction struct {
	repository WagerTransactionRepository
}

func NewCreateWagerTransaction(
	repository WagerTransactionRepository,
) *CreateWagerTransaction {
	return &CreateWagerTransaction{
		repository: repository,
	}
}

func (c *CreateWagerTransaction) Execute(
	ctx context.Context,
	transaction domain.WagerTransaction,
) (domain.WagerTransaction, error) {

	existing, err := c.repository.FindByIdempotencyKey(
		ctx,
		transaction.IdempotencyKey,
	)

	if err == nil {
		if existing.PayloadHash != transaction.PayloadHash {
			return domain.WagerTransaction{}, ErrIdempotencyConflict
		}

		return existing, nil
	}

	if !errors.Is(err, ErrWagerTransactionNotFound) {
		return domain.WagerTransaction{}, err
	}

	existing, err = c.repository.FindByProviderExternalID(
		ctx,
		transaction.ProviderID,
		transaction.ExternalTransactionID,
	)

	if err == nil {
		return domain.WagerTransaction{}, ErrProviderTransactionConflict
	}

	if !errors.Is(err, ErrWagerTransactionNotFound) {
		return domain.WagerTransaction{}, err
	}

	err = c.repository.Create(ctx, transaction)
	if err != nil {
		return domain.WagerTransaction{}, err
	}

	return transaction, nil
}
