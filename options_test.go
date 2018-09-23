package flags

import (
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
