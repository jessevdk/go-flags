package flags

import (
	"fmt"
	"net"
	"testing"
	"time"
)

type marshalled string

func (m *marshalled) UnmarshalFlag(value string) error {
	switch value {
	case "yes":
		*m = "true"
	case "no":
		*m = "false"
	default:
		return fmt.Errorf("`%s' is not a valid value, please specify `yes' or `no'", value)
	}

	return nil
}

func (m marshalled) MarshalFlag() (string, error) {
	if m == "true" {
		return "yes", nil
	}

	return "no", nil
}

type marshalledError bool

func (m marshalledError) MarshalFlag() (string, error) {
	return "", newErrorf(ErrMarshal, "Failed to marshal")
}

func TestUnmarshal(t *testing.T) {
	var opts = struct {
		Value marshalled `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v=yes")

	assertStringArray(t, ret, []string{})

	if opts.Value != "true" {
		t.Errorf("Expected Value to be \"true\"")
	}
}

func TestUnmarshalDefault(t *testing.T) {
	var opts = struct {
		Value marshalled `short:"v" default:"yes"`
	}{}

	ret := assertParseSuccess(t, &opts)

	assertStringArray(t, ret, []string{})

	if opts.Value != "true" {
		t.Errorf("Expected Value to be \"true\"")
	}
}

func TestUnmarshalOptional(t *testing.T) {
	var opts = struct {
		Value marshalled `short:"v" optional:"yes" optional-value:"yes"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v")

	assertStringArray(t, ret, []string{})

	if opts.Value != "true" {
		t.Errorf("Expected Value to be \"true\"")
	}
}

func TestUnmarshalError(t *testing.T) {
	var opts = struct {
		Value marshalled `short:"v"`
	}{}

	assertParseFail(t, ErrMarshal, fmt.Sprintf("invalid argument for flag `%cv' (expected flags.marshalled): `invalid' is not a valid value, please specify `yes' or `no'", defaultShortOptDelimiter), &opts, "-vinvalid")
}

func TestUnmarshalPositionalError(t *testing.T) {
	var opts = struct {
		Args struct {
			Value marshalled
		} `positional-args:"yes"`
	}{}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"invalid"})

	msg := "`invalid' is not a valid value, please specify `yes' or `no'"

	if err == nil {
		assertFatalf(t, "Expected error: %s", msg)
		return
	}

	if err.Error() != msg {
		assertErrorf(t, "Expected error message %#v, but got %#v", msg, err.Error())
	}
}

func TestMarshalError(t *testing.T) {
	var opts = struct {
		Value marshalledError `short:"v"`
	}{}

	p := NewParser(&opts, Default)
	o := p.Command.Groups()[0].Options()[0]

	_, err := convertToString(o.value, o.tag)

	assertError(t, err, ErrMarshal, "Failed to marshal")
}

type textMarshalled string

func (m *textMarshalled) UnmarshalText(text []byte) error {
	value := string(text)
	switch value {
	case "yes":
		*m = "true"
	case "no":
		*m = "false"
	default:
		return fmt.Errorf("`%s' is not a valid value, please specify `yes' or `no'", value)
	}

	return nil
}

func (m textMarshalled) MarshalText() ([]byte, error) {
	if m == "true" {
		return []byte("yes"), nil
	}

	return []byte("no"), nil
}

type textMarshalledError bool

func (m textMarshalledError) MarshalText() ([]byte, error) {
	return nil, newErrorf(ErrMarshal, "Failed to marshal")
}

type bothMarshalled string

func (m *bothMarshalled) UnmarshalFlag(value string) error {
	*m = "from-flag"
	return nil
}

func (m *bothMarshalled) UnmarshalText(text []byte) error {
	*m = "from-text"
	return nil
}

func (m bothMarshalled) MarshalFlag() (string, error) {
	return "from-flag", nil
}

func (m bothMarshalled) MarshalText() ([]byte, error) {
	return []byte("from-text"), nil
}

type textMarshalledStruct struct {
	Name string
}

func (m *textMarshalledStruct) UnmarshalText(text []byte) error {
	m.Name = string(text)
	return nil
}

func (m textMarshalledStruct) MarshalText() ([]byte, error) {
	return []byte(m.Name), nil
}

type textMarshalledBool bool

func (b *textMarshalledBool) UnmarshalText(text []byte) error {
	switch string(text) {
	case "yes", "true":
		*b = true
	case "no", "false":
		*b = false
	default:
		return fmt.Errorf("invalid bool: %s", text)
	}
	return nil
}

func TestTextUnmarshal(t *testing.T) {
	var opts = struct {
		Value textMarshalled `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v=yes")

	assertStringArray(t, ret, []string{})

	if opts.Value != "true" {
		t.Errorf("Expected Value to be \"true\"")
	}
}

func TestTextUnmarshalDefault(t *testing.T) {
	var opts = struct {
		Value textMarshalled `short:"v" default:"yes"`
	}{}

	ret := assertParseSuccess(t, &opts)

	assertStringArray(t, ret, []string{})

	if opts.Value != "true" {
		t.Errorf("Expected Value to be \"true\"")
	}
}

func TestTextUnmarshalOptional(t *testing.T) {
	var opts = struct {
		Value textMarshalled `short:"v" optional:"yes" optional-value:"yes"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v")

	assertStringArray(t, ret, []string{})

	if opts.Value != "true" {
		t.Errorf("Expected Value to be \"true\"")
	}
}

func TestTextUnmarshalError(t *testing.T) {
	var opts = struct {
		Value textMarshalled `short:"v"`
	}{}

	assertParseFail(t, ErrMarshal, fmt.Sprintf("invalid argument for flag `%cv' (expected flags.textMarshalled): `invalid' is not a valid value, please specify `yes' or `no'", defaultShortOptDelimiter), &opts, "-vinvalid")
}

func TestTextUnmarshalPositionalError(t *testing.T) {
	var opts = struct {
		Args struct {
			Value textMarshalled
		} `positional-args:"yes"`
	}{}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"invalid"})

	msg := "`invalid' is not a valid value, please specify `yes' or `no'"

	if err == nil {
		assertFatalf(t, "Expected error: %s", msg)
		return
	}

	if err.Error() != msg {
		assertErrorf(t, "Expected error message %#v, but got %#v", msg, err.Error())
	}
}

func TestTextMarshalError(t *testing.T) {
	var opts = struct {
		Value textMarshalledError `short:"v"`
	}{}

	p := NewParser(&opts, Default)
	o := p.Command.Groups()[0].Options()[0]

	_, err := convertToString(o.value, o.tag)

	assertError(t, err, ErrMarshal, "Failed to marshal")
}

func TestTextUnmarshalPrecedence(t *testing.T) {
	var opts = struct {
		Value bothMarshalled `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v=something")
	assertStringArray(t, ret, []string{})

	if opts.Value != "from-flag" {
		t.Errorf("Expected Value to be \"from-flag\", but got %q", opts.Value)
	}
}

func TestTextMarshalPrecedence(t *testing.T) {
	var opts = struct {
		Value bothMarshalled `short:"v"`
	}{
		Value: "val",
	}

	p := NewParser(&opts, Default)
	o := p.Command.Groups()[0].Options()[0]

	s, err := convertToString(o.value, o.tag)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if s != "from-flag" {
		t.Errorf("Expected \"from-flag\", but got %q", s)
	}
}

func TestTextUnmarshalPointer(t *testing.T) {
	var opts = struct {
		Value *textMarshalled `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v=yes")
	assertStringArray(t, ret, []string{})

	if opts.Value == nil {
		t.Fatal("Expected Value to not be nil")
	}
	if *opts.Value != "true" {
		t.Errorf("Expected Value to be \"true\", got %q", *opts.Value)
	}
}

func TestTextUnmarshalSlice(t *testing.T) {
	var opts = struct {
		Values []textMarshalled `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v=yes", "-v=no")
	assertStringArray(t, ret, []string{})

	if len(opts.Values) != 2 || opts.Values[0] != "true" || opts.Values[1] != "false" {
		t.Errorf("Expected [true, false], got %v", opts.Values)
	}
}

func TestTextUnmarshalMap(t *testing.T) {
	var opts = struct {
		Values map[string]textMarshalled `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v=key:yes")
	assertStringArray(t, ret, []string{})

	if opts.Values["key"] != "true" {
		t.Errorf("Expected key to be \"true\", got %q", opts.Values["key"])
	}
}

func TestTextUnmarshalStruct(t *testing.T) {
	var opts = struct {
		Value textMarshalledStruct `long:"val"`
	}{}

	ret := assertParseSuccess(t, &opts, "--val=hello")
	assertStringArray(t, ret, []string{})

	if opts.Value.Name != "hello" {
		t.Errorf("Expected Value.Name to be \"hello\", got %q", opts.Value.Name)
	}
}

func TestTextUnmarshalBool(t *testing.T) {
	var opts = struct {
		Value textMarshalledBool `long:"bool"`
	}{}

	ret := assertParseSuccess(t, &opts, "--bool=yes")
	assertStringArray(t, ret, []string{})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}
}

func TestTextUnmarshalStandardLibrary(t *testing.T) {
	var opts = struct {
		IP   net.IP    `long:"ip"`
		Time time.Time `long:"time"`
	}{}

	ret := assertParseSuccess(t, &opts, "--ip=192.168.1.1", "--time=2026-10-01T12:00:00Z")
	assertStringArray(t, ret, []string{})

	if opts.IP.String() != "192.168.1.1" {
		t.Errorf("Expected IP to be 192.168.1.1, got %v", opts.IP)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2026-10-01T12:00:00Z")
	if !opts.Time.Equal(expectedTime) {
		t.Errorf("Expected Time to be %v, got %v", expectedTime, opts.Time)
	}
}
