package flags_test

import (
	"github.com/jessevdk/go-flags"
	"testing"
)

func TestPassDoubleDash(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	p := flags.NewParser(&opts, flags.PassDoubleDash)
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

	p := flags.NewParser(&opts, flags.PassAfterNonOption)
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
