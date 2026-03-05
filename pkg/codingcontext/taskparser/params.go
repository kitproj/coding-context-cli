package taskparser

import (
	"errors"
	"fmt"
	"maps"
	"strconv"
	"strings"
)

const (
	// ArgumentsKey is the key used to store positional arguments in Params.
	ArgumentsKey = "ARGUMENTS"
)

var (
	// ErrEmptyKey is returned when a parameter key is empty.
	ErrEmptyKey = errors.New("empty key in parameter")
	// ErrInvalidEscapeSequence is returned when an escape sequence is invalid.
	ErrInvalidEscapeSequence = errors.New("invalid escape sequence")
	// ErrInvalidFormat is returned when the input format is invalid.
	ErrInvalidFormat = errors.New("invalid parameter format: missing '='")
	// ErrMismatchedQuotes is returned when quotes don't match.
	ErrMismatchedQuotes = errors.New("mismatched quote types")
	// ErrUnclosedQuote is returned when a quoted string is not properly closed.
	ErrUnclosedQuote = errors.New("unclosed quote")
)

// Params is a map of string keys to string slice values with
// convenience methods for accessing single and multiple values.
type Params map[string][]string

// ParseParams parses a parameter string into a Params map that supports both
// named parameters (key-value pairs) and positional arguments.
//
// The function provides a flexible, permissive parser that handles various
// quoting styles, escape sequences, and separators.
//
// Named Parameters:
//
//	Basic syntax: key=value
//	Multiple pairs can be separated by commas, spaces, or both
//	Whitespace around the = sign is optional
//	Keys are case-insensitive (converted to lowercase)
//	The same key can appear multiple times; all values are collected
//
//	Examples:
//	  "key=value"
//	  "key=value,foo=bar"
//	  "key = value, foo = bar"
//	  "key=value1 key=value2 key=value3"  // Multiple values for same key
//
// Positional Arguments:
//
//	Values without a key are treated as positional arguments and stored
//	under the ArgumentsKey constant ("ARGUMENTS").
//	Positional and named parameters can be interleaved.
//
//	Examples:
//	  "value"
//	  "value1 value2 value3"
//	  "value1, value2, value3"
//	  "key=value positional1 positional2"
//	  "positional1 key=value positional2"
//
// Quoted Values:
//
//	Both single and double quotes are supported for values containing
//	special characters. Quotes can be escaped within matching quote types.
//	Empty quoted values create a value with an empty string.
//
//	Examples:
//	  `key="string value"`
//	  `key='string value'`
//	  `key="value=with=equals"`
//	  `key="value,with,commas"`
//	  `key="bar\"baz\""`  // Escaped quotes
//
// Unquoted Values:
//
//	Unquoted values cannot contain spaces (spaces separate arguments).
//	Values containing =, ,, or spaces should be quoted.
//	Trailing whitespace is trimmed from unquoted values.
//
// Escape Sequences:
//
//	Escape sequences work in both quoted and unquoted contexts:
//	  Standard: \n (newline), \t (tab), \r (carriage return), \\
//	            (backslash), \" (double quote), \' (single quote)
//	  Numeric: \xHH (hex), \uHHHH (Unicode), \OOO (octal, 1-3 digits)
//	  Other: Any other escape returns the character after backslash
//
//	Examples:
//	  `key="line1\nline2\ttabbed"`
//	  `key="\x41\x42"`  // "AB"
//	  `key="\u00a0"`    // Non-breaking space
//
// Separators:
//
//	Multiple separators are supported: commas, spaces, or both.
//	Trailing separators are ignored.
//
//	Examples:
//	  "key=value,foo=bar"
//	  "key=value foo=bar"
//	  "key=value, foo=bar, baz=qux"
//
// Empty Values:
//
//	Unquoted empty: key= creates an empty slice []string{}
//	Quoted empty: key="" or key='' creates []string{""}
//
// Unicode Support:
//
//	Full Unicode and UTF-8 support for keys and values.
//	Unicode whitespace is recognized as separators.
//	All unicode whitespace is automatically trimmed from start/end of values.
//
//	Examples:
//	  "ключ=значение"
//	  "key=こんにちは"
//	  "emoji=🚀"
//
// Error Conditions:
//
//	Returns errors for:
//	  - Unclosed quotes: key="unclosed
//	  - Empty keys: =value
//	  - Invalid escape sequences: incomplete or invalid hex/unicode escapes
//	  - Mismatched quotes: key='value" or key="value'
//
// Return Value:
//
//	The returned Params map has:
//	  - Named parameters: lowercase keys with string slice values
//	  - Positional arguments: stored under ArgumentsKey ("ARGUMENTS")
//
//	Example:
//	  params, _ := ParseParams("key=value1 key=value2 arg1 arg2")
//	  // params["key"] = []string{"value1", "value2"}
//	  // params[ArgumentsKey] = []string{"arg1", "arg2"}
//
// See the Params type methods (Value, Values, Arguments, Lookup) for
// convenient access to parsed parameters.
func ParseParams(value string) (Params, error) {
	params := make(Params)

	// Handle empty input
	value = strings.TrimSpace(value)
	if value == "" {
		return params, nil
	}

	// Check for unclosed quotes
	if err := validateQuotes(value); err != nil {
		return nil, err
	}

	// Parse using Participle
	input, err := paramsParser().ParseString("", value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse params: %w", err)
	}

	// Convert parsed structure to Params map
	return convertToParams(input)
}

// convertToParams converts the parsed AST to Params map.
func convertToParams(input *ParamsInput) (Params, error) {
	params := make(Params)

	for _, item := range input.Items {
		// Skip separators
		if item.Separator != nil {
			continue
		}

		// Handle named parameters
		if item.Named != nil {
			if err := addNamedParam(params, item.Named); err != nil {
				return nil, err
			}

			continue
		}

		// Handle positional arguments
		if item.Positional != nil {
			value, _, err := extractValue(item.Positional)
			if err != nil {
				return nil, err
			}

			params[ArgumentsKey] = append(params[ArgumentsKey], value)
		}
	}

	return params, nil
}

// addNamedParam adds a named parameter to the params map.
func addNamedParam(params Params, named *NamedParam) error {
	key := strings.ToLower(named.Key)
	if key == "" {
		return ErrEmptyKey
	}

	if named.Value == nil {
		if params[key] == nil {
			params[key] = []string{}
		}

		return nil
	}

	value, wasQuoted, err := extractValue(named.Value)
	if err != nil {
		return err
	}

	if wasQuoted || value != "" {
		params[key] = append(params[key], value)
	} else if params[key] == nil {
		params[key] = []string{}
	}

	return nil
}

// extractValue extracts the string value from a Value node
// Returns the value, whether it was quoted, and any error.
func extractValue(val *Value) (string, bool, error) {
	raw := val.Raw

	// Check if it's a quoted string (need at least 2 chars for open/close quote)
	const minQuotedLen = 2
	if len(raw) >= minQuotedLen {
		if (raw[0] == '"' && raw[len(raw)-1] == '"') || (raw[0] == '\'' && raw[len(raw)-1] == '\'') {
			// Quoted value - extract content and process escapes
			content := raw[1 : len(raw)-1]

			processed, err := processEscapes(content)
			if err != nil {
				return "", true, err
			}

			return strings.TrimSpace(processed), true, nil
		}
	}

	// Unquoted value - process escapes
	processed, err := processEscapes(raw)
	if err != nil {
		return "", false, err
	}

	return strings.TrimSpace(processed), false, nil
}

// processEscapes processes all escape sequences in a string.
func processEscapes(s string) (string, error) {
	if !strings.Contains(s, "\\") {
		// Fast path: no escapes
		return s, nil
	}

	var result strings.Builder

	result.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			result.WriteByte(s[i])

			continue
		}

		if i+1 >= len(s) {
			// Incomplete escape at end - treat as literal backslash
			result.WriteByte('\\')

			continue
		}

		advance, err := writeEscapeChar(&result, s, i)
		if err != nil {
			return "", err
		}

		i += advance
	}

	return result.String(), nil
}

// writeEscapeChar writes the character for the escape sequence at s[i] to result.
// i points to the backslash; returns the index advance (not counting the backslash itself).
func writeEscapeChar(result *strings.Builder, s string, i int) (int, error) {
	next := s[i+1]

	// simpleEscapes maps single-character escape sequences to their byte values.
	simpleEscapes := map[byte]byte{
		'n':  '\n',
		't':  '\t',
		'r':  '\r',
		'\\': '\\',
		'"':  '"',
		'\'': '\'',
	}

	if b, ok := simpleEscapes[next]; ok {
		result.WriteByte(b)

		return 1, nil
	}

	switch next {
	case 'u':
		return processUnicodeEscape(result, s, i)
	case 'x':
		return processHexEscape(result, s, i)
	case '0', '1', '2', '3', '4', '5', '6', '7':
		return processOctalEscape(result, s, i)
	default:
		// Any other escape - use the character after backslash literally
		result.WriteByte(next)

		return 1, nil
	}
}

// processUnicodeEscape handles \uXXXX escape sequences.
// Returns the index advance past the full escape sequence.
func processUnicodeEscape(result *strings.Builder, s string, i int) (int, error) {
	const unicodeEscapeLen = 6 // \uXXXX: backslash + 'u' + 4 hex digits

	const unicodeAdvance = unicodeEscapeLen - 1 // advance past 'u' + 4 hex digits

	if i+unicodeEscapeLen > len(s) {
		return 0, fmt.Errorf("%w: incomplete \\u escape", ErrInvalidEscapeSequence)
	}

	hex := s[i+2 : i+6]

	val, err := strconv.ParseUint(hex, 16, 16)
	if err != nil {
		return 0, fmt.Errorf("%w: \\u%s", ErrInvalidEscapeSequence, hex)
	}

	result.WriteRune(rune(val))

	return unicodeAdvance, nil
}

// processHexEscape handles \xHH escape sequences.
// Returns the index advance past the full escape sequence.
func processHexEscape(result *strings.Builder, s string, i int) (int, error) {
	const hexEscapeLen = 4 // \xHH: backslash + 'x' + 2 hex digits

	const hexAdvance = hexEscapeLen - 1 // advance past 'x' + 2 hex digits

	if i+hexEscapeLen > len(s) {
		return 0, fmt.Errorf("%w: incomplete \\x escape", ErrInvalidEscapeSequence)
	}

	hex := s[i+2 : i+4]

	val, err := strconv.ParseUint(hex, 16, 8)
	if err != nil {
		return 0, fmt.Errorf("%w: \\x%s", ErrInvalidEscapeSequence, hex)
	}

	result.WriteByte(byte(val))

	return hexAdvance, nil
}

// processOctalEscape handles \OOO escape sequences (1-3 octal digits).
// Returns the index advance past the full escape sequence.
func processOctalEscape(result *strings.Builder, s string, i int) (int, error) {
	const octalStartOffset = 2

	const maxOctalDigits = 3

	end := i + octalStartOffset
	for end < len(s) && end < i+1+maxOctalDigits && s[end] >= '0' && s[end] <= '7' {
		end++
	}

	octal := s[i+1 : end]

	val, err := strconv.ParseUint(octal, 8, 8)
	if err != nil {
		return 0, fmt.Errorf("%w: \\%s", ErrInvalidEscapeSequence, octal)
	}

	result.WriteByte(byte(val))

	return end - i - 1, nil
}

// validateQuotes checks if all quoted strings in the input are properly closed.
func validateQuotes(input string) error {
	inDoubleQuote := false
	inSingleQuote := false
	escapeNext := false

	for _, r := range input {
		if escapeNext {
			escapeNext = false

			continue
		}

		if r == '\\' {
			escapeNext = true

			continue
		}

		if r == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
		} else if r == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
		}
	}

	if inDoubleQuote || inSingleQuote {
		return ErrUnclosedQuote
	}

	return nil
}

// Set parses and merges key=value parameters from the given string into the Params.
func (p Params) Set(value string) error {
	// Auto-quote values that need quoting for better CLI UX
	quotedValue := autoQuoteParamValue(value)

	params, err := ParseParams(quotedValue)
	if err != nil {
		return fmt.Errorf("failed to parse parameter '%s': %w", value, err)
	}

	maps.Copy(p, params)

	return nil
}

// autoQuoteParamValue automatically quotes the value part of a key=value parameter
// if it contains characters that require quoting (spaces, commas, equals, etc.)
// and is not already quoted.
func autoQuoteParamValue(input string) string {
	equalsIndex := strings.IndexByte(input, '=')
	if equalsIndex == -1 || equalsIndex == len(input)-1 {
		return input // No = sign or empty value
	}

	value := strings.TrimSpace(input[equalsIndex+1:])
	if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
		(value[0] == '\'' && value[len(value)-1] == '\'')) {
		return input // Already quoted
	}

	if needsQuoting(value) {
		return input[:equalsIndex+1] + strconv.Quote(value)
	}

	return input
}

// charsRequiringQuote is the set of runes that require a value to be quoted.
const charsRequiringQuote = " \t\n\r,=\"'\u00a0\u1680\u2000\u2001\u2002\u2003\u2004\u2005" +
	"\u2006\u2007\u2008\u2009\u200a\u202f\u205f\u3000"

// needsQuoting checks if a value contains characters that require quoting.
func needsQuoting(value string) bool {
	if value == "" {
		return false
	}

	for _, r := range value {
		if strings.ContainsRune(charsRequiringQuote, r) {
			return true
		}
	}

	return false
}

func (p Params) String() string {
	pairs := make([]string, 0, len(p))
	for key, values := range p {
		for _, value := range values {
			pairs = append(pairs, key+"="+value)
		}
	}

	return strings.Join(pairs, ",")
}

// Value returns the first value for the given key, or an empty string if
// the key is not found or has no values.
func (p Params) Value(key string) string {
	if p == nil {
		return ""
	}

	values := p[strings.ToLower(key)]
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

// Lookup returns the first value for the given key, or empty string and false if not found.
func (p Params) Lookup(key string) (string, bool) {
	if p == nil {
		return "", false
	}

	values := p[strings.ToLower(key)]
	if len(values) == 0 {
		return "", false
	}

	return values[0], true
}

// Values returns all values for the given key, or an empty slice if
// the key is not found.
func (p Params) Values(key string) []string {
	if p == nil {
		return nil
	}

	return p[strings.ToLower(key)]
}

// Arguments returns all positional arguments (values without keys), or an empty slice
// if there are no positional arguments. This is distinct from named parameters
// accessed via Value() or Values().
func (p Params) Arguments() []string {
	if p == nil {
		return nil
	}

	return p[ArgumentsKey]
}
