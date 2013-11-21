package flags_test

import (
	"github.com/jessevdk/go-flags"
	"testing"
)

func TestGroupInline(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Group struct {
			G bool `short:"g"`
		} `group:"Grouped Options"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "-v", "-g")

	assertStringArray(t, ret, []string{})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Group.G {
		t.Errorf("Expected Group.G to be true")
	}

	if p.Command.Group.Find("Grouped Options") == nil {
		t.Errorf("Expected to find group `Grouped Options'")
	}
}

func TestGroupAdd(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`
	} {}

	var grp = struct {
		G bool `short:"g"`
	} {}

	p := flags.NewParser(&opts, flags.Default)
	g, err := p.AddGroup("Grouped Options", "", &grp)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	ret, err := p.ParseArgs([]string{"-v", "-g", "rest"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	assertStringArray(t, ret, []string{"rest"})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !grp.G {
		t.Errorf("Expected Group.G to be true")
	}

	if p.Command.Group.Find("Grouped Options") != g {
		t.Errorf("Expected to find group `Grouped Options'")
	}

	if p.Groups()[1] != g {
		t.Errorf("Espected group #v,	 but got #v", g, p.Groups()[0])
	}

	if g.Options()[0].ShortName != 'g' {
		t.Errorf("Expected short name `g' but got %v", g.Options()[0].ShortName)
	}
}

func TestGroupNestedInline(t *testing.T) {
	var opts = struct {
		Value bool `short:"v"`

		Group struct {
			G bool `short:"g"`

			Nested struct {
				N string `long:"n"`
			} `group:"Nested Options"`
		} `group:"Grouped Options"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "-v", "-g", "--n", "n", "rest")

	assertStringArray(t, ret, []string{"rest"})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Group.G {
		t.Errorf("Expected Group.G to be true")
	}

	assertString(t, opts.Group.Nested.N, "n")

	if p.Command.Group.Find("Grouped Options") == nil {
		t.Errorf("Expected to find group `Grouped Options'")
	}

	if p.Command.Group.Find("Nested Options") == nil {
		t.Errorf("Expected to find group `Nested Options'")
	}
}
