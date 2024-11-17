package money

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// RoundingMode defines how monetary amounts should be rounded
type RoundingMode int

const (
	RoundHalfUp RoundingMode = iota
	RoundHalfDown
	RoundUp
	RoundDown
	RoundHalfEven // Also known as "banker's rounding"
)

// Money represents a monetary amount in the smallest currency unit (e.g., cents)
type Money struct {
	amount   int64  // Store amount in smallest currency unit
	currency string // ISO 4217 currency code
	scale    int    // Number of decimal places
	divisor  int64  // Divisor for converting to decimal
}

var (
	ErrInvalidCurrency                   = errors.New("invalid currency")
	ErrInvalidAmount                     = errors.New("invalid amount")
	ErrorInvalidRoundingMode             = errors.New("invalid rounding mode")
	ErrInvalidFactor                     = errors.New("invalid factor")
	ErrInvalidDivisor                    = errors.New("invalid divisor")
	ErrInvalidSplitParts                 = errors.New("invalid number of split parts")
	ErrCannotDealWithDifferentCurrencies = errors.New("cannot deal with different currencies")
	ErrParseAmount                       = errors.New("error parsing amount")

	numberRegex = regexp.MustCompile(`^-?\d*\.?\d+$`)
)

// round applies the specified rounding mode to a float64
func round(amount float64, scale int, mode RoundingMode) int64 {
	multiplier := math.Pow10(scale)
	switch mode {
	case RoundHalfUp:
		return int64(math.Round(amount * multiplier))
	case RoundHalfDown:
		return int64(amount*multiplier + 0.5 - 0.00001)
	case RoundUp:
		return int64(math.Ceil(amount * multiplier))
	case RoundDown:
		return int64(math.Floor(amount * multiplier))
	case RoundHalfEven:
		scaled := amount * multiplier
		_, frac := math.Modf(scaled)
		if math.Abs(frac) == 0.5 {
			if int64(scaled)%2 == 0 {
				return int64(math.Floor(scaled))
			}
			return int64(math.Ceil(scaled))
		}
		return int64(math.Round(scaled))
	default:
		return int64(math.Round(amount * multiplier))
	}
}

// New creates a new Money instance from a decimal amount
func New(amount float64, currencyCode string, mode RoundingMode) (*Money, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return nil, errors.New("invalid amount")
	}

	// Upper-case currency code
	currencyCode = strings.ToUpper(currencyCode)

	currency, ok := Currencies[currencyCode]
	if !ok {
		return nil, ErrInvalidCurrency
	}

	// Convert to cents (or smallest currency unit)
	cents := round(amount*math.Pow10(currency.Scale), currency.Scale, mode)

	return &Money{
		amount:   cents,
		currency: currency.Code,
		scale:    currency.Scale,
		divisor:  int64(math.Pow10(currency.Scale)),
	}, nil
}

// NewFromCents creates a new Money instance from an amount in cents
func NewFromSmallestUnit(cents int64, currencyCode string) (*Money, error) {
	currency, ok := Currencies[currencyCode]
	if !ok {
		return nil, ErrInvalidCurrency
	}

	return &Money{
		amount:   cents,
		currency: currencyCode,
		scale:    currency.Scale,
		divisor:  int64(math.Pow10(currency.Scale)),
	}, nil
}

// Amount returns the decimal representation of the monetary amount
func (m *Money) Amount() float64 {
	return float64(m.amount) / float64(m.divisor)
}

// Cents returns the amount in cents
func (m *Money) SmallestUnit() int64 {
	return m.amount
}

// Currency returns the currency code
func (m *Money) Currency() string {
	return m.currency
}

// String returns a formatted string representation
func (m *Money) String() string {
	amount := float64(m.amount) / float64(m.divisor)
	formatString := fmt.Sprintf("%%.%df %%s", m.scale)
	return fmt.Sprintf(formatString, amount, m.currency)
}

// Add adds two monetary amounts of the same currency
func (m *Money) Add(other *Money) (*Money, error) {
	if m.currency != other.currency {
		return nil, ErrCannotDealWithDifferentCurrencies
	}

	amount, err := NewFromSmallestUnit(m.amount+other.amount, m.currency)
	if err != nil {
		return nil, err
	}
	return amount, nil
}

// Subtract subtracts two monetary amounts of the same currency
func (m *Money) Subtract(other *Money) (*Money, error) {
	if m.currency != other.currency {
		return nil, ErrCannotDealWithDifferentCurrencies
	}

	amount, err := NewFromSmallestUnit(m.amount-other.amount, m.currency)
	if err != nil {
		return nil, err
	}
	return amount, nil
}

// Multiply multiplies the monetary amount by a factor
func (m *Money) Multiply(factor float64, mode RoundingMode) (*Money, error) {
	if math.IsNaN(factor) || math.IsInf(factor, 0) {
		return nil, ErrInvalidFactor
	}

	newAmount := round(float64(m.amount)*factor, m.scale, mode)
	amount, err := NewFromSmallestUnit(newAmount, m.currency)
	if err != nil {
		return nil, err
	}
	return amount, nil
}

// Divide divides the monetary amount by a factor
func (m *Money) Divide(divisor float64) (*Money, error) {
	if math.IsNaN(divisor) || math.IsInf(divisor, 0) {
		return nil, ErrInvalidDivisor
	}

	newAmount := int64(math.Round(float64(m.amount) / divisor))
	amount, err := NewFromSmallestUnit(newAmount, m.currency)
	if err != nil {
		return nil, err
	}
	return amount, nil
}

// Split divides an amount into n equal parts
func (m *Money) Split(n int) ([]*Money, error) {
	if n <= 0 {
		return nil, ErrInvalidSplitParts
	}

	// Calculate the base amount for each part
	baseAmount := m.amount / int64(n)
	remainder := m.amount % int64(n)

	results := make([]*Money, n)
	for i := 0; i < n; i++ {
		amount := baseAmount
		if int64(i) < remainder {
			amount++
		}
		results[i] = &Money{
			amount:   amount,
			currency: m.currency,
			scale:    m.scale,
		}
	}

	return results, nil
}

// Parse creates a Money instance from a string representation
func Parse(s string, currency Currency, mode RoundingMode) (*Money, error) {
	// Remove currency symbol, thousands separators, and normalize decimal separator
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, currency.DefaultFormat.Symbol, "")
	s = strings.ReplaceAll(s, currency.DefaultFormat.ThousandsSep, "")
	s = strings.ReplaceAll(s, currency.DefaultFormat.DecimalSep, ".")
	s = strings.TrimSpace(s)

	if !numberRegex.MatchString(s) {
		return nil, ErrParseAmount
	}

	amount, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseAmount, err)
	}

	return New(amount, currency.Code, mode)
}

// Format returns a formatted string representation using the provided options
func (m *Money) Format(opts FormatOptions) string {
	// Handle negative amounts
	sign := ""
	absAmount := m.amount
	if m.amount < 0 {
		sign = "-"
		absAmount = -m.amount
	}

	// Convert to decimal string with proper scale
	value := strconv.FormatInt(absAmount, 10)
	scale := m.scale

	// Pad with leading zeros if necessary
	for len(value) <= scale {
		value = "0" + value
	}

	// Insert decimal point
	decimalPos := len(value) - scale
	intPart := value[:decimalPos]
	fracPart := value[decimalPos:]

	// Add thousands separators
	if opts.ThousandsSep != "" {
		var result []string
		for i := len(intPart); i > 0; i -= 3 {
			start := max(0, i-3)
			group := intPart[start:i]
			if len(result) > 0 {
				result = append([]string{group}, result...)
			} else {
				result = []string{group}
			}
		}
		intPart = strings.Join(result, opts.ThousandsSep)
	}

	// Combine parts
	var result string
	if scale > 0 {
		result = intPart + opts.DecimalSep + fracPart
	} else {
		result = intPart
	}

	// Add currency symbol and code
	space := ""
	if opts.SpaceBetween {
		space = " "
	}

	code := ""
	if opts.ShowCurrency {
		code = space + m.currency
	}

	if opts.SymbolFirst {
		return sign + opts.Symbol + space + result + code
	}
	return sign + result + space + opts.Symbol + code
}
