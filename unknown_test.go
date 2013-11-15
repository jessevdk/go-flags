package flags

import (
	"github.com/jessevdk/go-flags"
	"testing"
)

func TestUnknownFlags(t *testing.T) {
	var opts = struct {
		Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
	}{}

	args := []string{
		"-f",
	}

	p := flags.NewParser(&opts, 0)
	args, err := p.ParseArgs(args)

	if err == nil {
		t.Fatal("Expected error for unknown argument")
	}
}

func TestIgnoreUnknownFlags(t *testing.T) {
	var opts = struct {
		Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
	}{}

	args := []string{
		"hello",
		"world",
		"-v",
		"--foo=bar",
		"--verbose",
		"-f",
	}

	p := flags.NewParser(&opts, flags.IgnoreUnknown)
	args, err := p.ParseArgs(args)

	if err != nil {
		t.Fatal(err)
	}

	exargs := []string{
		"hello",
		"world",
		"--foo=bar",
		"-f",
	}

	issame := (len(args) == len(exargs))

	if issame {
		for i := 0; i < len(args); i++ {
			if args[i] != exargs[i] {
				issame = false
				break
			}
		}
	}

	if !issame {
		t.Fatalf("Expected %v but got %v", exargs, args)
	}
}
