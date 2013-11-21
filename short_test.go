package flags_test

import (
	"github.com/jessevdk/go-flags"
	"testing"
)

func assertString(t *testing.T, a string, b string) {
	if a != b {
		t.Errorf("Expected %#v, but got %#v", b, a)
	}
}
func assertStringArray(t *testing.T, a []string, b []string) {
	if len(a) != len(b) {
		t.Errorf("Expected %#v, but got %#v", b, a)
		return
	}

	for i, v := range a {
		if b[i] != v {
			t.Errorf("Expected %#v, but got %#v", b, a)
			return
		}
	}
}

func assertBoolArray(t *testing.T, a []bool, b []bool) {
	if len(a) != len(b) {
		t.Errorf("Expected %#v, but got %#v", b, a)
		return
	}

	for i, v := range a {
		if b[i] != v {
			t.Errorf("Expected %#v, but got %#v", b, a)
			return
		}
	}
}

func assertParseSuccess(t *testing.T, data interface{}, args ...string) []string {
	parser := flags.NewParser(data, flags.Default &^ flags.PrintErrors)
	ret, err := parser.ParseArgs(args)

	if err != nil {
		t.Fatalf("Unexpected parse error: %s", err)
		return nil
	}

	return ret
}

func assertParseFail(t *testing.T, typ flags.ErrorType, msg string, data interface{}, args ...string) {
	parser := flags.NewParser(data, flags.Default &^ flags.PrintErrors)
	_, err := parser.ParseArgs(args)

	if err == nil {
		t.Fatalf("Expected error: %s", msg)
		return
	}

	if e, ok := err.(*flags.Error); !ok {
		t.Fatalf("Expected Error type, but got %#v", err)
		return
	} else {
		if e.Type != typ {
			t.Errorf("Expected error type {%s}, but got {%s}", typ, e.Type)
		}

		if e.Message != msg {
			t.Errorf("Expected error message %#v, but got %#v", msg, e.Message)
		}
	}
}

func TestShort(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v")

	assertStringArray(t, ret, []string{})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}
}

func TestShortMultiConcat(t *testing.T) {
	var opts = struct {
		V bool `short:"v"`
		O bool `short:"o"`
		F bool `short:"f"`
	}{}

	ret := assertParseSuccess(t, &opts, "-vo", "-f")

	assertStringArray(t, ret, []string{})

	if !opts.V {
		t.Errorf("Expected V to be true")
	}

	if !opts.O {
		t.Errorf("Expected O to be true")
	}

	if !opts.F {
		t.Errorf("Expected F to be true")
	}
}

func TestShortMultiSlice(t *testing.T) {
	var opts = struct {
		Values []bool `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v", "-v")

	assertStringArray(t, ret, []string{})
	assertBoolArray(t, opts.Values, []bool{true, true})
}

func TestShortMultiSliceConcat(t *testing.T) {
	var opts = struct {
		Values []bool `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-vvv")

	assertStringArray(t, ret, []string{})
	assertBoolArray(t, opts.Values, []bool{true, true, true})
}

func TestShortWithEqualArg(t *testing.T) {
	var opts = struct {
		Value string `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v=value")

	assertStringArray(t, ret, []string{})
	assertString(t, opts.Value, "value")
}

func TestShortWithArg(t *testing.T) {
	var opts = struct {
		Value string `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-vvalue")

	assertStringArray(t, ret, []string{})
	assertString(t, opts.Value, "value")
}

func TestShortArg(t *testing.T) {
	var opts = struct {
		Value string `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v", "value")

	assertStringArray(t, ret, []string{})
	assertString(t, opts.Value, "value")
}

func TestShortMultiWithEqualArg(t *testing.T) {
	var opts = struct {
		F []bool `short:"f"`
		Value string `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-ffv=value")

	assertStringArray(t, ret, []string{})
	assertBoolArray(t, opts.F, []bool{true, true})
	assertString(t, opts.Value, "value")
}

func TestShortMultiArg(t *testing.T) {
	var opts = struct {
		F []bool `short:"f"`
		Value string `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-ffv", "value")

	assertStringArray(t, ret, []string{})
	assertBoolArray(t, opts.F, []bool{true, true})
	assertString(t, opts.Value, "value")
}

func TestShortMultiArgConcatFail(t *testing.T) {
	var opts = struct {
		F []bool `short:"f"`
		Value string `short:"v"`
	}{}

	assertParseFail(t, flags.ErrUnknownFlag, "", &opts, "-ffvvalue")
}

func TestShortMultiArgConcat(t *testing.T) {
	var opts = struct {
		F []bool `short:"f"`
		Value string `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-vff")

	assertStringArray(t, ret, []string{})
	assertString(t, opts.Value, "ff")
}
