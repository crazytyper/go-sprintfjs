// Package sprintfjs implements sprintfs compatible formatting.
// See https://github.com/alexei/sprintf.js
package sprintfjs

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	// reNotString    = regexp.MustCompile("[^s]")
	// reNotBool      = regexp.MustCompile("[^t]")
	reNotType      = regexp.MustCompile("[^T]")
	reNotPrimitive = regexp.MustCompile("[^v]")
	reNumericArg   = regexp.MustCompile("[bcdiefguxX]")
	reNotJSON      = regexp.MustCompile("[^j]")
	reJSON         = regexp.MustCompile("[j]")
	reSign         = regexp.MustCompile("^[+-]")
	reNumber       = regexp.MustCompile("[diefg]")
	reText         = regexp.MustCompile("^[^\x25]+")
	reModulo       = regexp.MustCompile("^\x25{2}")
	rePlaceholder  = regexp.MustCompile(`^\x25(?:([1-9]\d*)\$|\(([^)]+)\))?(\+)?(0|'[^$])?(-)?(\d+)?(?:\.(\d+))?([b-gijostTuvxX])`)
	reKey          = regexp.MustCompile(`^(?i:([a-z_][a-z_\d]*))`)
	reKeyAccess    = regexp.MustCompile(`^\.(?i:([a-z_][a-z_\d]*))`)
	reIndexAccess  = regexp.MustCompile(`^\[(\d+)\]`)
)

// ASTNode is a node in the abstract syntax tree
type ASTNode struct {
	Text        string
	Placeholder string
	ParamNo     int
	Keys        []string
	Sign        string
	Pad         string
	Align       string
	Width       int
	Precision   string
	Type        string
}

// AST is an abstract syntax tree
type AST []ASTNode

// Parse parses a format string into an abstract syntax tree.
func Parse(format string) (AST, error) {
	ast := AST{}
	argNames := 0

	for len(format) > 0 {
		l := 0
		if match := reText.FindAllString(format, 1); len(match) > 0 {
			ast = append(ast, ASTNode{Text: match[0]})
			l = len(match[0])
		} else if match := reModulo.FindAllString(format, 1); len(match) > 0 {
			ast = append(ast, ASTNode{Text: "%"})
			l = len(match[0])
		} else if ms := rePlaceholder.FindAllStringSubmatch(format, 1); len(ms) > 0 {
			m := ms[0]
			l = len(m[0])
			node := ASTNode{
				Placeholder: m[0],
				Sign:        m[3],
				Pad:         m[4],
				Align:       m[5],
				Precision:   m[7],
				Type:        m[8],
			}

			if m[1] != "" {
				paramNo, err := strconv.Atoi(m[1])
				if err != nil {
					return nil, fmt.Errorf("[sprintf] failed to parse positional argument %q: %v", m[1], err)
				}
				node.ParamNo = paramNo
			}
			if m[6] != "" {
				width, err := strconv.Atoi(m[6])
				if err != nil {
					return nil, fmt.Errorf("[sprintf] failed to parse width %q: %v", m[6], err)
				}
				node.Width = width
			}

			if m[2] != "" {
				argNames |= 1
				keys := []string{}
				keyNames := m[2]

				if ms := reKey.FindAllStringSubmatch(keyNames, 1); len(ms) > 0 {
					m := ms[0]
					keys = append(keys, m[1])
					keyLen := len(m[0])
					for {
						keyNames = keyNames[keyLen:]
						if keyNames == "" {
							break
						}

						if ms := reKeyAccess.FindAllStringSubmatch(keyNames, 1); len(ms) > 0 {
							keys = append(keys, ms[0][1])
							keyLen = len(ms[0][0])
						} else if ms := reIndexAccess.FindAllStringSubmatch(keyNames, 1); len(ms) > 0 {
							keyLen = len(ms[0][0])
						} else {
							return nil, errors.New("[sprintf] failed to parse named argument key")
						}
					}
				} else {
					return nil, errors.New("[sprintf] failed to parse named argument key")
				}
				node.Keys = keys
			} else {
				argNames |= 2
			}

			if argNames == 3 {
				return nil, errors.New("[sprintf] mixing positional and named placeholders is not (yet) supported")
			}

			ast = append(ast, node)
		} else {
			return nil, errors.New("[sprintf] unexpected placeholder")
		}

		if l >= len(format) {
			break
		}
		format = format[l:]
	}
	return ast, nil
}

// Format formats a string based on the instructions in `format` using the values in `args`.
//  ## Format specification
//  The placeholders in the format string are marked by % and are followed by one or more of these elements, in this order:
//  * An optional number followed by a $ sign that selects which argument index to use for the value.
//    If not specified, arguments will be placed in the same order as the placeholders in the input string.
//  * An optional + sign that forces to preceed the result with a plus or minus sign on numeric values.
//    By default, only the - sign is used on negative numbers.
//  * An optional padding specifier that says what character to use for padding (if specified).
//    Possible values are 0 or any other character precedeed by a ' (single quote). The default is to pad with spaces.
//  * An optional - sign, that causes sprintf to left-align the result of this placeholder.
//    The default is to right-align the result.
//  * An optional number, that says how many characters the result should have.
//    If the value to be returned is shorter than this number, the result will be padded.
//    When used with the j (JSON) type specifier, the padding length specifies the tab size used for indentation.
//  * An optional precision modifier, consisting of a . (dot) followed by a number, that says how many digits should be displayed for floating point numbers.
//    When used with the g type specifier, it specifies the number of significant digits.
//    When used on a string, it causes the result to be truncated.
//  * A type specifier that can be any of:
//    * % — yields a literal % character
//    * b — yields an integer as a binary number
//    * c — yields an integer as the character with that ASCII value
//    * d or i — yields an integer as a signed decimal number
//    * e — yields a float using scientific notation
//    * u — yields an integer as an unsigned decimal number
//    * f — yields a float as is; see notes on precision above
//    * g — yields a float as is; see notes on precision above
//    * o — yields an integer as an octal number
//    * s — yields a string as is
//    * t — yields true or false
//    * T — yields the type of the argument1
//    * v — yields the primitive value of the specified argument
//    * x — yields an integer as a hexadecimal number (lower-case)
//    * X — yields an integer as a hexadecimal number (upper-case)
//    * j — yields a JavaScript object or array as a JSON encoded string
func Format(format string, args ...interface{}) (string, error) {
	ast, err := Parse(format)
	if err != nil {
		return "", err
	}
	return FormatAST(ast, args...)
}

// FormatAST formats an abstract syntax tree returned by `Parse`.
func FormatAST(ast AST, args ...interface{}) (string, error) {
	cursor := 0

	output := strings.Builder{}

	for _, node := range ast {
		if node.Text != "" {
			output.WriteString(node.Text)
		} else {
			arg, nextCursor, err := argumentValue(node, args, cursor)
			if err != nil {
				return "", err
			}
			cursor = nextCursor

			f, err := formatPlaceholder(node, arg)
			if err != nil {
				return "", err
			}

			if _, err = output.WriteString(f); err != nil {
				return "", err
			}
		}
	}
	return output.String(), nil
}

func argumentValue(ph ASTNode, args []interface{}, cursor int) (arg interface{}, nextCursor int, err error) {
	if ph.Keys != nil { // keyword argument

		if cursor < 0 || cursor >= len(args) {
			return nil, cursor, fmt.Errorf("[sprintf] Implicit argument index is out of range. Not enough arguments, need at least %d", cursor+1)
		}

		arg = args[cursor]

		for _, key := range ph.Keys {
			if arg == nil {
				return nil, cursor, fmt.Errorf("[sprintf] Cannot access property %q of nil in %q", key, strings.Join(ph.Keys, "."))
			}
			marg, ok := arg.(map[string]interface{})
			if !ok {
				return nil, cursor, fmt.Errorf("[sprintf] Cannot access property %q in value of type %T", key, arg)
			}
			arg = marg[key]
		}

		return arg, cursor, nil
	}

	if ph.ParamNo != 0 { // positional argument (explicit)
		if ph.ParamNo < 1 || ph.ParamNo > len(args) {
			return nil, cursor, fmt.Errorf("[sprintf] Positional argument index %d is out of range", ph.ParamNo)
		}
		return args[ph.ParamNo-1], cursor, nil
	}

	// positional argument (implicit)
	if cursor < 0 || cursor >= len(args) {
		return nil, cursor, fmt.Errorf("[sprintf] Implicit argument index is out of range. Not enough arguments, need at least %d", cursor+1)
	}
	return args[cursor], cursor + 1, nil
}

func formatPlaceholder(ph ASTNode, value interface{}) (formatted string, err error) {
	if reNotType.MatchString(ph.Type) && reNotPrimitive.MatchString(ph.Type) && isFunc(value) {
		value = reflect.ValueOf(value).Call([]reflect.Value{})
	}

	numberValue := NewNumber(value)
	if reNumericArg.MatchString(ph.Type) && numberValue.IsNaN() {
		return "", fmt.Errorf("[sprintf] expecting number but found %T", value)
	}

	formattedValue := ""
	switch ph.Type[0] {
	case 'c':
		formattedValue = fmt.Sprintf("%c", value)
	case 'b', 'd', 'i', 'u', 'e', 'f', 'g', 'o', 'x', 'X':
		formattedValue, err = formatWithPrecision(ph.Type, ph.Precision, numberValue)
	case 'j':
		formattedValue, err = formatJSON(value, ph.Width)
		if err == nil {
			return formattedValue, nil // bail out early. we do not want signs or padding on JSON
		}
	case 's':
		formattedValue, err = formatWithPrecision(ph.Type, ph.Precision, value)
	case 't':
		formattedValue, err = formatWithPrecision(ph.Type, ph.Precision, coerceBoolean(value))
	case 'T':
		formattedValue = typeName(value)
	case 'v':
		formattedValue, err = formatWithPrecision(ph.Type, ph.Precision, value)
	default:
		formattedValue = fmt.Sprint(value)
	}

	if err != nil {
		return "", fmt.Errorf("[sprintf] failed to format value %v as %q: %v", value, ph.Placeholder, err)
	}

	signChar := ""
	if reNumber.MatchString(ph.Type) {
		if positive := numberValue.IsPositive(); !positive || ph.Sign != "" {
			signChar = sign(positive)
			formattedValue = reSign.ReplaceAllString(formattedValue, "") // remove sign
		}
	}

	return alignedPad(formattedValue, ph.Width, ph.Pad, ph.Align, signChar), nil
}

func formatWithPrecision(typ, precision string, value interface{}) (string, error) {
	if precision == "" {
		return fmt.Sprintf("%"+typ, value), nil
	}

	if typ == "t" {
		// go does not support precision for "%t"
		width, err := strconv.Atoi(precision)
		if err != nil {
			return "", fmt.Errorf("[sprintf] failed to parse precision %q: %v", precision, err)
		}
		return trim(fmt.Sprint(value), width), nil
	}

	return fmt.Sprintf("%."+precision+typ, value), nil
}

func formatJSON(value interface{}, indent int) (string, error) {
	var js []byte
	var err error
	if indent > 0 {
		js, err = json.MarshalIndent(value, "", strings.Repeat(" ", indent))
	} else {
		js, err = json.Marshal(value)
	}
	return string(js), err
}

func alignedPad(value string, width int, padChar string, align string, sign string) string {

	if padChar == "" {
		padChar = " "
	} else if len(padChar) > 1 {
		padChar = padChar[1:]
	}

	padLen := width - len(sign) - len(value)

	pad := ""
	if width > 0 && padLen > 0 {
		pad = strings.Repeat(padChar, padLen)
	}

	if align == "-" {
		return sign + value + pad // e.g. "-3     "
	}

	if padChar == "0" {
		return sign + pad + value // e.g. "-000003"
	}

	return pad + sign + value // e.g. "     -3"
}

func trim(value string, width int) string {
	if width < 0 || width >= len(value) {
		return value
	}
	return value[:width]
}

func typeName(v interface{}) string {
	if v == nil {
		return "null"
	}

	tv := reflect.TypeOf(v)
	switch tv.Kind() {
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Func:
		return "function"
	}

	switch v.(type) {
	case bool:
		return "boolean"
	case string:
		return "string"
	case *regexp.Regexp:
		return "regexp"
	}

	if !NewNumber(v).IsNaN() {
		return "number"
	}

	return "object"
}

func sign(positive bool) string {
	if positive {
		return "+"
	}
	return "-"
}

func isFunc(v interface{}) bool {
	return reflect.ValueOf(v).Kind() == reflect.Func
}

func coerceBoolean(v interface{}) bool {
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Ptr {
		return !vv.IsNil()
	}

	switch v := v.(type) {
	case int:
		return v != 0
	case int8:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case bool:
		return v
	}
	if s := fmt.Sprint(v); s == "0" || s == "" {
		return false
	}
	return true
}
