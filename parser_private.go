package flags

import (
	"fmt"
	"strings"
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

func (p *Parser) storeDefaults() {
	for _, grp := range p.Groups {
		grp.storeDefaults()
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
	name = strings.ToLower(name)

	var option *Option
	var group *Group

	for _, grp := range p.Groups {
		if opt := grp.LongNames[name]; opt != nil {
			option = opt
			group = grp
		}
	}

	if option != nil {
		return p.parseOption(group, args, name, option, true, argument, index)
	}

	return newError(ErrUnknownFlag,
			fmt.Sprintf("unknown flag `%s'", name)),
		index
}

func (p *Parser) getShort(name rune) (*Option, *Group) {
	var option *Option
	var group *Group

	for _, grp := range p.Groups {
		if opt := grp.ShortNames[name]; opt != nil {
			option = opt
			group = grp
		}
	}

	if option != nil {
		return option, group
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

func (p *Parser) parseIni(ini Ini) error {
	for groupName, section := range ini {
		group := p.GroupsMap[groupName]

		if group == nil {
			return newError(ErrUnknownGroup,
				fmt.Sprintf("could not find option group `%s'", groupName))
		}

		for name, val := range section {
			opt, usedName := group.lookupByName(name, true)

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

			opt.iniUsedName = usedName

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
