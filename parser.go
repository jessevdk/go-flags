package flags

import (
	"unicode/utf8"
	"strings"
	"io"
	"fmt"
	"errors"
)

var ErrExpectedArgument = errors.New("expected option argument")
var ErrUnknownFlag = errors.New("unknown flag")
var ErrHelp = errors.New("help shown")
var ErrNoArgumentForBool = errors.New("bool flags cannot have arguments")

type Help struct {
	writer io.Writer
	IsHelp bool `long:"help" short:"h" description:"Show help options"`
}

type Parser struct {
	Groups []*Group
	ApplicationName string
	Usage string
	PassDoubleDash bool
	IgnoreUnknown bool

	Help *Help
	ErrorAt interface{}
}

func NewParser(appname string, groups ...*Group) *Parser {
	return &Parser {
		ApplicationName: appname,
		Groups: groups,
		PassDoubleDash: false,
		IgnoreUnknown: false,
		Usage: "[OPTIONS]",
	}
}

func (p *Parser) AddGroup(groups ...*Group) {
	p.Groups = append(p.Groups, groups...)
}

func (p *Parser) AddHelp(writer io.Writer) {
	if p.Help == nil {
		p.Help = &Help {
			writer: writer,
		}

		p.AddGroup(NewGroup("Help Options", p.Help))
	}
}

func (p *Parser) parseOption(group *Group, args []string, name string, info *Info, canarg bool, argument *string, index int) (error, int) {
	if info.IsBool() {
		if canarg && argument != nil {
			p.ErrorAt = info
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
		p.ErrorAt = info
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

	p.ErrorAt = fmt.Sprintf("--%v", name)
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
			p.ErrorAt = info
			return ErrExpectedArgument, index
		}

		return p.parseOption(grp, args, string(names), info, islast, argument, index)
	}

	p.ErrorAt = fmt.Sprintf("-%v", string(names))
	return ErrUnknownFlag, index
}

func (p *Parser) PrintError(writer io.Writer, err error) {
	s := fmt.Sprintf("Error at `%v': %s\n", p.ErrorAt, err)
	writer.Write([]byte(s))
}

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
				islast := (j + clen == len(short))

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

	if p.Help != nil && p.Help.IsHelp {
		p.ShowHelp(p.Help.writer)
		p.ErrorAt = "--help"

		return ret, ErrHelp
	}

	return ret, nil
}
