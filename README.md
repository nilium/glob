# glob

The glob package is a simple package for doing simple pattern matching using wildcards. It exports three functions and one type:

- `type glob.GlobPattern`
  
    A compiled glob pattern.

    - `func (p *GlobPattern) Matches(str string) bool`

        Returns whether the pattern matches the string str. Does not return an
        error.

- `func glob.NewPattern(pattern string) *GlobPattern, err`

    Attempts to compile the pattern string into a GlobPattern. If successful,
    return the GlobPattern and nil, otherwise returns nil and an error.

- `func glob.Matches(pattern any, str string) bool, err`

    Returns whether or not the given pattern, which may be a `*GlobPattern`
    or string (if a string, a new GlobPattern will be compiled as needed),
    matches the given string str. May return an error.

    This is a general purpose function that can be used if you don't plan to
    reuse a pattern many times or can't guarantee you'll always use the same
    pattern.


## License

The glob package is distributed under the Boost Software License, Version 1.0.
See accompanying file LICENSE_1_0.txt or copy at
<http://www.boost.org/LICENSE_1_0.txt>.
