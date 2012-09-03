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

func (p *Parser) parseOption(group *Group, args []string, name string, info *Info, canarg bool, argument *string, index int) (error, int) {
	var err error

	if !info.canArgument() {
		if canarg && argument != nil {
			return newError(ErrNoArgumentForBool,
			                fmt.Sprintf("bool flag `%s' cannot have an argument", info)),
			       index
		}

		err = info.Set(nil)
	} else if canarg && (argument != nil || index < len(args)) {
		if argument == nil {
			argument = &args[index]
			index++
		}

		err = info.Set(argument)
	} else if info.OptionalArgument {
		err = info.Set(&info.Default)
	} else {
		return newError(ErrExpectedArgument,
			        fmt.Sprintf("expected argument for flag `%s'", info)),
			index
	}

	if err != nil {
		if _, ok := err.(*Error); !ok {
			err = newError(ErrMarshal,
			               fmt.Sprintf("invalid argument for flag `%s' (expected %s)",
			                           info,
			                           info.value.Type()))
		}
	}

	return err, index
}

func (p *Parser) parseLong(args []string, name string, argument *string, index int) (error, int) {
	for _, grp := range p.Groups {
		if info := grp.LongNames[name]; info != nil {
			return p.parseOption(grp, args, name, info, true, argument, index)
		}
	}

	return newError(ErrUnknownFlag,
		        fmt.Sprintf("unknown flag `%s'", name)),
		index
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
		if info.canArgument() && !islast && !info.OptionalArgument {
			return newError(ErrExpectedArgument,
			                fmt.Sprintf("expected argument for flag `%s'", info)),
			        index
		}

		return p.parseOption(grp, args, string(names), info, islast, argument, index)
	}

	return newError(ErrUnknownFlag,
		        fmt.Sprintf("unknown flag `%s'", string(names))),
		index
}
