package flags

import (
	"bytes"
	"io"
	"os"
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
	Verbose  bool `short:"v" long:"verbose"`
	Debug    bool `short:"d" long:"debug"`
	Version  bool `long:"version"`
	Required bool `long:"required" required:"true"`

	AddCommand struct {
		Positional struct {
			Filename Filename
		} `positional-args:"yes"`
	} `command:"add"`

	AddMultiCommand struct {
		Positional struct {
			Filename []Filename
		} `positional-args:"yes"`
	} `command:"add-multi"`

	RemoveCommand struct {
		Other bool     `short:"o"`
		File  Filename `short:"f" long:"filename"`
	} `command:"rm"`

	RenameCommand struct {
		Completed TestComplete `short:"c" long:"completed"`
	} `command:"rename"`
}

type completionTest struct {
	Args      []string
	Completed []string
}

var completionTests []completionTest

func init() {
	_, sourcefile, _, _ := runtime.Caller(0)
	completionTestSourcedir := filepath.Join(filepath.SplitList(path.Dir(sourcefile))...)

	completionTestFilename := []string{filepath.Join(completionTestSourcedir, "completion.go"), filepath.Join(completionTestSourcedir, "completion_test.go")}

	completionTests = []completionTest{
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
			[]string{"--debug", "--required", "--verbose", "--version"},
		},

		{
			// Long names partial
			[]string{"--ver"},
			[]string{"--verbose", "--version"},
		},

		{
			// Commands
			[]string{""},
			[]string{"add", "add-multi", "rename", "rm"},
		},

		{
			// Commands partial
			[]string{"r"},
			[]string{"rename", "rm"},
		},

		{
			// Positional filename
			[]string{"add", filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},

		{
			// Multiple positional filename (1 arg)
			[]string{"add-multi", filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},
		{
			// Multiple positional filename (2 args)
			[]string{"add-multi", filepath.Join(completionTestSourcedir, "completion.go"), filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},
		{
			// Multiple positional filename (3 args)
			[]string{"add-multi", filepath.Join(completionTestSourcedir, "completion.go"), filepath.Join(completionTestSourcedir, "completion.go"), filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},

		{
			// Flag filename
			[]string{"rm", "-f", path.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},

		{
			// Flag short concat last filename
			[]string{"rm", "-of", path.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},

		{
			// Flag concat filename
			[]string{"rm", "-f" + path.Join(completionTestSourcedir, "completion")},
			[]string{"-f" + completionTestFilename[0], "-f" + completionTestFilename[1]},
		},

		{
			// Flag equal concat filename
			[]string{"rm", "-f=" + path.Join(completionTestSourcedir, "completion")},
			[]string{"-f=" + completionTestFilename[0], "-f=" + completionTestFilename[1]},
		},

		{
			// Flag concat long filename
			[]string{"rm", "--filename=" + path.Join(completionTestSourcedir, "completion")},
			[]string{"--filename=" + completionTestFilename[0], "--filename=" + completionTestFilename[1]},
		},

		{
			// Flag long filename
			[]string{"rm", "--filename", path.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},

		{
			// Custom completed
			[]string{"rename", "-c", "hello un"},
			[]string{"hello universe"},
		},
	}
}

func TestCompletion(t *testing.T) {
	p := NewParser(&completionTestOptions, Default)
	c := &completion{parser: p}

	for _, test := range completionTests {
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

func TestParserCompletion(t *testing.T) {
	os.Setenv("GO_FLAGS_COMPLETION", "1")

	for _, test := range completionTests {
		tmp := os.Stdout

		r, w, _ := os.Pipe()
		os.Stdout = w

		out := make(chan string)

		go func() {
			var buf bytes.Buffer

			io.Copy(&buf, r)

			out <- buf.String()
		}()

		p := NewParser(&completionTestOptions, None)

		_, err := p.ParseArgs(append([]string{"__complete", "--"}, test.Args...))

		w.Close()

		os.Stdout = tmp

		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		got := strings.Split(strings.Trim(<-out, "\n"), "\n")

		if !reflect.DeepEqual(got, test.Completed) {
			t.Errorf("Expected: %#v\nGot: %#v", test.Completed, got)
		}
	}

	os.Setenv("GO_FLAGS_COMPLETION", "")
}
