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

func TestIgnoreUnknownSingleOptions(t *testing.T) {
	for _, arg := range []string{"-x", "-x=ignored", "-界", "-\xff", "--unknown=ignored"} {
		t.Run(arg, func(t *testing.T) {
			var opts struct {
				Verbose bool `short:"v"`
			}

			args, err := NewParser(&opts, IgnoreUnknown).ParseArgs([]string{arg, "-v", "tail"})
			if err != nil {
				t.Fatal(err)
			}
			if !opts.Verbose {
				t.Fatal("Expected the known option to be parsed")
			}
			assertStringArray(t, args, []string{arg, "tail"})
		})
	}
}

// Regression test for issue #399.
func TestIgnoreUnknownStackedShort(t *testing.T) {
	for _, test := range []struct {
		arg   string
		wantA []bool
		wantB []bool
	}{
		{"-axb", []bool{true}, []bool{true}},
		{"-xb", nil, []bool{true}},
		{"-xyab", []bool{true}, []bool{true}},
		{"-axyb", []bool{true}, []bool{true}},
		{"-xb=ignored", nil, []bool{true}},
		{"-xy", nil, nil},
		{"-界x", nil, nil},
	} {
		t.Run(test.arg, func(t *testing.T) {
			var opts struct {
				A []bool `short:"a"`
				B []bool `short:"b"`
			}

			p := NewParser(&opts, IgnoreUnknown)
			args, err := p.ParseArgs([]string{test.arg, "tail"})
			if err != nil {
				t.Fatal(err)
			}

			assertBoolArray(t, opts.A, test.wantA)
			assertBoolArray(t, opts.B, test.wantB)
			assertStringArray(t, args, []string{"tail"})
		})
	}
}

// Without IgnoreUnknown an unknown short in a stacked cluster must still
// produce an ErrUnknownFlag (control case for the contract change).
func TestStackedShortUnknownFlagError(t *testing.T) {
	for _, test := range []struct {
		arg   string
		wantA bool
	}{
		{"-axb", true},
		{"-xb=ignored", false},
		{"-xy", false},
	} {
		t.Run(test.arg, func(t *testing.T) {
			var opts struct {
				A bool `short:"a" long:"alpha"`
				B bool `short:"b" long:"beta"`
			}

			_, err := NewParser(&opts, None).ParseArgs([]string{test.arg})
			flagsErr, ok := err.(*Error)
			if !ok || flagsErr.Type != ErrUnknownFlag {
				t.Fatalf("Expected ErrUnknownFlag but got %v", err)
			}
			if opts.A != test.wantA || opts.B {
				t.Fatalf("Expected parsing to stop at the unknown short, got A=%v B=%v", opts.A, opts.B)
			}
		})
	}
}

func TestIgnoreUnknownStackedShortPositional(t *testing.T) {
	var opts struct {
		A          bool `short:"a"`
		B          bool `short:"b"`
		Positional struct {
			Value int
		} `positional-args:"yes"`
	}

	args, err := NewParser(&opts, IgnoreUnknown).ParseArgs([]string{"-axb", "42", "tail"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.A || !opts.B || opts.Positional.Value != 42 {
		t.Fatalf("Unexpected parsed options A=%v B=%v positional=%v", opts.A, opts.B, opts.Positional.Value)
	}
	assertStringArray(t, args, []string{"tail"})
}

func TestIgnoreUnknownShortConcatArgument(t *testing.T) {
	for _, arg := range []string{"-vxb", "-v=xb"} {
		t.Run(arg, func(t *testing.T) {
			var opts struct {
				Value string `short:"v"`
				B     bool   `short:"b"`
			}

			args, err := NewParser(&opts, IgnoreUnknown).ParseArgs([]string{arg, "tail"})
			if err != nil {
				t.Fatal(err)
			}
			assertString(t, opts.Value, "xb")
			if opts.B {
				t.Fatal("Expected the attached value not to be parsed as short options")
			}
			assertStringArray(t, args, []string{"tail"})
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
	assertStringArray(t, args, []string{"tail"})
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
	assertStringArray(t, args, []string{"value"})
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
		{"bool argument", []string{"-b=42"}, ErrNoArgumentForBool},
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
		{"ignore unknown", IgnoreUnknown, 0, true, []string{"tail"}},
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
