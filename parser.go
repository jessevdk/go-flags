// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"unicode/utf8"
)

// A Parser provides command line option parsing. It can contain several
// option groups each with their own set of options.
type Parser struct {
	Commander

	// The option groups available to the parser
	Groups    []*Group
	GroupsMap map[string]*Group

	// The parser application name
	ApplicationName string

	// The application short description
	Description string

	// The usage (e.g. [OPTIONS] <filename>)
	Usage string

	Options Options

	currentCommand       *Group
	currentCommandString []string
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
	HelpFlag = 1 << iota

	// Pass all arguments after a double dash, --, as remaining command line
	// arguments (i.e. they will not be parsed for flags)
	PassDoubleDash

	// Ignore any unknown options and pass them as remaining command line
	// arguments
	IgnoreUnknown

	// Print any errors which occured during parsing to os.Stderr
	PrintErrors

	// Pass all arguments after the first non option. This is equivalent
	// to strict POSIX processing
	PassAfterNonOption

	// A convenient default set of options
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

// ParseIni is a convenience function to parse command line options with default
// settings from an ini file. The provided data is a pointer to a struct
// representing the default option group (named "Application Options"). For
// more control, use flags.NewParser.
func ParseIni(filename string, data interface{}) error {
	return NewParser(data, Default).ParseIniFile(filename)
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
		Commander: Commander{
			Commands: make(map[string]*Group),
		},

		ApplicationName: appname,
		Groups:          groups,
		GroupsMap:       make(map[string]*Group),
		Options:         options,
		Usage:           "[OPTIONS]",
	}

	ret.EachGroup(func(index int, grp *Group) {
		ret.GroupsMap[strings.ToLower(grp.Name)] = grp
	})

	return ret
}

// AddGroup adds a new group to the parser with the given name and data. The
// data needs to be a pointer to a struct from which the fields indicate which
// options are in the group.
func (p *Parser) AddGroup(name string, data interface{}) *Parser {
	group := NewGroup(name, data)

	p.Groups = append(p.Groups, group)

	group.each(0, func(index int, grp *Group) {
		p.GroupsMap[strings.ToLower(group.Name)] = grp
	})

	return p
}

// AddCommand adds a new command to the parser with the given name and data. The
// data needs to be a pointer to a struct from which the fields indicate which
// options are in the command.
func (p *Parser) AddCommand(command string, description string, longDescription string, data interface{}) *Parser {
	p.AddGroup(description, data)

	group := p.Groups[len(p.Groups)-1]
	group.IsCommand = true
	group.LongDescription = longDescription

	p.Commands[command] = group

	return p
}

// Parse parses the command line arguments from os.Args using Parser.ParseArgs.
// For more detailed information see ParseArgs.
func (p *Parser) Parse() ([]string, error) {
	return p.ParseArgs(os.Args[1:])
}

// ParseIniFile parses flags from an ini formatted file. See ParseIni for more
// information on the ini file foramt. The returned errors can be of the type
// flags.Error or flags.IniError.
func (p *Parser) ParseIniFile(filename string) error {
	p.storeDefaults()

	ini, err := readIniFromFile(filename)

	if err != nil {
		return err
	}

	return p.parseIni(ini)
}

func (p *Parser) EachGroup(cb func(int, *Group)) {
	if p.currentCommand != nil {
		p.currentCommand.each(0, cb)
	} else {
		p.eachTopLevelGroup(cb)
	}
}

// ParseIni parses flags from an ini format. You can use ParseIniFile as a
// convenience function to parse from a filename instead of a general
// io.Reader.
//
// The format of the ini file is as follows:
//
// [Option group name]
// option = value
//
// Each section in the ini file represents an option group in the flags parser.
// The default flags parser option group (i.e. when using flags.Parse) is
// named 'Application Options'. The ini option name is matched in the following
// order:
//
// 1. Compared to the ini-name tag on the option struct field (if present)
// 2. Compared to the struct field name
// 3. Compared to the option long name (if present)
// 4. Compared to the option short name (if present)
//
// The returned errors can be of the type flags.Error or
// flags.IniError.
func (p *Parser) ParseIni(reader io.Reader) error {
	p.storeDefaults()

	ini, err := readIni(reader, "")

	if err != nil {
		return err
	}

	return p.parseIni(ini)
}

// WriteIniToFile writes the flags as ini format into a file. See WriteIni
// for more information. The returned error occurs when the specified file
// could not be opened for writing.
func (p *Parser) WriteIniToFile(filename string, options IniOptions) error {
	file, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer file.Close()
	p.WriteIni(file, options)

	return nil
}

// WriteIni writes the current values of all the flags to an ini format.
// See ParseIni for more information on the ini file format. You typically
// call this only after settings have been parsed since the default values of each
// option are stored just before parsing the flags (this is only relevant when
// IniIncludeDefaults is _not_ set in options).
func (p *Parser) WriteIni(writer io.Writer, options IniOptions) {
	writeIni(p, writer, options)
}

func (p *Parser) levenshtein(s string, t string) int {
	if len(s) == 0 {
		return len(t)
	}

	if len(t) == 0 {
		return len(s)
	}

	var l1, l2, l3 int

	if len(s) == 1 {
		l1 = len(t) + 1
	} else {
		l1 = p.levenshtein(s[1:len(s)-1], t) + 1
	}

	if len(t) == 1 {
		l2 = len(s) + 1
	} else {
		l2 = p.levenshtein(t[1:len(t)-1], s) + 1
	}

	l3 = p.levenshtein(s[1:len(s)], t[1:len(t)])

	if s[0] != t[0] {
		l3 += 1
	}

	if l2 < l1 {
		l1 = l2
	}

	if l1 < l3 {
		return l1
	}

	return l3
}

func (p *Parser) closest(cmd string, commands []string) (string, int) {
	if len(commands) == 0 {
		return "", 0
	}

	mincmd := -1
	mindist := -1

	for i, c := range commands {
		l := p.levenshtein(cmd, c)

		if mincmd < 0 || l < mindist {
			mindist = l
			mincmd = i
		}
	}

	return commands[mincmd], mindist
}

func argumentIsOption(arg string) bool {
	return len(arg) > 0 && arg[0] == '-'
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
	ret := make([]string, 0, len(args))
	i := 0

	p.storeDefaults()

	if (p.Options & HelpFlag) != None {
		var help struct {
			ShowHelp func() error `short:"h" long:"help" description:"Show this help message"`
		}

		help.ShowHelp = func() error {
			var b bytes.Buffer
			p.WriteHelp(&b)
			return newError(ErrHelp, b.String())
		}

		helpgrp := NewGroup("Help Options", &help)

		// Append the help group to toplevel
		p.Groups = append([]*Group{helpgrp}, p.Groups...)

		// Also append the help group to every command
		p.Commander.EachCommand(func(command string, grp *Group) {
			grp.EmbeddedGroups = append([]*Group{helpgrp}, grp.EmbeddedGroups...)
		})

		p.Options &^= HelpFlag
	}

	required := make(map[*Option]struct{})
	commands := make(map[string]*Group)

	for command, subgroup := range p.Commands {
		commands[command] = subgroup
	}

	// Mark required arguments in a map
	for _, group := range p.Groups {
		for _, option := range group.Options {
			if option.Required {
				required[option] = struct{}{}
			}
		}

		// Initial set of commands
		for command, subgroup := range group.Commands {
			commands[command] = subgroup
		}
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

		// If the argument is not an option then
		// 1) Check for subcommand
		// 2) Append it to the rest if subcommand is not found
		if !argumentIsOption(arg) {
			if cmdgroup := commands[arg]; cmdgroup != nil {
				// Set current 'root' group
				p.currentCommand = cmdgroup

				p.currentCommandString = append(p.currentCommandString,
					arg)

				commands = cmdgroup.Commands
			} else {
				if (p.Options & PassAfterNonOption) != None {
					ret = append(ret, args[(i-1):]...)
					break
				} else {
					ret = append(ret, arg)
				}
			}

			continue
		}

		pos := strings.Index(arg, "=")
		var argument *string
		var optname string

		if pos >= 0 {
			rest := arg[pos+1:]
			argument = &rest
			optname = arg[:pos]
		} else {
			optname = arg
		}

		var err error
		var option *Option

		if strings.HasPrefix(optname, "--") {
			err, i, option = p.parseLong(args, optname[2:], argument, i)
		} else {
			short := optname[1:]

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

				err, i, option = p.parseShort(args, c, islast, argument, i)

				if err != nil || islast {
					break
				}
			}
		}

		if err != nil {
			ignoreUnknown := (p.Options & IgnoreUnknown) != None

			parseErr, ok := err.(*Error)
			if !ok {
				parseErr = newError(ErrUnknown, err.Error())
			}

			if ignoreUnknown {
				ret = append(ret, arg)
			}

			if !(parseErr.Type == ErrUnknownFlag && ignoreUnknown) {
				if (p.Options & PrintErrors) != None {
					if parseErr.Type == ErrHelp {
						fmt.Fprintln(os.Stderr, err)
					} else {
						fmt.Fprintf(os.Stderr, "Flags error: %s\n", err.Error())
					}
				}

				return nil, wrapError(err)
			}
		} else {
			delete(required, option)
		}
	}

	if len(required) > 0 {
		names := make([]string, 0, len(required))

		for k, _ := range required {
			names = append(names, "`"+k.String()+"'")
		}

		var msg string

		if len(names) == 1 {
			msg = fmt.Sprintf("the required flag %s was not specified", names[0])
		} else {
			msg = fmt.Sprintf("the required flags %s and %s were not specified",
				strings.Join(names[:len(names)-1], ", "), names[len(names)-1])
		}

		err := newError(ErrRequired, msg)

		if (p.Options & PrintErrors) != None {
			fmt.Fprintln(os.Stderr, err)
		}

		return nil, err
	}

	if p.currentCommand != nil {
		// Execute group
		cmd, ok := p.currentCommand.data.(Command)

		if ok {
			err := cmd.Execute(ret)

			if err != nil && (p.Options&PrintErrors) != None {
				fmt.Fprintln(os.Stderr, err)
			}

			return nil, err
		}
	} else if len(commands) != 0 {
		cmdnames := make([]string, 0, len(commands))

		for k, _ := range commands {
			cmdnames = append(cmdnames, k)
		}

		sort.Strings(cmdnames)
		var msg string

		if len(ret) != 0 {
			c, l := p.closest(ret[0], cmdnames)
			msg = fmt.Sprintf("Unknown command `%s'", ret[0])

			if float32(l)/float32(len(c)) < 0.5 {
				msg = fmt.Sprintf("%s, did you mean `%s'?", msg, c)
			} else {
				msg = fmt.Sprintf("%s. Please specify one command of: %s",
					msg,
					strings.Join(cmdnames, ", "))
			}
		} else {
			msg = fmt.Sprintf("Please specify one command of: %s", strings.Join(cmdnames, ", "))
		}

		err := newError(ErrRequired, msg)

		if (p.Options & PrintErrors) != None {
			fmt.Fprintln(os.Stderr, msg)
		}

		return nil, err
	}

	return ret, nil
}
