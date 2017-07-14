package flags

import (
	"testing"
)

func TestTagMissingColon(t *testing.T) {
	var opts = struct {
		Value bool `short`
	}{}

	assertParseFail(t, ErrTag, "expected `:' after key name, but got end of tag (in `short`)", &opts, "")
}

func TestTagMissingValue(t *testing.T) {
	var opts = struct {
		Value bool `short:`
	}{}

	assertParseFail(t, ErrTag, "expected `\"' to start tag value at end of tag (in `short:`)", &opts, "")
}

func TestTagMissingQuote(t *testing.T) {
	var opts = struct {
		Value bool `short:"v`
	}{}

	assertParseFail(t, ErrTag, "expected end of tag value `\"' at end of tag (in `short:\"v`)", &opts, "")
}

func TestTagNewline(t *testing.T) {
	var opts = struct {
		Value bool `long:"verbose" description:"verbose
something"`
	}{}

	assertParseFail(t, ErrTag, "unexpected newline in tag value `description' (in `long:\"verbose\" description:\"verbose\nsomething\"`)", &opts, "")
}

func TestTagMultiline(t *testing.T) {
	type Opts struct {
		Value int `
			long:"value"
			default:"41"
			description:"value of something"`
	}

	{
		var opts Opts
		assertParseSuccess(t, &opts, "--value=42")
		if opts.Value != 42 {
			t.Fatal("expect value to be 42")
		}
	}

	{
		var opts Opts
		assertParseSuccess(t, &opts, "")
		if opts.Value != 41 {
			t.Fatalf("expect value %d to be 41", opts.Value)
		}
	}
}
