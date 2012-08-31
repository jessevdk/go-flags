package flags

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Expected an argument but got none
var ErrExpectedArgument = errors.New("expected option argument")

// Unknown option flag was found
var ErrUnknownFlag = errors.New("unknown flag")

// The builtin help message was printed
var ErrHelp = errors.New("help shown")

// An argument for a boolean value was specified
var ErrNoArgumentForBool = errors.New("bool flags cannot have arguments")

type help struct {
	writer io.Writer
	IsHelp bool `long:"help" short:"h" description:"Show help options"`
}

// A Parser provides command line option parsing. It can contain several
// option groups each with their own set of options.
//
// Example:
//     type Options struct {
//         Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"
//     }
//
//     opts := new(Options)
//
//     parser := flags.NewParser("testapp")
//     parser.AddHelp(os.Stderr)
//
//     parser.AddGroup(flags.NewGroup("Application Options", opts))
//
//     args, err := parser.Parser(os.Args[1:])
//
//     if err != nil {
//         if err != flags.ErrHelp {
//             parser.PrintError(os.Stderr)
//         }
//
//         os.Exit(1)
//     }
type Parser struct {
	// The option groups available to the parser
	Groups          []*Group

	// The parser application name
	ApplicationName string

	// The usage (e.g. [OPTIONS] <filename>)
	Usage           string

	// If true, all arguments after a double dash (--) will be passed
	// as remaining command line arguments. Defaults to false.
	PassDoubleDash  bool

	// If true, unknown command line arguments are ignored and passed as
	// remaining command line arguments. Defaults to false.
	IgnoreUnknown   bool

	help    *help
	errorAt interface{}
}

// NewParser creates a new parser. The appname is used to display the
// executable name in the builtin help message. An initial set of option groups
// can be specified when constructing a parser, but you can also add additional
// option groups later (see Parser.AddGroup).
func NewParser(appname string, groups ...*Group) *Parser {
	return &Parser{
		ApplicationName: appname,
		Groups:          groups,
		PassDoubleDash:  false,
		IgnoreUnknown:   false,
		Usage:           "[OPTIONS]",
	}
}

// AddGroup adds one or more option groups to the parser. It returns the parser
// itself again so multiple calls can be chained.
func (p *Parser) AddGroup(groups ...*Group) *Parser {
	p.Groups = append(p.Groups, groups...)
	return p
}

// AddHelp adds a common help option group to the parser. The help option group
// contains only one option, with short name -h and long name --help.
// When either -h or --help appears, a help message listing all the options
// will be written to the specified writer. It returns the parser itself
// again so multiple calls can be chained.
func (p *Parser) AddHelp(writer io.Writer) *Parser {
	if p.help == nil {
		p.help = &help{
			writer: writer,
		}

		p.AddGroup(NewGroup("Help Options", p.help))
	}

	return p
}

func (p *Parser) parseOption(group *Group, args []string, name string, info *Info, canarg bool, argument *string, index int) (error, int) {
	if info.IsBool() {
		if canarg && argument != nil {
			p.errorAt = info
			return ErrNoArgumentForBool, index
		}

		info.Set("")
	} else if canarg && (argument != nil || index < len(args)) {
		if argument == nil {
			argument = &args[index]
			index++
		}

		info.Set(*argument)
	} else if info.OptionalArgument {
		info.Set(info.Default)
	} else {
		p.errorAt = info
		return ErrExpectedArgument, index
	}

	return nil, index
}

func (p *Parser) parseLong(args []string, name string, argument *string, index int) (error, int) {
	for _, grp := range p.Groups {
		if info := grp.LongNames[name]; info != nil {
			return p.parseOption(grp, args, name, info, true, argument, index)
		}
	}

	p.errorAt = fmt.Sprintf("--%v", name)
	return ErrUnknownFlag, index
}

func (p *Parser) getShort(name rune) (*Info, *Group) {
	for _, grp := range p.Groups {
		info := grp.ShortNames[name]

		if info != nil {
			return info, grp
		}
	}

	return nil, nil
}

func (p *Parser) parseShort(args []string, name rune, islast bool, argument *string, index int) (error, int) {
	names := make([]byte, utf8.RuneLen(name))
	utf8.EncodeRune(names, name)

	info, grp := p.getShort(name)

	if info != nil {
		if !info.IsBool() && !islast && !info.OptionalArgument {
			p.errorAt = info
			return ErrExpectedArgument, index
		}

		return p.parseOption(grp, args, string(names), info, islast, argument, index)
	}

	p.errorAt = fmt.Sprintf("-%v", string(names))
	return ErrUnknownFlag, index
}

// PrintError prints an error which occurred while parsing command line
// arguments. This is more useful than simply printing the error message
// because context information on where the error occurred will also be
// written. The error is printed to writer (commonly os.Stderr).
func (p *Parser) PrintError(writer io.Writer, err error) {
	s := fmt.Sprintf("Error at `%v': %s\n", p.errorAt, err)
	writer.Write([]byte(s))
}

// Parse parses the command line arguments according to the option groups that
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
func (p *Parser) Parse(args []string) ([]string, error) {
	ret := make([]string, 0, len(args))
	i := 0

	for i < len(args) {
		arg := args[i]
		i++

		// When PassDoubleDash is set and we encounter a --, then
		// simply append all the rest as arguments and break out
		if p.PassDoubleDash && arg == "--" {
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

					if info != nil && !info.IsBool() {
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
			if p.IgnoreUnknown {
				ret = append(ret, arg)
			} else {
				return nil, err
			}
		}
	}

	if p.help != nil && p.help.IsHelp {
		p.ShowHelp(p.help.writer)
		p.errorAt = "--help"

		return ret, ErrHelp
	}

	return ret, nil
}
