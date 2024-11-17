# Money Package for Go

A robust and precise money handling package for Go that provides safe arithmetic operations and formatting for monetary values. This package is designed to handle monetary calculations while avoiding floating-point precision errors.

## Features

- Safe arithmetic operations (Add, Subtract, Multiply, Divide)
- Support for multiple currencies using ISO 4217 currency codes
- Multiple rounding modes (Half Up, Half Down, Up, Down, Half Even/Banker's)
- Amount splitting functionality
- Comprehensive formatting options
- Parse string representations of monetary amounts
- Handles different scales (decimal places) for different currencies

## Installation

```bash
go get github.com/Ceesaxp/money
```

## Usage

### Creating Money Objects

```go
// Create from decimal amount
money, err := money.New(10.99, "USD", money.RoundHalfUp)

// Create from smallest unit (e.g., cents)
cents, err := money.NewFromSmallestUnit(1099, "USD")
```

### Basic Operations

```go
// Addition
sum, err := money1.Add(money2)

// Subtraction
difference, err := money1.Subtract(money2)

// Multiplication
product, err := money.Multiply(2.5, money.RoundHalfUp)

// Division
result, err := money.Divide(3)

// Split amount
parts, err := money.Split(3) // Splits into 3 equal parts
```

### Formatting

```go
money, _ := money.New(1234567.89, "USD", money.RoundHalfUp)

opts := money.FormatOptions{
    Symbol:        "$",
    DecimalSep:    ".",
    ThousandsSep:  ",",
    SymbolFirst:   true,
    ShowCurrency:  true,
    SpaceBetween:  true,
}

formatted := money.Format(opts) // "$1,234,567.89 USD"
```

### Parsing

```go
currency := money.Currency{
    Code: "USD",
    Scale: 2,
    DefaultFormat: money.FormatOptions{
        Symbol: "$",
        DecimalSep: ".",
        ThousandsSep: ",",
    },
}

money, err := money.Parse("$1,234.56", currency, money.RoundHalfUp)
```

## Rounding Modes

The package supports five rounding modes:

- `RoundHalfUp`: Round towards nearest neighbor, ties away from zero
- `RoundHalfDown`: Round towards nearest neighbor, ties towards zero
- `RoundUp`: Always round away from zero
- `RoundDown`: Always round towards zero
- `RoundHalfEven`: Round towards nearest neighbor, ties towards even number (Banker's rounding)

## Error Handling

The package provides several error types for different scenarios:

- `ErrInvalidCurrency`: Invalid currency code
- `ErrInvalidAmount`: Invalid monetary amount
- `ErrorInvalidRoundingMode`: Invalid rounding mode
- `ErrInvalidFactor`: Invalid multiplication factor
- `ErrInvalidDivisor`: Invalid division divisor
- `ErrInvalidSplitParts`: Invalid number of parts for splitting
- `ErrCannotDealWithDifferentCurrencies`: Operation between different currencies
- `ErrParseAmount`: Error parsing string amount

## Thread Safety

The Money type is immutable and safe for concurrent use. All operations return new Money instances rather than modifying existing ones.

## Best Practices

1. Always check returned errors from operations
2. Use appropriate rounding modes for your use case
3. Be aware of currency mixing restrictions
4. Use proper error handling for parsing operations

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This code is under the MIT license. See [LICENSE](LICENSE) for details.

## Credits

Developed by [github.com/Ceesaxp]

## Support

Feel free to add issues to the GitHub repository for any questions or suggestions.
