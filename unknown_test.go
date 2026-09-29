package flags

import (
	"testing"
)

func TestUnknownFlags(t *testing.T) {
	var opts = struct {
		Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
	}{}

	args := []string{
		"-f",
	}

	p := NewParser(&opts, 0)
	args, err := p.ParseArgs(args)

	if err == nil {
		t.Fatal("Expected error for unknown argument")
	}
}

func TestIgnoreUnknownFlags(t *testing.T) {
	var opts = struct {
		Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
	}{}

	args := []string{
		"hello",
		"world",
		"-v",
		"--foo=bar",
		"--verbose",
		"-f",
	}

	p := NewParser(&opts, IgnoreUnknown)
	args, err := p.ParseArgs(args)

	if err != nil {
		t.Fatal(err)
	}

	exargs := []string{
		"hello",
		"world",
		"--foo=bar",
		"-f",
	}

	issame := (len(args) == len(exargs))

	if issame {
		for i := 0; i < len(args); i++ {
			if args[i] != exargs[i] {
				issame = false
				break
			}
		}
	}

	if !issame {
		t.Fatalf("Expected %v but got %v", exargs, args)
	}
}

// Regression test for issue #399.
func TestIgnoreUnknownStackedShort(t *testing.T) {
	for _, test := range []struct {
		arg   string
		wantA []bool
	}{
		{"-axb", []bool{true}},
		{"-xb", nil},
		{"-xyab", []bool{true}},
		{"-axyb", []bool{true}},
		{"-xb=ignored", nil},
	} {
		t.Run(test.arg, func(t *testing.T) {
			var opts struct {
				A []bool `short:"a"`
				B []bool `short:"b"`
			}

			p := NewParser(&opts, IgnoreUnknown)
			args, err := p.ParseArgs([]string{test.arg})
			if err != nil {
				t.Fatal(err)
			}

			assertBoolArray(t, opts.A, test.wantA)
			assertBoolArray(t, opts.B, []bool{true})
			assertStringArray(t, args, []string{test.arg})
		})
	}
}

// Without IgnoreUnknown an unknown short in a stacked cluster must still
// produce an ErrUnknownFlag (control case for the contract change).
func TestStackedShortUnknownFlagError(t *testing.T) {
	var opts = struct {
		A bool `short:"a" long:"alpha"`
		B bool `short:"b" long:"beta"`
	}{}

	p := NewParser(&opts, 0)
	_, err := p.ParseArgs([]string{"-axb"})

	if err == nil {
		t.Fatal("Expected an error for the unknown short without IgnoreUnknown")
	}

	flagsErr, ok := err.(*Error)
	if !ok || flagsErr.Type != ErrUnknownFlag {
		t.Fatalf("Expected ErrUnknownFlag but got %v", err)
	}

	if !opts.A || opts.B {
		t.Fatalf("Expected parsing to stop at the unknown short, got A=%v B=%v", opts.A, opts.B)
	}
}

func TestIgnoreUnknownStackedShortParsers(t *testing.T) {
	for _, arg := range []string{"-Qq", "-qQ"} {
		t.Run(arg, func(t *testing.T) {
			var upper struct {
				Quiet bool `short:"Q"`
			}
			var lower struct {
				Quiet bool `short:"q"`
			}

			args, err := NewParser(&upper, IgnoreUnknown).ParseArgs([]string{arg})
			if err != nil {
				t.Fatal(err)
			}
			assertStringArray(t, args, []string{arg})

			args, err = NewParser(&lower, IgnoreUnknown).ParseArgs(args)
			if err != nil {
				t.Fatal(err)
			}
			assertStringArray(t, args, []string{arg})

			if !upper.Quiet || !lower.Quiet {
				t.Fatalf("Expected both parsers to recognize their short option, got Q=%v q=%v", upper.Quiet, lower.Quiet)
			}
		})
	}
}

func TestIgnoreUnknownStackedShortArgument(t *testing.T) {
	var opts struct {
		Value int `short:"值"`
	}

	args, err := NewParser(&opts, IgnoreUnknown).ParseArgs([]string{"-x值", "42", "tail"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Value != 42 {
		t.Fatalf("Expected the last short option to consume its argument, got %v", opts.Value)
	}
	assertStringArray(t, args, []string{"-x值", "tail"})
}

func TestIgnoreUnknownStackedShortOptionalArgument(t *testing.T) {
	var opts struct {
		Value string `short:"v" optional:"yes" optional-value:"fallback"`
	}

	args, err := NewParser(&opts, IgnoreUnknown).ParseArgs([]string{"-xv", "value"})
	if err != nil {
		t.Fatal(err)
	}
	assertString(t, opts.Value, "fallback")
	assertStringArray(t, args, []string{"-xv", "value"})
}

func TestIgnoreUnknownStackedShortError(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		typ  ErrorType
	}{
		{"missing argument", []string{"-xv"}, ErrExpectedArgument},
		{"invalid argument", []string{"-xv", "bad"}, ErrMarshal},
		{"non-final argument", []string{"-xvb", "42"}, ErrExpectedArgument},
		{"cleared argument", []string{"-xv=42"}, ErrExpectedArgument},
		{"help", []string{"-xhb"}, ErrHelp},
	} {
		t.Run(test.name, func(t *testing.T) {
			var opts struct {
				Value int  `short:"v"`
				B     bool `short:"b"`
			}

			_, err := NewParser(&opts, IgnoreUnknown|HelpFlag).ParseArgs(test.args)
			flagsErr, ok := err.(*Error)
			if !ok || flagsErr.Type != test.typ {
				t.Fatalf("Expected %v instead of an ignored unknown flag, got %v", test.typ, err)
			}
			if opts.B {
				t.Fatal("Expected parsing to stop at the option error")
			}
		})
	}
}

func TestStackedShortUnknownFlagHandler(t *testing.T) {
	for _, test := range []struct {
		name    string
		options Options
		calls   int
		wantB   bool
		args    []string
	}{
		{"handler", None, 1, false, []string{"tail"}},
		{"ignore unknown", IgnoreUnknown, 0, true, []string{"-axby", "tail"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var opts struct {
				A bool `short:"a"`
				B bool `short:"b"`
			}

			p := NewParser(&opts, test.options)
			calls := 0
			p.UnknownOptionHandler = func(option string, arg SplitArgument, args []string) ([]string, error) {
				calls++
				assertString(t, option, "axby")
				return args, nil
			}

			args, err := p.ParseArgs([]string{"-axby", "tail"})
			if err != nil {
				t.Fatal(err)
			}
			if calls != test.calls {
				t.Fatalf("Expected %v handler calls, got %v", test.calls, calls)
			}
			if !opts.A || opts.B != test.wantB {
				t.Fatalf("Unexpected parsed options A=%v B=%v", opts.A, opts.B)
			}
			assertStringArray(t, args, test.args)
		})
	}
}
