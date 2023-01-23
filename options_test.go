package flags

import (
	"reflect"
	"strings"
	"testing"
)

func TestPassDoubleDash(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	p := NewParser(&opts, PassDoubleDash)
	ret, err := p.ParseArgs([]string{"-v", "--", "-v", "-g"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assertStringArray(t, ret, []string{"-v", "-g"})
}

func TestPassAfterNonOption(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "arg", "-v", "-g"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assertStringArray(t, ret, []string{"arg", "-v", "-g"})
}

func TestPassAfterNonOptionWithPositional(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []string `required:"yes"`
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "arg", "-v", "-g"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assertStringArray(t, ret, []string{})
	assertStringArray(t, opts.Positional.Rest, []string{"arg", "-v", "-g"})
}

func TestPassAfterNonOptionWithPositionalIntPass(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []int `required:"yes"`
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "1", "2", "3"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assertStringArray(t, ret, []string{})
	for i, rest := range opts.Positional.Rest {
		if rest != i+1 {
			assertErrorf(t, "Expected %v got %v", i+1, rest)
		}
	}
}

func TestPassAfterNonOptionWithPositionalIntFail(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []int `required:"yes"`
		} `positional-args:"yes"`
	}{}

	tests := []struct {
		opts        []string
		errContains string
		ret         []string
	}{
		{
			[]string{"-v", "notint1", "notint2", "notint3"},
			"notint1",
			[]string{"notint1", "notint2", "notint3"},
		},
		{
			[]string{"-v", "1", "notint2", "notint3"},
			"notint2",
			[]string{"1", "notint2", "notint3"},
		},
	}

	for _, test := range tests {
		p := NewParser(&opts, PassAfterNonOption)
		ret, err := p.ParseArgs(test.opts)

		if err == nil {
			assertErrorf(t, "Expected error")
			return
		}

		if !strings.Contains(err.Error(), test.errContains) {
			assertErrorf(t, "Expected the first illegal argument in the error")
		}

		assertStringArray(t, ret, test.ret)
	}
}

func TestTerminatedOptions(t *testing.T) {
	type testOpt struct {
		Slice         []int      `short:"s" long:"slice" terminator:"END"`
		MultipleSlice [][]string `short:"m" long:"multiple" terminator:";"`
		Bool          bool       `short:"v"`
	}

	tests := []struct {
		summary               string
		parserOpts            Options
		args                  []string
		expectedSlice         []int
		expectedMultipleSlice [][]string
		expectedBool          bool
		expectedRest          []string
		shouldErr             bool
	}{
		{
			summary: "Terminators usage",
			args: []string{
				"-s", "1", "2", "3", "END",
				"-m", "bin", "-xyz", "--foo", "bar", "-v", "foo bar", ";",
				"-v",
				"-m", "-xyz", "--foo",
			},
			expectedSlice: []int{1, 2, 3},
			expectedMultipleSlice: [][]string{
				{"bin", "-xyz", "--foo", "bar", "-v", "foo bar"},
				{"-xyz", "--foo"},
			},
			expectedBool: true,
		}, {
			summary: "Slice overwritten",
			args: []string{
				"-s", "1", "2", "END",
				"-s", "3", "4",
			},
			expectedSlice: []int{3, 4},
		}, {
			summary: "Terminator omitted for last opt",
			args: []string{
				"-s", "1", "2", "3",
			},
			expectedSlice: []int{1, 2, 3},
		}, {
			summary: "Shortnames jumbled",
			args: []string{
				"-vm", "--foo", "-v", "bar", ";",
				"-s", "1", "2",
			},
			expectedSlice:         []int{1, 2},
			expectedMultipleSlice: [][]string{{"--foo", "-v", "bar"}},
			expectedBool:          true,
		}, {
			summary: "Terminator as a token",
			args: []string{
				"-m", "--foo", "-v;", "-v",
			},
			expectedMultipleSlice: [][]string{{"--foo", "-v;", "-v"}},
		}, {
			summary:    "DoubleDash",
			parserOpts: PassDoubleDash,
			args: []string{
				"-m", "--foo", "--", "bar", ";",
				"-v",
				"--", "--foo", "bar",
			},
			expectedMultipleSlice: [][]string{{"--foo", "--", "bar"}},
			expectedBool:          true,
			expectedRest:          []string{"--foo", "bar"},
		}, {
			summary:   "--opt=foo syntax",
			args:      []string{"-m=foo", "bar"},
			shouldErr: true,
		}, {
			summary:   "--opt= syntax",
			args:      []string{"-m=", "foo", "bar"},
			shouldErr: true,
		}, {
			summary:               "No args",
			args:                  []string{"-m", ";", "-s", "END"},
			expectedMultipleSlice: [][]string{{}},
		}, {
			summary:               "No args, no terminator",
			args:                  []string{"-m"},
			expectedMultipleSlice: [][]string{{}},
		}, {
			summary: "Missing args in the middle",
			args: []string{
				"-m", "a", ";",
				"-m", ";",
				"-m", "b",
			},
			expectedMultipleSlice: [][]string{{"a"}, {}, {"b"}},
		}, {
			summary:               "Nil args",
			args:                  []string{"-m", ""},
			expectedMultipleSlice: [][]string{{""}},
		},
	}

	for _, test := range tests {
		t.Run(test.summary, func(t *testing.T) {
			opts := testOpt{}
			p := NewParser(&opts, test.parserOpts)
			rest, err := p.ParseArgs(test.args)

			if err != nil {
				if !test.shouldErr {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}
			if test.shouldErr {
				t.Errorf("Expected error")
			}

			if opts.Bool != test.expectedBool {
				t.Errorf("Expected Bool to be %v, got %v", test.expectedBool, opts.Bool)
			}

			if !reflect.DeepEqual(opts.Slice, test.expectedSlice) {
				t.Errorf("Expected Slice to be %v, got %v", test.expectedSlice, opts.Slice)
			}

			if !reflect.DeepEqual(opts.MultipleSlice, test.expectedMultipleSlice) {
				t.Log(reflect.DeepEqual(opts.MultipleSlice, [][]string{nil}))
				t.Errorf("Expected MultipleSlice to be %v, got %v", test.expectedMultipleSlice, opts.MultipleSlice)
			}

			assertStringArray(t, rest, test.expectedRest)
		})
	}
}
