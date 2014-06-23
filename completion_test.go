package flags

import (
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type TestComplete struct {
}

func (t *TestComplete) Complete(match string) []Completion {
	options := []string{
		"hello world",
		"hello universe",
		"hello multiverse",
	}

	ret := make([]Completion, 0, len(options))

	for _, o := range options {
		if strings.HasPrefix(o, match) {
			ret = append(ret, Completion{
				Item: o,
			})
		}
	}

	return ret
}

var completionTestOptions struct {
	Verbose bool `short:"v" long:"verbose"`
	Debug   bool `short:"d" long:"debug"`
	Version bool `long:"version"`

	AddCommand struct {
		Positional struct {
			Filename Filename
		} `positional-args:"yes"`
	} `command:"add"`

	RemoveCommand struct {
		Other bool     `short:"o"`
		File  Filename `short:"f" long:"filename"`
	} `command:"rm"`

	RenameCommand struct {
		Completed TestComplete `short:"c" long:"completed"`
	} `command:"rename"`
}

func TestCompletion(t *testing.T) {
	_, sourcefile, _, _ := runtime.Caller(0)
	sourcedir := filepath.Join(filepath.SplitList(path.Dir(sourcefile))...)

	excompl := []string{filepath.Join(sourcedir, "completion.go"), filepath.Join(sourcedir, "completion_test.go")}

	tests := []struct {
		Args      []string
		Completed []string
	}{
		{
			// Short names
			[]string{"-"},
			[]string{"-d", "-v"},
		},

		{
			// Short names concatenated
			[]string{"-dv"},
			[]string{"-dv"},
		},

		{
			// Long names
			[]string{"--"},
			[]string{"--debug", "--verbose", "--version"},
		},

		{
			// Long names partial
			[]string{"--ver"},
			[]string{"--verbose", "--version"},
		},

		{
			// Commands
			[]string{""},
			[]string{"add", "rename", "rm"},
		},

		{
			// Commands partial
			[]string{"r"},
			[]string{"rename", "rm"},
		},

		{
			// Positional filename
			[]string{"add", filepath.Join(sourcedir, "completion")},
			excompl,
		},

		{
			// Flag filename
			[]string{"rm", "-f", path.Join(sourcedir, "completion")},
			excompl,
		},

		{
			// Flag short concat last filename
			[]string{"rm", "-of", path.Join(sourcedir, "completion")},
			excompl,
		},

		{
			// Flag concat filename
			[]string{"rm", "-f" + path.Join(sourcedir, "completion")},
			[]string{"-f" + excompl[0], "-f" + excompl[1]},
		},

		{
			// Flag equal concat filename
			[]string{"rm", "-f=" + path.Join(sourcedir, "completion")},
			[]string{"-f=" + excompl[0], "-f=" + excompl[1]},
		},

		{
			// Flag concat long filename
			[]string{"rm", "--filename=" + path.Join(sourcedir, "completion")},
			[]string{"--filename=" + excompl[0], "--filename=" + excompl[1]},
		},

		{
			// Flag long filename
			[]string{"rm", "--filename", path.Join(sourcedir, "completion")},
			excompl,
		},

		{
			// Custom completed
			[]string{"rename", "-c", "hello un"},
			[]string{"hello universe"},
		},
	}

	p := NewParser(&completionTestOptions, Default)
	c := &completion{parser: p}

	for _, test := range tests {
		ret := c.complete(test.Args)
		items := make([]string, len(ret))

		for i, v := range ret {
			items[i] = v.Item
		}

		if !reflect.DeepEqual(items, test.Completed) {
			t.Errorf("Args: %#v\n  Expected: %#v\n  Got:     %#v", test.Args, test.Completed, items)
		}
	}
}
