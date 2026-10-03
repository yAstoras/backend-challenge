package application

import (
	"context"
	"errors"
	"time"

	"github.com/yasto/backend-challenge/internal/domain"
)

var ErrUnsupportedTransactionType = errors.New("unsupported transaction type")

type ProcessWagerTransaction struct {
	walletRepository WalletRepository
	ledgerRepository LedgerRepository
}

func NewProcessWagerTransaction(
	walletRepository WalletRepository,
	ledgerRepository LedgerRepository,
) *ProcessWagerTransaction {
	return &ProcessWagerTransaction{
		walletRepository: walletRepository,
		ledgerRepository: ledgerRepository,
	}
}

func (p *ProcessWagerTransaction) Execute(
	ctx context.Context,
	transaction domain.WagerTransaction,
) error {
	wallet, err := p.walletRepository.FindByID(ctx, transaction.WalletID)
	if err != nil {
		return err
	}

	balanceBefore := wallet.Balance

	if err := transaction.StartProcessing(); err != nil {
		return err
	}

	switch transaction.Type {
	case domain.TransactionBet:
		if err := wallet.Debit(transaction.Amount); err != nil {
			return err
		}

	case domain.TransactionWin:
		if err := wallet.Credit(transaction.Amount); err != nil {
			return err
		}

	case domain.TransactionLoss:
		// LOSS não altera o saldo.

	case domain.TransactionRefund:
		if err := wallet.Credit(transaction.Amount); err != nil {
			return err
		}

	case domain.TransactionRollback:
		if err := wallet.Credit(transaction.Amount); err != nil {
			return err
		}

	default:
		return ErrUnsupportedTransactionType
	}

	if err := transaction.Process(); err != nil {
		return err
	}

	if transaction.Type == domain.TransactionBet ||
		transaction.Type == domain.TransactionWin ||
		transaction.Type == domain.TransactionRefund ||
		transaction.Type == domain.TransactionRollback {

		direction := domain.LedgerDebit

		if transaction.Type == domain.TransactionWin ||
			transaction.Type == domain.TransactionRefund ||
			transaction.Type == domain.TransactionRollback {
			direction = domain.LedgerCredit
		}

		ledgerEntry, err := domain.NewWalletLedgerEntry(
			transaction.ID+"-ledger",
			wallet.ID,
			transaction.ID,
			direction,
			transaction.Amount,
			balanceBefore,
			wallet.Balance,
			time.Now(),
		)
		if err != nil {
			return err
		}

		if err := p.ledgerRepository.Save(ctx, ledgerEntry); err != nil {
			return err
		}
	}

	return p.walletRepository.Save(ctx, wallet)
}
