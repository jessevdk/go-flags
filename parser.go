// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

import (
	"fmt"
	"os"
	"path"
	"strings"
	"unicode/utf8"
)

// A Parser provides command line option parsing. It can contain several
// option groups each with their own set of options.
type Parser struct {
	// The option groups available to the parser
	Groups    []*Group
	GroupsMap map[string]*Group

	// The parser application name
	ApplicationName string

	// The usage (e.g. [OPTIONS] <filename>)
	Usage string

	Options Options
}

// Parser options
type Options uint

const (
	// No options
	None Options = 0

	// Add a default Help Options group to the parser containing -h and
	// --help options. When either -h or --help is specified on the command
	// line, a pretty formatted help message will be printed to os.Stderr.
	// The parser will return ErrHelp.
	ShowHelp = 1 << iota

	// Pass all arguments after a double dash, --, as remaining command line
	// arguments (i.e. they will not be parsed for flags)
	PassDoubleDash

	// Ignore any unknown options and pass them as remaining command line
	// arguments
	IgnoreUnknown

	// Print any errors which occured during parsing to os.Stderr
	PrintErrors

	// A convenient default set of options
	Default = ShowHelp | PrintErrors | PassDoubleDash
)

// Parse is a convenience function to parse command line options with default
// settings. The provided data is a pointer to a struct representing the
// default option group (named "Application Options"). For more control, use
// flags.NewParser.
func Parse(data interface{}) ([]string, error) {
	return NewParser(data, Default).Parse()
}

// ParseIni is a convenience function to parse command line options with default
// settings from an ini file. The provided data is a pointer to a struct
// representing the default option group (named "Application Options"). For
// more control, use flags.NewParser.
func ParseIni(filename string, data interface{}) error {
	return NewParser(data, Default).ParseIni(filename)
}

// NewParser creates a new parser. It uses os.Args[0] as the application
// name and then calls Parser.NewNamedParser (see Parser.NewNamedParser for
// more details). The provided data is a pointer to a struct representing the
// default option group (named "Application Options"), or nil if the default
// group should not be added. The options parameter specifies a set of options
// for the parser.
func NewParser(data interface{}, options Options) *Parser {
	if data == nil {
		return NewNamedParser(path.Base(os.Args[0]), options)
	}

	return NewNamedParser(path.Base(os.Args[0]), options, NewGroup("Application Options", data))
}

// NewNamedParser creates a new parser. The appname is used to display the
// executable name in the builtin help message. An initial set of option groups
// can be specified when constructing a parser, but you can also add additional
// option groups later (see Parser.AddGroup).
func NewNamedParser(appname string, options Options, groups ...*Group) *Parser {
	ret := &Parser{
		ApplicationName: appname,
		Groups:          groups,
		GroupsMap:       make(map[string]*Group),
		Options:         options,
		Usage:           "[OPTIONS]",
	}

	for _, group := range groups {
		ret.GroupsMap[group.Name] = group
	}

	return ret
}

// AddGroup adds a new group to the parser with the given name and data. The
// data needs to be a pointer to a struct from which the fields indicate which
// options are in the group.
func (p *Parser) AddGroup(name string, data interface{}) *Parser {
	group := NewGroup(name, data)

	p.Groups = append(p.Groups, group)
	p.GroupsMap[name] = group

	return p
}

// Parse parses the command line arguments from os.Args using Parser.ParseArgs.
// For more detailed information see ParseArgs.
func (p *Parser) Parse() ([]string, error) {
	return p.ParseArgs(os.Args[1:])
}

// Parse parses the command line arguments from os.Args using Parser.ParseArgs.
// For more detailed information see ParseArgs. The returned errors can be of
// the type flags.Error or flags.IniError.
func (p *Parser) ParseIni(filename string) error {
	ini, err := readIni(filename)

	if err != nil {
		return err
	}

	for groupName, section := range ini {
		group := p.GroupsMap[groupName]

		if group == nil {
			return newError(ErrUnknownGroup,
				fmt.Sprintf("could not find option group `%s'", groupName))
		}

		for name, val := range section {
			opt := group.lookupByName(name, true)

			if opt == nil {
				if (p.Options & IgnoreUnknown) == None {
					return newError(ErrUnknownFlag,
						fmt.Sprintf("unknown option: %s", name))
				}

				continue
			}

			if opt.options.Get("no-ini") != "" {
				continue
			}

			pval := &val

			if opt.isBool() && len(val) == 0 {
				pval = nil
			}

			if err := opt.Set(pval); err != nil {
				return wrapError(err)
			}
		}
	}

	return nil
}

// ParseArgs parses the command line arguments according to the option groups that
// were added to the parser. On successful parsing of the arguments, the
// remaining, non-option, arguments (if any) are returned. The returned error
// indicates a parsing error and can be used with PrintError to display
// contextual information on where the error occurred exactly.
//
// When the common help group has been added (AddHelp) and either -h or --help
// was specified in the command line arguments, a help message will be
// automatically printed. Furthermore, the special error ErrHelp is returned
// to indicate that the help was shown. It is up to the caller to exit the
// program if so desired.
func (p *Parser) ParseArgs(args []string) ([]string, error) {
	ret := make([]string, 0, len(args))
	i := 0

	if (p.Options & ShowHelp) != None {
		var help struct {
			ShowHelp func() error `short:"h" long:"help" description:"Show this help message"`
		}

		help.ShowHelp = func() error {
			p.ShowHelp(os.Stderr)
			return ErrHelp
		}

		p.Groups = append([]*Group{NewGroup("Help Options", &help)}, p.Groups...)
		p.Options &^= ShowHelp
	}

	for i < len(args) {
		arg := args[i]
		i++

		// When PassDoubleDash is set and we encounter a --, then
		// simply append all the rest as arguments and break out
		if (p.Options&PassDoubleDash) != None && arg == "--" {
			ret = append(ret, args[i:]...)
			break
		}

		// If the argument is not an option, then append it to the rest
		if arg[0] != '-' {
			ret = append(ret, arg)
			continue
		}

		pos := strings.Index(arg, "=")
		var argument *string

		if pos >= 0 {
			rest := arg[pos+1:]
			argument = &rest
			arg = arg[:pos]
		}

		var err error

		if strings.HasPrefix(arg, "--") {
			err, i = p.parseLong(args, arg[2:], argument, i)
		} else {
			short := arg[1:]

			for j, c := range short {
				clen := utf8.RuneLen(c)
				islast := (j+clen == len(short))

				if !islast && argument == nil {
					rr := short[j+clen:]
					next, _ := utf8.DecodeRuneInString(rr)
					info, _ := p.getShort(c)

					if info != nil && info.canArgument() {
						if snext, _ := p.getShort(next); snext == nil {
							// Consider the next stuff as an argument
							argument = &rr
							islast = true
						}
					}
				}

				err, i = p.parseShort(args, c, islast, argument, i)

				if err != nil || islast {
					break
				}
			}
		}

		if err != nil {
			if (p.Options & IgnoreUnknown) != None {
				ret = append(ret, arg)
			} else {
				if (p.Options&PrintErrors) != None && err != ErrHelp {
					fmt.Fprintf(os.Stderr, "Flags error: %s\n", err.Error())
				}

				return nil, wrapError(err)
			}
		}
	}

	return ret, nil
}
