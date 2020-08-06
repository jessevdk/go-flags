package flags

import (
	"testing"
)

func TestParserBuilder(t *testing.T) {
	type Opts struct {
		Value  bool `short:"v"`
		Nested struct {
			Foo string `long:"foo" default:"z" env:"FOO"`
		} `group:"nested" namespace:"nested" env-namespace:"NESTED"`
	}

	var opts = Opts{}
	var err error

	_, err = NewParser(&opts, None).WithNamespaceDelimiter("_").ParseArgs([]string{"--nested_foo=\"val\""})
	if err != nil {
		assertErrorf(t, "Parser returned unexpected error %v", err)
	}
	assertString(t, opts.Nested.Foo, "val")

	_, err = NewParser(&opts, None).WithNamespaceDelimiter("-").ParseArgs([]string{"--nested-foo=\"val\""})
	if err != nil {
		assertErrorf(t, "Parser returned unexpected error %v", err)
	}
	assertString(t, opts.Nested.Foo, "val")
}
