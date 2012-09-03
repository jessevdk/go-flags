package flags

import (
	"fmt"
	"unicode/utf8"
)

func (p *Parser) removeGroup(group *Group) {
	for i, grp := range p.Groups {
		if grp == group {
			p.Groups = append(p.Groups[:i], p.Groups[i+1:]...)
			break
		}
	}
}

func (p *Parser) parseOption(group *Group, args []string, name string, option *Option, canarg bool, argument *string, index int) (error, int) {
	var err error

	if !option.canArgument() {
		if canarg && argument != nil {
			return newError(ErrNoArgumentForBool,
			                fmt.Sprintf("bool flag `%s' cannot have an argument", option)),
			       index
		}

		err = option.Set(nil)
	} else if canarg && (argument != nil || index < len(args)) {
		if argument == nil {
			argument = &args[index]
			index++
		}

		err = option.Set(argument)
	} else if option.OptionalArgument {
		err = option.Set(&option.Default)
	} else {
		return newError(ErrExpectedArgument,
			        fmt.Sprintf("expected argument for flag `%s'", option)),
			index
	}

	if err != nil {
		if _, ok := err.(*Error); !ok {
			err = newError(ErrMarshal,
			               fmt.Sprintf("invalid argument for flag `%s' (expected %s)",
			                           option,
			                           option.value.Type()))
		}
	}

	return err, index
}

func (p *Parser) parseLong(args []string, name string, argument *string, index int) (error, int) {
	for _, grp := range p.Groups {
		if option := grp.LongNames[name]; option != nil {
			return p.parseOption(grp, args, name, option, true, argument, index)
		}
	}

	return newError(ErrUnknownFlag,
		        fmt.Sprintf("unknown flag `%s'", name)),
		index
}

func (p *Parser) getShort(name rune) (*Option, *Group) {
	for _, grp := range p.Groups {
		option := grp.ShortNames[name]

		if option != nil {
			return option, grp
		}
	}

	return nil, nil
}

func (p *Parser) parseShort(args []string, name rune, islast bool, argument *string, index int) (error, int) {
	names := make([]byte, utf8.RuneLen(name))
	utf8.EncodeRune(names, name)

	option, grp := p.getShort(name)

	if option != nil {
		if option.canArgument() && !islast && !option.OptionalArgument {
			return newError(ErrExpectedArgument,
			                fmt.Sprintf("expected argument for flag `%s'", option)),
			        index
		}

		return p.parseOption(grp, args, string(names), option, islast, argument, index)
	}

	return newError(ErrUnknownFlag,
		        fmt.Sprintf("unknown flag `%s'", string(names))),
		index
}
