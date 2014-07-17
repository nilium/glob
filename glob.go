//
// Copyright Noel Cower 2014.
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
//  http://www.boost.org/LICENSE_1_0.txt)
//

// Package glob provides rudimentary pattern matching functions using
// shell-like wildcards `*` and `?`.
package glob

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// GlobPattern is a compiled glob pattern.
type GlobPattern struct {
	pattern string
	steps   []*globScanner
}

// scanFunc implementations attempt to match something followed by a given
// substring that may be empty. If the match is successful, they return true,
// a slice of the input string sans the matched bytes, and the number of bytes
// consumed by the match. If the match fails, they must return false, the input
// string, and 0.
type scanFunc func(input, substr string) (bool, string, int)
type globKind int

const (
	globMany   globKind = iota
	globOne    globKind = iota
	globString globKind = iota
	globEnd    globKind = iota
)

func (k globKind) String() string {
	switch k {
	case globMany:
		return "globMany"
	case globOne:
		return "globOne"
	case globString:
		return "globString"
	case globEnd:
		return "globEnd"
	default:
		return "unknown"
	}
}

// Creates a new GlobPattern based on pattern and returns it. Patterns consist
// of varying sequences of chars interspersed with wildcards -- either `*` or
// `?` to match 1 or more characters or a single character, respectively. Any
// character may be escaped with a backslash (\) to produce the same literal
// character in the string. Escaping any other character will yield the escaped
// character. Avoid escaping characters where possible, as this introduces
// additional complexity into the pattern.
func NewPattern(pattern string) (*GlobPattern, error) {
	steps := compileGlobPattern(pattern)
	if steps == nil {
		return nil, errors.New("Unable to compile glob pattern")
	} else if len(steps) == 0 {
		return nil, errors.New("Compiled glob pattern was empty")
	}
	return &GlobPattern{pattern, steps}, nil
}

// String returns the pattern this GlobPattern was compiled with.
func (p *GlobPattern) String() string {
	return p.pattern
}

// Matches returns whether the glob pattern p matches str.
func (p *GlobPattern) Matches(str string) bool {
	steps := p.steps
	var numSteps = len(steps)
	var stepIndex = 0
	var substr = str
	var matches bool = false
	var bytesConsumed int = 0
	var firstMany int = -1
	var firstManySubstr string = substr
	var wasReset = false

	for stepIndex < numSteps {
		test := steps[stepIndex]

		if (firstMany == -1 || firstMany == stepIndex) && test.kind == globMany {
			firstMany = stepIndex
			firstManySubstr = substr
		}

		matches, substr, bytesConsumed = test.scanner(substr, test.substr)
		if firstMany == stepIndex && matches {
			firstManySubstr = firstManySubstr[bytesConsumed-len(test.substr):]
		}

		if !matches {
			if firstMany == -1 || stepIndex == 0 || wasReset || len(firstManySubstr) == 0 {
				return false
			}

			stepIndex = firstMany
			substr = firstManySubstr[1:]
			wasReset = true
		} else {
			stepIndex++
			wasReset = false
		}

	}

	return len(substr) == 0 && stepIndex == numSteps
}

type globScanner struct {
	scanner scanFunc
	kind    globKind
	substr  string
	start   int
}

// Matches returns whether the glob pattern matches str. If an error occurs
// (i.e., the pattern is somehow invalid), it will always return false and an
// error. Otherwise, if the pattern is valid, it will return true or false
// depending on whether it matches str and a nil error.
//
// pattern may be either a GlobPattern or a string. If it's a string, it will
// be parsed and compiled on demand. Any other type is an error.
func Matches(pattern interface{}, str string) (bool, error) {
	if ps, ok := pattern.(string); ok {
		p, err := NewPattern(ps)
		if err != nil {
			return false, err
		}
		return p.Matches(str), nil
	} else if p, ok := pattern.(GlobPattern); ok {
		return p.Matches(str), nil
	}
	return false, errors.New("Invalid pattern type -- must be string or GlobPattern")
}

// consumeAllPreceding consumes one or more characters in a string up to the
// given substring. If it successfully finds substr in the string, it returns a
// slice of str starting after the found substring. substr may be empty.
// On failure, returns false, str, and 0.
func consumeAllPreceding(str, substr string) (bool, string, int) {
	if len(str) == 0 {
		return false, str, 0
	} else if len(substr) == 0 {
		return true, str[len(str):], len(str)
	}

	offset := 0
	subIndex := strings.Index(str, substr)
	for subIndex != -1 {
		offset += subIndex
		if subIndex > 0 {
			return true, str[offset+len(substr):], offset + len(substr)
		}
		subIndex = strings.Index(str[offset+subIndex+1:], substr)
	}

	return false, str, 0
}

// consumeOnePreceding consumes single code that must be followed by the given
// substring. substr may be empty.
func consumeOnePreceding(str, substr string) (bool, string, int) {
	if len(str) < 1 {
		return false, str, 0
	}

	r := strings.NewReader(str)
	_, size, err := r.ReadRune()

	switch {
	case err != nil:
	case r.Len() < len(substr):
		return false, str, 0
	}

	if err != nil {
		return false, str, 0
	} else if len(str) < len(substr)+1 {
		return false, str, 0
	} else if len(substr) == 0 {
		return true, str[size:], size
	} else if str[size:size+len(substr)] != substr {
		return true, str[size+len(substr):], size + len(substr)
	}

	return false, str, 0
}

// consumeSubstring matches str if it begins with substring. If successful, it
// returns true, str sliced past substr, and len(substr).
func consumeSubstring(str, substr string) (bool, string, int) {
	if len(str) < len(substr) {
		return false, str, 0
	} else if len(substr) == 0 {
		return true, str, 0
	} else if str[:len(substr)] != substr {
		return false, str, 0
	}
	return true, str[len(substr):], len(substr)
}

// consumeEnd consumes only the end of a string. It only matches if len(str) is
// 0 and len(substr) is 0. It will always return str without slicing it.
// The number of bytes it consumes is always 0.
func consumeEnd(str, substr string) (bool, string, int) {
	return len(str) == 0 && len(substr) == 0, str, 0
}

// compileGlobPattern takes a given pattern string consisting of typical
// wildcard characters *, ?, or any literal string and returns a compiled slice
// of scanner functions.
//
// Any character in the pattern string can be escaped using a backslash to
// produce the literal character following it rather than a special character.
func compileGlobPattern(pattern string) []*globScanner {
	// compile scanner function array
	wildcards := make([]*globScanner, 0, 4)
	for index, code := range pattern {
		var fn scanFunc = nil
		var start int = -1
		var kind globKind
		switch {
		case code == '\\':
			fn = consumeSubstring
			kind = globString
		case code == '*':
			fn = consumeAllPreceding
			kind = globMany
		case code == '?':
			fn = consumeOnePreceding
			kind = globOne
		case index == 0:
			fn = consumeSubstring
			start = index
			kind = globString
		default:
			continue
		}

		numWildcards := len(wildcards)
		if numWildcards > 0 {
			last := wildcards[numWildcards-1]
			if code == '\\' && len(last.substr) == 0 {
				last.start += utf8.RuneLen(code)
				continue
			} else {
				last.substr = pattern[last.start:index]
			}
		}

		if start == -1 {
			start = index + utf8.RuneLen(code)
		}

		wildcards = append(wildcards, &globScanner{fn, kind, "", start})
	}

	numWildcards := len(wildcards)
	if numWildcards > 0 {
		last := wildcards[numWildcards-1]
		last.substr = pattern[last.start:]
	}

	wildcards = append(wildcards, &globScanner{consumeEnd, globEnd, "", len(pattern)})

	return wildcards
}
