package money

import (
	"math"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency string
		mode     RoundingMode
		want     *Money
		wantErr  error
	}{
		{
			name:     "valid USD amount",
			amount:   123.45,
			currency: "USD",
			mode:     RoundHalfUp,
			want:     &Money{amount: 12345, currency: "USD", scale: 2, divisor: 100},
			wantErr:  nil,
		},
		{
			name:     "invalid currency",
			amount:   100,
			currency: "INVALID",
			mode:     RoundHalfUp,
			want:     nil,
			wantErr:  ErrInvalidCurrency,
		},
		{
			name:     "NaN amount",
			amount:   math.NaN(),
			currency: "USD",
			mode:     RoundHalfUp,
			want:     nil,
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "Infinite amount",
			amount:   math.Inf(1),
			currency: "USD",
			mode:     RoundHalfUp,
			want:     nil,
			wantErr:  ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.amount, tt.currency, tt.mode)
			if err != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRounding(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		mode     RoundingMode
		currency string
		want     int64
	}{
		{"round half up 1.5", 1.5, RoundHalfUp, "USD", 150},
		{"round half up 1.4", 1.4, RoundHalfUp, "USD", 140},
		{"round half down 1.5", 1.5, RoundHalfDown, "USD", 150},
		{"round half down 1.6", 1.6, RoundHalfDown, "USD", 160},
		{"round up 1.1", 1.1, RoundUp, "USD", 111}, // 1.1 * 100 = 110, but float is 110.00000000000001 -> 111 after ceiling rounding
		{"round down 1.9", 1.9, RoundDown, "USD", 190},
		{"round half even 1.5", 1.5, RoundHalfEven, "USD", 150},
		{"round half even 2.5", 2.5, RoundHalfEven, "USD", 250},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(tt.amount, tt.currency, tt.mode)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if m.SmallestUnit() != tt.want {
				t.Errorf("Got %v, want %v", m.SmallestUnit(), tt.want)
			}
		})
	}
}

func TestMoney_Add(t *testing.T) {
	tests := []struct {
		name    string
		m1      *Money
		m2      *Money
		want    *Money
		wantErr error
	}{
		{
			name: "valid addition",
			m1:   &Money{amount: 100, currency: "USD", scale: 2},
			m2:   &Money{amount: 200, currency: "USD", scale: 2},
			want: &Money{amount: 300, currency: "USD", scale: 2, divisor: 100},
		},
		{
			name:    "different currencies",
			m1:      &Money{amount: 100, currency: "USD", scale: 2},
			m2:      &Money{amount: 200, currency: "EUR", scale: 2},
			want:    nil,
			wantErr: ErrCannotDealWithDifferentCurrencies,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m1.Add(tt.m2)
			if err != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMoney_Split(t *testing.T) {
	tests := []struct {
		name    string
		amount  int64
		parts   int
		want    []int64
		wantErr error
	}{
		{
			name:   "even split",
			amount: 100,
			parts:  2,
			want:   []int64{50, 50},
		},
		{
			name:   "uneven split",
			amount: 100,
			parts:  3,
			want:   []int64{34, 33, 33},
		},
		{
			name:    "invalid parts",
			amount:  100,
			parts:   0,
			want:    nil,
			wantErr: ErrInvalidSplitParts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewFromSmallestUnit(tt.amount, "USD")
			got, err := m.Split(tt.parts)
			if err != tt.wantErr {
				t.Errorf("Split() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				amounts := make([]int64, len(got))
				for i, money := range got {
					amounts[i] = money.amount
				}
				if !reflect.DeepEqual(amounts, tt.want) {
					t.Errorf("Split() = %v, want %v", amounts, tt.want)
				}
			}
		})
	}
}

func TestMoney_Format(t *testing.T) {
	tests := []struct {
		name  string
		money *Money
		opts  FormatOptions
		want  string
	}{
		{
			name:  "standard format",
			money: &Money{amount: 123456, currency: "USD", scale: 2},
			opts: FormatOptions{
				Symbol:       "$",
				DecimalSep:   ".",
				ThousandsSep: ",",
				SymbolFirst:  true,
				ShowCurrency: true,
				SpaceBetween: true,
			},
			want: "$ 1,234.56 USD", // we asked for a space between the symbol and the amount
		},
		{
			name:  "negative amount",
			money: &Money{amount: -123456, currency: "USD", scale: 2},
			opts: FormatOptions{
				Symbol:       "$",
				DecimalSep:   ".",
				ThousandsSep: ",",
				SymbolFirst:  true,
				ShowCurrency: true,
				SpaceBetween: true,
			},
			want: "-$ 1,234.56 USD", // negative sign is part of the symbol
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.money.Format(tt.opts)
			if got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		currency Currency
		mode     RoundingMode
		want     *Money
		wantErr  error
	}{
		{
			name:  "valid amount",
			input: "$1,234.56",
			currency: Currency{
				Code:  "USD",
				Scale: 2,
				DefaultFormat: FormatOptions{
					Symbol:       "$",
					DecimalSep:   ".",
					ThousandsSep: ",",
				},
			},
			mode: RoundHalfUp,
			want: &Money{amount: 123456, currency: "USD", scale: 2, divisor: 100},
		},
		{
			name:  "invalid format",
			input: "invalid",
			currency: Currency{
				Code:  "USD",
				Scale: 2,
			},
			mode:    RoundHalfUp,
			want:    nil,
			wantErr: ErrParseAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input, tt.currency, tt.mode)
			if err != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
