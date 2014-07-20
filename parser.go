// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

import (
	"os"
	"path"
	"strings"
)

// A Parser provides command line option parsing. It can contain several
// option groups each with their own set of options.
type Parser struct {
	// Embedded, see Command for more information
	*Command

	// A usage string to be displayed in the help message.
	Usage string

	// Option flags changing the behavior of the parser.
	Options Options

	// NamespaceDelimiter separates group namespaces and option long names
	NamespaceDelimiter string

	internalError error
}

// Options provides parser options that change the behavior of the option
// parser.
type Options uint

const (
	// None indicates no options.
	None Options = 0

	// HelpFlag adds a default Help Options group to the parser containing
	// -h and --help options. When either -h or --help is specified on the
	// command line, the parser will return the special error of type
	// ErrHelp. When PrintErrors is also specified, then the help message
	// will also be automatically printed to os.Stderr.
	HelpFlag = 1 << iota

	// PassDoubleDash passes all arguments after a double dash, --, as
	// remaining command line arguments (i.e. they will not be parsed for
	// flags).
	PassDoubleDash

	// IgnoreUnknown ignores any unknown options and passes them as
	// remaining command line arguments instead of generating an error.
	IgnoreUnknown

	// PrintErrors prints any errors which occurred during parsing to
	// os.Stderr.
	PrintErrors

	// PassAfterNonOption passes all arguments after the first non option
	// as remaining command line arguments. This is equivalent to strict
	// POSIX processing.
	PassAfterNonOption

	// Default is a convenient default set of options which should cover
	// most of the uses of the flags package.
	Default = HelpFlag | PrintErrors | PassDoubleDash
)

// Parse is a convenience function to parse command line options with default
// settings. The provided data is a pointer to a struct representing the
// default option group (named "Application Options"). For more control, use
// flags.NewParser.
func Parse(data interface{}) ([]string, error) {
	return NewParser(data, Default).Parse()
}

// ParseArgs is a convenience function to parse command line options with default
// settings. The provided data is a pointer to a struct representing the
// default option group (named "Application Options"). The args argument is
// the list of command line arguments to parse. If you just want to parse the
// default program command line arguments (i.e. os.Args), then use flags.Parse
// instead. For more control, use flags.NewParser.
func ParseArgs(data interface{}, args []string) ([]string, error) {
	return NewParser(data, Default).ParseArgs(args)
}

// NewParser creates a new parser. It uses os.Args[0] as the application
// name and then calls Parser.NewNamedParser (see Parser.NewNamedParser for
// more details). The provided data is a pointer to a struct representing the
// default option group (named "Application Options"), or nil if the default
// group should not be added. The options parameter specifies a set of options
// for the parser.
func NewParser(data interface{}, options Options) *Parser {
	p := NewNamedParser(path.Base(os.Args[0]), options)

	if data != nil {
		g, err := p.AddGroup("Application Options", "", data)

		if err == nil {
			g.parent = p
		}

		p.internalError = err
	}

	return p
}

// NewNamedParser creates a new parser. The appname is used to display the
// executable name in the built-in help message. Option groups and commands can
// be added to this parser by using AddGroup and AddCommand.
func NewNamedParser(appname string, options Options) *Parser {
	p := &Parser{
		Command:            newCommand(appname, "", "", nil),
		Options:            options,
		NamespaceDelimiter: ".",
	}

	p.Command.parent = p

	if len(os.Getenv("GO_FLAGS_COMPLETION")) != 0 {
		p.AddCommand("__complete", "completion", "automatic flags completion", &completion{parser: p})
	}

	return p
}

// Parse parses the command line arguments from os.Args using Parser.ParseArgs.
// For more detailed information see ParseArgs.
func (p *Parser) Parse() ([]string, error) {
	return p.ParseArgs(os.Args[1:])
}

// ParseArgs parses the command line arguments according to the option groups that
// were added to the parser. On successful parsing of the arguments, the
// remaining, non-option, arguments (if any) are returned. The returned error
// indicates a parsing error and can be used with PrintError to display
// contextual information on where the error occurred exactly.
//
// When the common help group has been added (AddHelp) and either -h or --help
// was specified in the command line arguments, a help message will be
// automatically printed. Furthermore, the special error type ErrHelp is returned.
// It is up to the caller to exit the program if so desired.
func (p *Parser) ParseArgs(args []string) ([]string, error) {
	if p.internalError != nil {
		return nil, p.internalError
	}

	p.clearIsSet()

	// Add built-in help group to all commands if necessary
	if (p.Options & HelpFlag) != None {
		p.addHelpGroups(p.showBuiltinHelp)
	}

	s := &parseState{
		args:    args,
		retargs: make([]string, 0, len(args)),
	}

	p.fillParseState(s)

	for !s.eof() {
		arg := s.pop()

		// When PassDoubleDash is set and we encounter a --, then
		// simply append all the rest as arguments and break out
		if (p.Options&PassDoubleDash) != None && arg == "--" {
			s.addArgs(s.args...)
			break
		}

		if !argumentIsOption(arg) {
			// Note: this also sets s.err, so we can just check for
			// nil here and use s.err later
			if p.parseNonOption(s) != nil {
				break
			}

			continue
		}

		var err error

		prefix, optname, islong := stripOptionPrefix(arg)
		optname, _, argument := splitOption(prefix, optname, islong)

		if islong {
			err = p.parseLong(s, optname, argument)
		} else {
			err = p.parseShort(s, optname, argument)
		}

		if err != nil {
			ignoreUnknown := (p.Options & IgnoreUnknown) != None
			parseErr := wrapError(err)

			if !(parseErr.Type == ErrUnknownFlag && ignoreUnknown) {
				s.err = parseErr
				break
			}

			if ignoreUnknown {
				s.addArgs(arg)
			}
		}
	}

	if s.err == nil {
		p.eachCommand(func(c *Command) {
			c.eachGroup(func(g *Group) {
				for _, option := range g.options {
					if option.isSet {
						continue
					}

					option.clearDefault()
				}
			})
		}, true)

		s.checkRequired(p)
	}

	var reterr error

	if s.err != nil {
		reterr = p.printError(s.err)
	} else if len(s.command.commands) != 0 && !s.command.SubcommandsOptional {
		reterr = p.printError(s.estimateCommand())
	} else if cmd, ok := s.command.data.(Commander); ok {
		reterr = p.printError(cmd.Execute(s.retargs))
	} else if cmd, ok := s.command.data.(CommanderNoArgs); ok {
		if len(s.retargs) > 0 {
			reterr = p.printError(newErrorf(ErrTooManyArgs,
				"too many arguments: %s",
				strings.Join(s.retargs, ", ")))
		} else {
			reterr = p.printError(cmd.Execute())
		}
	}

	if reterr != nil {
		return append([]string{s.arg}, s.args...), reterr
	}

	return s.retargs, nil
}
