package sprintfjs

import (
	"fmt"
	"strconv"
)

// Number represents a number.
// Similar in Javascript strings are also considered numbers.
type Number struct {
	value interface{}
}

// NewNumber creates a new number.
func NewNumber(value interface{}) Number {
	return Number{value}
}

// Format implements `fmt.Formatter`
func (n Number) Format(f fmt.State, c rune) {
	switch c {
	case 'b':
		fmt.Fprintf(f, "%b", n.value)

	case 'u':
		i64, err := n.Unsigned().Int64()
		if err != nil {
			fmt.Fprintf(f, "%%!u(%T=%v)", n.value, n.value)
			return
		}
		fmt.Fprintf(f, "%d", i64)

	case 'i', 'd':
		i64, err := n.Int64()
		if err != nil {
			fmt.Fprintf(f, "%%!%c(%T=%v)", c, n.value, n.value)
			return
		}
		fmt.Fprintf(f, "%d", i64)

	case 'e', 'f', 'g':
		f64, err := n.Float64()
		if err != nil {
			fmt.Fprintf(f, "%%!%c(%T=%v)", c, n.value, n.value)
			return
		}
		prec, ok := f.Precision()
		if !ok {
			prec = -1
		}
		s := strconv.FormatFloat(f64, byte(c), prec, 64)
		if c == 'e' {
			s = trimExcessZerosFromExponent(s) // "2e+00" => "2e+0"
		}
		fmt.Fprint(f, s)

	case 'x', 'X', 'o':
		i64, err := n.Unsigned().Int64()
		if err != nil {
			fmt.Fprintf(f, "%%!%c(%T=%v)", c, n.value, n.value)
			return
		}
		fmt.Fprintf(f, fmt.Sprintf("%%%c", c), i64)
	}
}

// IsPositive returns true if the number is considered positive.
func (n Number) IsPositive() bool {
	switch v := n.value.(type) {
	case int:
		return v >= 0
	case int8:
		return v >= 0
	case int32:
		return v >= 0
	case int64:
		return v >= 0
	case float32:
		return v >= 0
	case float64:
		return v >= 0
	case uint, uint8, uint32, uint64:
		return true
	case string:
		f64, err := n.Float64()
		if err != nil {
			return false
		}
		return f64 >= 0
	}
	return false
}

// Float64 returns the number as float64, performs conversion if necessary.
func (n Number) Float64() (float64, error) {
	switch v := n.value.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	}
	return 0.0, fmt.Errorf("Cannot use %T as float64", n.value)
}

// Int64 returns the number as int64, performs conversion if necessary.
func (n Number) Int64() (int64, error) {
	switch v := n.value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	}
	return 0.0, fmt.Errorf("Cannot use %T as int64", n.value)
}

// IsNaN returns true if the number is not a number.
func (n Number) IsNaN() bool {
	switch n.value.(type) {
	case int, int8, int32, int64, uint, uint8, uint32, uint64, float32, float64:
		return false
	case string:
		if _, err := n.Float64(); err == nil {
			return false
		}
	}
	return true
}

// Unsigned returns an unsigned version of the number.
func (n Number) Unsigned() Number {
	return NewNumber(unsigned(n.value))
}

func unsigned(v interface{}) interface{} {
	switch v := v.(type) {
	case int:
		return uint32(v)
	case int8:
		return uint32(v)
	case int32:
		return uint32(v)
	case int64:
		return uint64(v)
	}
	return v
}

// trimExcessZerosFromExponent removes duplicate zeros for a zero exponent: 2e+00 => 2e+0
func trimExcessZerosFromExponent(s string) string {
	l := len(s) -1
	for l > 0 {
		c := s[l]
		if c == '+' {
			return s[:l+2] // key only the 1st "0" after "+"
		}
		if c != '0' {
			break
		}
		l--
	}
	return s
}