package domain

import "testing"

func TestMoneyAdd(t *testing.T) {
	moneyA := NewMoney(1000, "BRL")
	moneyB := NewMoney(500, "BRL")

	result, err := moneyA.Add(moneyB)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Amount != 1500 {
		t.Fatalf("expected 1500, got %d", result.Amount)
	}

	if result.Currency != "BRL" {
		t.Fatalf("expected BRL, got %s", result.Currency)
	}
}

func TestMoneyAddOverflow(t *testing.T) {
	moneyA := NewMoney(9223372036854775807, "BRL")
	moneyB := NewMoney(1, "BRL")

	_, err := moneyA.Add(moneyB)

	if err != ErrMoneyOverflow {
		t.Fatalf("expected money overflow error, got %v", err)
	}
}

func TestMoneySubtractOverflow(t *testing.T) {
	moneyA := NewMoney(-9223372036854775808, "BRL")
	moneyB := NewMoney(1, "BRL")

	_, err := moneyA.Subtract(moneyB)

	if err != ErrMoneyOverflow {
		t.Fatalf("expected money overflow error, got %v", err)
	}
}

func TestNewMoneyFromString(t *testing.T) {
	money, err := NewMoneyFromString("25.50", "BRL")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if money.Amount != 2550 {
		t.Fatalf("expected 2550, got %d", money.Amount)
	}

	if money.Currency != "BRL" {
		t.Fatalf("expected BRL, got %s", money.Currency)
	}
}

func TestNewMoneyFromStringRejectsTooManyDecimals(t *testing.T) {
	_, err := NewMoneyFromString("25.123", "BRL")

	if err != ErrInvalidAmount {
		t.Fatalf("expected invalid amount error, got %v", err)
	}
}

func TestNewMoneyFromStringRejectsNegative(t *testing.T) {
	_, err := NewMoneyFromString("-25.00", "BRL")

	if err != ErrInvalidAmount {
		t.Fatalf("expected invalid amount error, got %v", err)
	}
}

func TestNewMoneyFromStringRejectsInvalidCurrency(t *testing.T) {
	_, err := NewMoneyFromString("25.00", "BR")

	if err != ErrInvalidAmount {
		t.Fatalf("expected invalid amount error, got %v", err)
	}

	_, err = NewMoneyFromString("25.00", "brl")

	if err != ErrInvalidAmount {
		t.Fatalf("expected invalid amount error, got %v", err)
	}
}

func TestNewMoneyFromStringRejectsScientificNotation(t *testing.T) {
	_, err := NewMoneyFromString("1e2", "BRL")

	if err != ErrInvalidAmount {
		t.Fatalf("expected invalid amount error, got %v", err)
	}

	_, err = NewMoneyFromString("1E2", "BRL")

	if err != ErrInvalidAmount {
		t.Fatalf("expected invalid amount error, got %v", err)
	}
}

func TestNewMoneyFromStringRejectsMoreThanTwoDecimals(t *testing.T) {
	_, err := NewMoneyFromString("25.123", "BRL")

	if err != ErrInvalidAmount {
		t.Fatalf("expected invalid amount error, got %v", err)
	}

	_, err = NewMoneyFromString("25.50.10", "BRL")

	if err != ErrInvalidAmount {
		t.Fatalf("expected invalid amount error, got %v", err)
	}
}

func TestNewMoneyFromStringRejectsMalformedDecimal(t *testing.T) {
	invalidAmounts := []string{
		".50",
		"25.",
		"25.50.10",
	}

	for _, amount := range invalidAmounts {
		_, err := NewMoneyFromString(amount, "BRL")

		if err != ErrInvalidAmount {
			t.Fatalf("expected invalid amount for %q, got %v", amount, err)
		}
	}
}

func TestMoneyNegate(t *testing.T) {
	money := NewMoney(2500, "BRL")

	result, err := money.Negate()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Amount != -2500 {
		t.Fatalf("expected -2500, got %d", result.Amount)
	}

	if result.Currency != "BRL" {
		t.Fatalf("expected BRL, got %s", result.Currency)
	}
}

func TestMoneyNegateOverflow(t *testing.T) {
	money := NewMoney(-9223372036854775808, "BRL")

	_, err := money.Negate()

	if err != ErrMoneyOverflow {
		t.Fatalf("expected money overflow error, got %v", err)
	}
}

func TestMoneyDoubleNegate(t *testing.T) {
	money := NewMoney(2500, "BRL")

	result, err := money.Negate()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err = result.Negate()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Amount != 2500 {
		t.Fatalf("expected 2500, got %d", result.Amount)
	}
}

func TestMoneyCompare(t *testing.T) {
	tests := []struct {
		name     string
		first    Money
		second   Money
		expected int
	}{
		{
			name:     "first is smaller",
			first:    NewMoney(5000, "BRL"),
			second:   NewMoney(10000, "BRL"),
			expected: -1,
		},
		{
			name:     "values are equal",
			first:    NewMoney(10000, "BRL"),
			second:   NewMoney(10000, "BRL"),
			expected: 0,
		},
		{
			name:     "first is greater",
			first:    NewMoney(15000, "BRL"),
			second:   NewMoney(10000, "BRL"),
			expected: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.first.Compare(test.second)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != test.expected {
				t.Fatalf("expected %d, got %d", test.expected, result)
			}
		})
	}
}

func TestMoneyCompareDifferentCurrencies(t *testing.T) {
	moneyBRL := NewMoney(10000, "BRL")
	moneyUSD := NewMoney(10000, "USD")

	_, err := moneyBRL.Compare(moneyUSD)

	if err != ErrCurrencyMismatch {
		t.Fatalf("expected currency mismatch error, got %v", err)
	}
}

func TestMoneyZero(t *testing.T) {
	money := Zero("BRL")

	if money.Amount != 0 {
		t.Fatalf("expected 0, got %d", money.Amount)
	}

	if money.Currency != "BRL" {
		t.Fatalf("expected BRL, got %s", money.Currency)
	}
}

func TestMoneyString(t *testing.T) {
	tests := []struct {
		name     string
		money    Money
		expected string
	}{
		{
			name:     "twenty five reais and fifty cents",
			money:    NewMoney(2550, "BRL"),
			expected: "25.50",
		},
		{
			name:     "ten reais",
			money:    NewMoney(1000, "BRL"),
			expected: "10.00",
		},
		{
			name:     "one cent",
			money:    NewMoney(1, "BRL"),
			expected: "0.01",
		},
		{
			name:     "zero",
			money:    NewMoney(0, "BRL"),
			expected: "0.00",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.money.String()

			if result != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestMoneyStringNegative(t *testing.T) {
	money := NewMoney(-2550, "BRL")

	result := money.String()

	if result != "-25.50" {
		t.Fatalf("expected -25.50, got %s", result)
	}
}

func TestMoneyStringMinInt64(t *testing.T) {
	money := NewMoney(-9223372036854775808, "BRL")

	result := money.String()

	expected := "-92233720368547758.08"

	if result != expected {
		t.Fatalf("expected %s, got %s", expected, result)
	}
}

func TestNewMoneyFromStringRejectsPositiveSign(t *testing.T) {
	_, err := NewMoneyFromString("+25.00", "BRL")

	if err != ErrInvalidAmount {
		t.Fatalf("expected invalid amount error, got %v", err)
	}
}
