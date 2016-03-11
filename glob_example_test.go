package glob

import "fmt"

func ExampleNewPattern() {
	var p *GlobPattern
	var err error

	p, err = NewPattern("foo/bar/*/baz")
	if err != nil {
		panic(err)
	}

	m := p.Matches("foo/bar/quz/baz")
	if m {
		fmt.Println("Matches!")
	}

	m = p.Matches("foo/bar/")
	if !m {
		fmt.Println("Doesn't match!")
	}

	// Output:
	// Matches!
	// Doesn't match!
}

func ExampleMatches() {
	var m bool
	var err error

	// Match
	m, err = Matches(PatternStr("foo/bar/*/baz"), "foo/bar/qux/baz")

	switch {
	case err != nil:
		panic(err)
	case m:
		fmt.Println("Matches!")
	}

	// No match
	m, err = Matches(PatternStr("foo/bar/*/baz?"), "foo/bar/qux/baz")

	switch {
	case err != nil:
		panic(err)
	case !m:
		fmt.Println("Doesn't match!")
	}

	// Output:
	// Matches!
	// Doesn't match!
}
