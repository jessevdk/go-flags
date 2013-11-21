package flags_test

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"testing"
)

type marshalled bool

func (m *marshalled) UnmarshalFlag(value string) error {
	if value == "yes" {
		*m = true
	} else if value == "no" {
		*m = false
	} else {
		return fmt.Errorf("`%s' is not a valid value, please specify `yes' or `no'", value)
	}

	return nil
}

func (m marshalled) MarshalFlag() string {
	if m {
		return "yes"
	}

	return "no"
}

func TestMarshal(t *testing.T) {
	var opts = struct {
		Value marshalled `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v=yes")

	assertStringArray(t, ret, []string{})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}
}

func TestMarshalError(t *testing.T) {
	var opts = struct {
		Value marshalled `short:"v"`
	}{}

	assertParseFail(t, flags.ErrMarshal, "invalid argument for flag `-v' (expected flags_test.marshalled): `invalid' is not a valid value, please specify `yes' or `no'", &opts, "-vinvalid")
}
