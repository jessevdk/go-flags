package flags_test

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"io/ioutil"
	"io"
	"os"
	"os/exec"
	"testing"
)

func helpDiff(a, b string) (string, error) {
	atmp, err := ioutil.TempFile("", "help-diff")

	if err != nil {
		return "", err
	}

	btmp, err := ioutil.TempFile("", "help-diff")

	if err != nil {
		return "", err
	}

	if _, err := io.WriteString(atmp, a); err != nil {
		return "", err
	}

	if _, err := io.WriteString(btmp, b); err != nil {
		return "", err
	}

	fmt.Println(atmp.Name(), btmp.Name())

	ret, err := exec.Command("diff", "-u", "-d", "--label", "got", atmp.Name(), "--label", "expected", btmp.Name()).Output()

	os.Remove(atmp.Name())
	os.Remove(btmp.Name())

	return string(ret), nil
}

func TestHelp(t *testing.T) {
	var opts = struct {
		Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
		Call func(string) `short:"c" description:"Call phone number"`
		PtrSlice []*string `long:"ptrslice" description:"A slice of pointers to string"`

		Other struct{
			StringSlice []string `short:"s" description:"A slice of strings"`
			IntMap map[string]int `long:"intmap" description:"A map from string to int"`
		} `group:"Other Options"`
	}{}

	p := flags.NewNamedParser("TestHelp", flags.HelpFlag)
	p.AddGroup("Application Options", "The application options", &opts)

	_, err := p.ParseArgs([]string{"--help"})

	if err == nil {
		t.Fatalf("Expected help error")
	}

	if e, ok := err.(*flags.Error); !ok {
		t.Fatalf("Expected flags.Error, but got %#T", err)
	} else {
		if e.Type != flags.ErrHelp {
			t.Errorf("Expected flags.ErrHelp type, but got %s", e.Type)
		}

		expected := `Usage:
  TestHelp [OPTIONS]

Application Options:
  -v, --verbose   Show verbose debug information
  -c=             Call phone number
      --ptrslice= A slice of pointers to string

Other Options:
  -s=             A slice of strings
      --intmap=   A map from string to int

Help Options:
  -h, --help      Show this help message
`

		if e.Message != expected {
			ret, err := helpDiff(e.Message, expected)

			if err != nil {
				t.Errorf("Unexpected diff error: %s", err)
				t.Errorf("Unexpected help message, expected:\n\n%s\n\nbut got\n\n%s", expected, e.Message)
			} else {
				t.Errorf("Unexpected help message:\n\n%s", ret)
			}
		}
	}
}

