package flags

import (
	"testing"
)

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		s    string
		t    string
		dist int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "abc", 0},
		{"a", "abcdef", 5},
		{"abcdef", "a", 5},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"flaw", "lawn", 2},
		{"rmive", "remove", 2},

		// Multi-byte characters count as a single edit each
		{"héllo", "hello", 1},
		{"日本語", "日本", 1},
		{"日本語", "日本語", 0},
		{"ünïcode", "unicode", 2},
		{"日本語", "", 3},
	}

	for _, test := range tests {
		if dist := levenshtein(test.s, test.t); dist != test.dist {
			assertErrorf(t, "Expected levenshtein(%q, %q) to be %d, but got %d", test.s, test.t, test.dist, dist)
		}
	}
}

func TestLevenshteinSymmetric(t *testing.T) {
	words := []string{"", "a", "add", "addd", "remove", "rmive", "日本語", "héllo"}

	for _, a := range words {
		for _, b := range words {
			if levenshtein(a, b) != levenshtein(b, a) {
				assertErrorf(t, "levenshtein is not symmetric for %q and %q: %d != %d",
					a, b, levenshtein(a, b), levenshtein(b, a))
			}
		}
	}
}

func TestClosestChoice(t *testing.T) {
	if c, _ := closestChoice("rmive", []string{"add", "remove", "list"}); c != "remove" {
		assertErrorf(t, "Expected `remove', but got `%s'", c)
	}

	if c, l := closestChoice("x", nil); c != "" || l != 0 {
		assertErrorf(t, "Expected empty choice, but got `%s' (%d)", c, l)
	}
}

// A single character input is far away from every long command name, so it
// should not produce a `did you mean' suggestion.
func TestUnknownCommandNoBogusSuggestion(t *testing.T) {
	var opts struct{}

	p := NewNamedParser("test", None)
	p.AddGroup("Application Options", "", &opts)
	p.AddCommand("autocomplete-install", "", "", &struct{}{})
	p.AddCommand("branch", "", "", &struct{}{})

	_, err := p.ParseArgs([]string{"q"})

	assertError(t, err, ErrUnknownCommand,
		"Unknown command `q'. Please specify one command of: autocomplete-install or branch")
}
