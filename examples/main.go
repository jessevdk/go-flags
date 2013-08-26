package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

type EditorOptions struct {
	Input  string `short:"i" long:"input" description:"Input file" default:"-"`
	Output string `short:"o" long:"output" description:"Output file" default:"-"`
}

type Options struct {
	// Example of verbosity with level
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`

	// Example of optional value
	User string `short:"u" long:"user" description:"User name" optional:"yes" optional-value:"pancake"`

	// Example of map with multiple default values
	Users map[string]string `long:"users" description:"User e-mail map" default:"system:system@example.org" default:"admin:admin@example.org"`

	// Example of option group
	Editor EditorOptions `group:"Editor Options"`
}

var options Options

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}
