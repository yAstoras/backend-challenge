package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

var (
	ErrInvalidAmount    = errors.New("invalid amount")
	ErrCurrencyMismatch = errors.New("currency mismatch")
	ErrMoneyOverflow    = errors.New("money overflow")
)

type Money struct {
	Amount   int64
	Currency string
}

func NewMoney(amount int64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

func Zero(currency string) Money {
	return Money{
		Amount:   0,
		Currency: currency,
	}
}

func NewMoneyFromString(amount string, currency string) (Money, error) {
	if !isValidCurrency(currency) {
		return Money{}, ErrInvalidAmount
	}

	if strings.ContainsAny(amount, "eE") {
		return Money{}, ErrInvalidAmount
	}

	parts := strings.Split(amount, ".")

	if len(parts) > 2 {
		return Money{}, ErrInvalidAmount
	}

	if len(parts) == 2 && len(parts[1]) > 2 {
		return Money{}, ErrInvalidAmount
	}

	if parts[0] == "" {
		return Money{}, ErrInvalidAmount
	}

	if len(parts) == 2 && parts[1] == "" {
		return Money{}, ErrInvalidAmount
	}

	if amount == "" {
		return Money{}, ErrInvalidAmount
	}

	if strings.HasPrefix(amount, "+") {
		return Money{}, ErrInvalidAmount
	}

	if strings.ContainsAny(amount, "eE") {
		return Money{}, ErrInvalidAmount
	}

	value := new(big.Rat)

	if _, ok := value.SetString(amount); !ok {
		return Money{}, ErrInvalidAmount
	}

	scaled := new(big.Rat).Mul(value, big.NewRat(100, 1))

	if !scaled.IsInt() {
		return Money{}, ErrInvalidAmount
	}
	if scaled.Sign() < 0 {
		return Money{}, ErrInvalidAmount
	}

	if !scaled.Num().IsInt64() {
		return Money{}, ErrInvalidAmount
	}

	return Money{
		Amount:   scaled.Num().Int64(),
		Currency: currency,
	}, nil
}

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}

	result := m.Amount + other.Amount

	if (other.Amount > 0 && result < m.Amount) ||
		(other.Amount < 0 && result > m.Amount) {
		return Money{}, ErrMoneyOverflow
	}

	return Money{
		Amount:   result,
		Currency: m.Currency,
	}, nil
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}

	result := m.Amount - other.Amount

	if (other.Amount < 0 && result < m.Amount) ||
		(other.Amount > 0 && result > m.Amount) {
		return Money{}, ErrMoneyOverflow
	}

	return Money{
		Amount:   result,
		Currency: m.Currency,
	}, nil
}

func isValidCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}

	for _, character := range currency {
		if character < 'A' || character > 'Z' {
			return false
		}
	}

	return true
}

func (m Money) Negate() (Money, error) {
	if m.Amount == -9223372036854775808 {
		return Money{}, ErrMoneyOverflow
	}

	return Money{
		Amount:   -m.Amount,
		Currency: m.Currency,
	}, nil
}

func (m Money) Compare(other Money) (int, error) {
	if m.Currency != other.Currency {
		return 0, ErrCurrencyMismatch
	}

	if m.Amount < other.Amount {
		return -1, nil
	}

	if m.Amount > other.Amount {
		return 1, nil
	}

	return 0, nil
}

func (m Money) String() string {
	sign := ""
	var amount uint64

	if m.Amount < 0 {
		sign = "-"
		amount = uint64(-(m.Amount + 1)) + 1
	} else {
		amount = uint64(m.Amount)
	}

	reais := amount / 100
	centavos := amount % 100

	return fmt.Sprintf("%s%d.%02d", sign, reais, centavos)
}
