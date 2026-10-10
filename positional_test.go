package flags

import (
	"os"
	"testing"
)

func TestPositionalEnvSingle(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("TEST_POS", "from_env")

	var opts struct {
		Positional struct {
			Arg string `positional-arg-name:"ARG" env:"TEST_POS"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, Default)
	ret, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertString(t, opts.Positional.Arg, "from_env")
	assertStringArray(t, ret, []string{})
}

func TestPositionalEnvSingleOverride(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("TEST_POS", "from_env")

	var opts struct {
		Positional struct {
			Arg string `positional-arg-name:"ARG" env:"TEST_POS"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, Default)
	ret, err := p.ParseArgs([]string{"from_cli"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertString(t, opts.Positional.Arg, "from_cli")
	assertStringArray(t, ret, []string{})
}

func TestPositionalEnvSlice(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("TEST_SLICE", "alpha,beta,gamma")

	var opts struct {
		Positional struct {
			Rest []string `positional-arg-name:"REST" env:"TEST_SLICE" env-delim:","`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, Default)
	ret, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertStringArray(t, opts.Positional.Rest, []string{"alpha", "beta", "gamma"})
	assertStringArray(t, ret, []string{})
}

func TestPositionalEnvSliceOverride(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("TEST_SLICE", "alpha,beta,gamma")

	var opts struct {
		Positional struct {
			Rest []string `positional-arg-name:"REST" env:"TEST_SLICE" env-delim:","`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, Default)
	ret, err := p.ParseArgs([]string{"cli1", "cli2"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertStringArray(t, opts.Positional.Rest, []string{"cli1", "cli2"})
	assertStringArray(t, ret, []string{})
}

func TestPositionalEnvRequiredPass(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("TEST_REQ", "provided_via_env")

	var opts struct {
		Positional struct {
			Arg string `positional-arg-name:"ARG" env:"TEST_REQ" required:"yes"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatalf("Expected required positional argument to pass with env fallback, got: %v", err)
	}

	assertString(t, opts.Positional.Arg, "provided_via_env")
}

func TestPositionalEnvRequiredFail(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Unsetenv("TEST_REQ")

	var opts struct {
		Positional struct {
			Arg string `positional-arg-name:"ARG" env:"TEST_REQ" required:"yes"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{})
	assertError(t, err, ErrRequired, "the required argument `ARG` was not provided")
}

func TestPositionalIssue408OnPosField(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("FOO", "foo_val")
	os.Setenv("BAR", "a non-empty string")

	type PosArg struct {
		Pos string `env:"BAR" positional-arg-name:"POSITIONAL_ARG"`
	}

	type Args struct {
		Foo string `env:"FOO" short:"f" long:"foo"`
		Bar PosArg `positional-args:"1"`
	}

	var opts Args
	p := NewParser(&opts, Default)
	_, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertString(t, opts.Foo, "foo_val")
	assertString(t, opts.Bar.Pos, "a non-empty string")
}

func TestPositionalIssue408OnParentStruct(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("FOO", "foo_val")
	os.Setenv("BAR", "a non-empty string")

	type PosArg struct {
		Pos string `positional-arg-name:"POSITIONAL_ARG"`
	}

	type Args struct {
		Foo string `env:"FOO" short:"f" long:"foo"`
		Bar PosArg `positional-args:"1" env:"BAR"`
	}

	var opts Args
	p := NewParser(&opts, Default)
	_, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertString(t, opts.Foo, "foo_val")
	assertString(t, opts.Bar.Pos, "a non-empty string")
}

func TestPositionalEnvNamespace(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("APP_POS", "namespaced_val")

	type PosArg struct {
		Pos string `positional-arg-name:"POS" env:"POS"`
	}

	type Args struct {
		Bar PosArg `positional-args:"yes" env-namespace:"APP"`
	}

	var opts Args
	p := NewParser(&opts, Default)
	_, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertString(t, opts.Bar.Pos, "namespaced_val")
}

func TestPositionalEnvMultipleArgs(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("FIRST_ARG", "env_first")
	os.Setenv("SECOND_ARG", "env_second")

	var opts struct {
		Positional struct {
			First  string `positional-arg-name:"FIRST" env:"FIRST_ARG"`
			Second string `positional-arg-name:"SECOND" env:"SECOND_ARG"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, Default)
	_, err := p.ParseArgs([]string{"cli_first"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertString(t, opts.Positional.First, "cli_first")
	assertString(t, opts.Positional.Second, "env_second")
}

func TestPositionalEnvInvalidConversion(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("INT_ARG", "not_a_number")

	var opts struct {
		Positional struct {
			Number int `positional-arg-name:"NUM" env:"INT_ARG"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{})
	if err == nil {
		t.Fatalf("Expected error for invalid int conversion, got nil")
	}
}

func TestPositionalEnvSubcommand(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("SUBCMD_ARG", "sub_from_env")

	type SubCommand struct {
		Positional struct {
			Item string `positional-arg-name:"ITEM" env:"SUBCMD_ARG"`
		} `positional-args:"yes"`
	}

	var opts struct {
		Sub SubCommand `command:"sub"`
	}

	p := NewParser(&opts, Default)
	_, err := p.ParseArgs([]string{"sub"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertString(t, opts.Sub.Positional.Item, "sub_from_env")
}

func TestPositionalEnvSubcommandOverride(t *testing.T) {
	oldEnv := EnvSnapshot()
	defer oldEnv.Restore()

	os.Setenv("SUBCMD_ARG", "sub_from_env")

	type SubCommand struct {
		Positional struct {
			Item string `positional-arg-name:"ITEM" env:"SUBCMD_ARG"`
		} `positional-args:"yes"`
	}

	var opts struct {
		Sub SubCommand `command:"sub"`
	}

	p := NewParser(&opts, Default)
	_, err := p.ParseArgs([]string{"sub", "sub_from_cli"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertString(t, opts.Sub.Positional.Item, "sub_from_cli")
}
