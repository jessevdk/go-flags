package flags

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestDurationFlags(t *testing.T) {
	var opts = struct {
		DurationLong  time.Duration `long:"duration"`
		DurationShort time.Duration `short:"d"`
	}{}

	ret := assertParseSuccess(t, &opts, "--duration=1h2m3s", "-d", "45s")
	assertStringArray(t, ret, []string{})

	expectedLong, _ := time.ParseDuration("1h2m3s")
	expectedShort, _ := time.ParseDuration("45s")

	if opts.DurationLong != expectedLong {
		t.Errorf("Expected DurationLong to be %v, got %v", expectedLong, opts.DurationLong)
	}

	if opts.DurationShort != expectedShort {
		t.Errorf("Expected DurationShort to be %v, got %v", expectedShort, opts.DurationShort)
	}
}

func TestDurationDefault(t *testing.T) {
	var opts = struct {
		Duration time.Duration `long:"duration" default:"10m"`
	}{}

	ret := assertParseSuccess(t, &opts)
	assertStringArray(t, ret, []string{})

	expected, _ := time.ParseDuration("10m")
	if opts.Duration != expected {
		t.Errorf("Expected Duration to be %v, got %v", expected, opts.Duration)
	}
}

func TestDurationOptional(t *testing.T) {
	var opts = struct {
		Duration time.Duration `long:"duration" optional:"yes" optional-value:"5s"`
	}{}

	ret := assertParseSuccess(t, &opts, "--duration")
	assertStringArray(t, ret, []string{})

	expected, _ := time.ParseDuration("5s")
	if opts.Duration != expected {
		t.Errorf("Expected Duration to be %v, got %v", expected, opts.Duration)
	}
}

func TestDurationEnv(t *testing.T) {
	os.Setenv("TEST_DURATION_ENV", "30s")
	defer os.Unsetenv("TEST_DURATION_ENV")

	var opts = struct {
		Duration time.Duration `long:"duration" env:"TEST_DURATION_ENV"`
	}{}

	ret := assertParseSuccess(t, &opts)
	assertStringArray(t, ret, []string{})

	expected, _ := time.ParseDuration("30s")
	if opts.Duration != expected {
		t.Errorf("Expected Duration to be %v, got %v", expected, opts.Duration)
	}
}

func TestDurationSlice(t *testing.T) {
	var opts = struct {
		Durations []time.Duration `long:"duration"`
	}{}

	ret := assertParseSuccess(t, &opts, "--duration=1s", "--duration=2m", "--duration=3h")
	assertStringArray(t, ret, []string{})

	d1, _ := time.ParseDuration("1s")
	d2, _ := time.ParseDuration("2m")
	d3, _ := time.ParseDuration("3h")
	expected := []time.Duration{d1, d2, d3}

	if !reflect.DeepEqual(opts.Durations, expected) {
		t.Errorf("Expected %v, got %v", expected, opts.Durations)
	}
}

func TestDurationSliceDefaults(t *testing.T) {
	var opts = struct {
		Durations []time.Duration `long:"duration" default:"100ms" default:"200ms"`
	}{}

	ret := assertParseSuccess(t, &opts)
	assertStringArray(t, ret, []string{})

	d1, _ := time.ParseDuration("100ms")
	d2, _ := time.ParseDuration("200ms")
	expected := []time.Duration{d1, d2}

	if !reflect.DeepEqual(opts.Durations, expected) {
		t.Errorf("Expected %v, got %v", expected, opts.Durations)
	}
}

func TestDurationMap(t *testing.T) {
	var opts = struct {
		Durations map[string]time.Duration `long:"duration"`
	}{}

	ret := assertParseSuccess(t, &opts, "--duration=first:1s", "--duration=second:2m")
	assertStringArray(t, ret, []string{})

	d1, _ := time.ParseDuration("1s")
	d2, _ := time.ParseDuration("2m")
	expected := map[string]time.Duration{
		"first":  d1,
		"second": d2,
	}

	if !reflect.DeepEqual(opts.Durations, expected) {
		t.Errorf("Expected %v, got %v", expected, opts.Durations)
	}
}

func TestDurationPointer(t *testing.T) {
	var opts = struct {
		Duration *time.Duration `long:"duration"`
	}{}

	ret := assertParseSuccess(t, &opts, "--duration=15s")
	assertStringArray(t, ret, []string{})

	expected, _ := time.ParseDuration("15s")
	if opts.Duration == nil || *opts.Duration != expected {
		t.Errorf("Expected %v, got %v", expected, opts.Duration)
	}
}

func TestDurationPositional(t *testing.T) {
	var opts = struct {
		Args struct {
			Timeout time.Duration `positional-arg-name:"timeout"`
		} `positional-args:"yes"`
	}{}

	ret := assertParseSuccess(t, &opts, "250ms")
	assertStringArray(t, ret, []string{})

	expected, _ := time.ParseDuration("250ms")
	if opts.Args.Timeout != expected {
		t.Errorf("Expected Timeout to be %v, got %v", expected, opts.Args.Timeout)
	}
}

func TestDurationInvalid(t *testing.T) {
	var opts = struct {
		Duration time.Duration `long:"duration"`
	}{}

	assertParseFail(t, ErrMarshal, fmt.Sprintf("invalid argument for flag `%sduration' (expected time.Duration): time: invalid duration \"not-a-duration\"", defaultLongOptDelimiter), &opts, "--duration=not-a-duration")
}

func TestDurationPositionalInvalid(t *testing.T) {
	var opts = struct {
		Args struct {
			Timeout time.Duration `positional-arg-name:"timeout"`
		} `positional-args:"yes"`
	}{}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"invalid"})

	msg := "time: invalid duration \"invalid\""

	if err == nil {
		assertFatalf(t, "Expected error: %s", msg)
		return
	}

	if err.Error() != msg {
		assertErrorf(t, "Expected error message %#v, but got %#v", msg, err.Error())
	}
}

func TestDurationConvertToString(t *testing.T) {
	d, _ := time.ParseDuration("2h30m")
	var opts = struct {
		Duration time.Duration `long:"duration"`
	}{
		Duration: d,
	}

	p := NewParser(&opts, Default)
	o := p.Command.Groups()[0].Options()[0]

	s, err := convertToString(o.value, o.tag)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assertString(t, s, "2h30m0s")
}

func TestDurationConvertToStringPointer(t *testing.T) {
	d, _ := time.ParseDuration("45s")
	var opts = struct {
		Duration *time.Duration `long:"duration"`
		Nil      *time.Duration `long:"nil"`
	}{
		Duration: &d,
		Nil:      nil,
	}

	p := NewParser(&opts, Default)
	optsList := p.Command.Groups()[0].Options()

	s, err := convertToString(optsList[0].value, optsList[0].tag)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	assertString(t, s, "45s")

	sNil, err := convertToString(optsList[1].value, optsList[1].tag)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	assertString(t, sNil, "")
}

func TestDurationIni(t *testing.T) {
	var opts = struct {
		Duration time.Duration `long:"duration"`
	}{}

	iniParser := NewIniParser(NewParser(&opts, Default))
	iniData := "[Application Options]\nduration = 15m\n"
	err := iniParser.Parse(strings.NewReader(iniData))
	if err != nil {
		t.Fatalf("Unexpected error parsing INI: %v", err)
	}

	expected, _ := time.ParseDuration("15m")
	if opts.Duration != expected {
		t.Errorf("Expected Duration %v, got %v", expected, opts.Duration)
	}
}
