package flags_test

import (
	"github.com/jessevdk/go-flags"
	"testing"
)

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

func parseSuccess(t *testing.T, data interface{}, args ...string) []string {
	ret, err := flags.ParseArgs(data, args)

	if err != nil {
		t.Fatalf("Unexpected parse error: %s", err)
		return nil
	}

	return ret
}

func TestShort(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	}{}

	ret := parseSuccess(t, &opts, "-v")

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

	ret := parseSuccess(t, &opts, "-vo", "-f")

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

	ret := parseSuccess(t, &opts, "-v", "-v")

	assertStringArray(t, ret, []string{})
	assertBoolArray(t, opts.Values, []bool{true, true})
}

func TestShortMultiSliceConcat(t *testing.T) {
	var opts = struct {
		Values []bool `short:"v"`
	}{}

	ret := parseSuccess(t, &opts, "-vvv")

	assertStringArray(t, ret, []string{})
	assertBoolArray(t, opts.Values, []bool{true, true, true})
}

