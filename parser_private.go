package flags

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func (p *Parser) storeDefaults() {
	p.EachGroup(func(index int, grp *Group) {
		grp.storeDefaults()
	})
}

func (p *Parser) parseOption(group *Group, args []string, name string, option *Option, canarg bool, argument *string, index int) (error, int, *Option) {
	var err error

	if !option.canArgument() {
		if canarg && argument != nil {
			return newError(ErrNoArgumentForBool,
					fmt.Sprintf("bool flag `%s' cannot have an argument", option)),
				index,
				option
		}

		err = option.Set(nil)
	} else if canarg && (argument != nil || index < len(args) && !argumentIsOption(args[index])) {
		if argument == nil {
			argument = &args[index]
			index++
		}

		err = option.Set(argument)
	} else if option.OptionalArgument {
		option.clear()

		for _, v := range option.OptionalValue {
			err = option.Set(&v)

			if err != nil {
				break
			}
		}
	} else {
		return newError(ErrExpectedArgument,
				fmt.Sprintf("expected argument for flag `%s'", option)),
			index,
			option
	}

	if err != nil {
		if _, ok := err.(*Error); !ok {
			err = newError(ErrMarshal,
				fmt.Sprintf("invalid argument for flag `%s' (expected %s): %s",
					option,
					option.Value.Type(),
					err.Error()))
		}
	}

	return err, index, option
}

func (p *Parser) parseLong(args []string, name string, argument *string, index int) (error, int, *Option) {
	name = strings.ToLower(name)

	var option *Option
	var group *Group

	p.EachGroup(func(index int, grp *Group) {
		if opt := grp.LongNames[name]; opt != nil {
			option = opt
			group = grp
		}
	})

	if option != nil {
		return p.parseOption(group, args, name, option, true, argument, index)
	}

	return newError(ErrUnknownFlag,
			fmt.Sprintf("unknown flag `%s'", name)),
		index,
		nil
}

func (p *Parser) getShort(name rune) (*Option, *Group) {
	var option *Option
	var group *Group

	p.EachGroup(func(index int, grp *Group) {
		if opt := grp.ShortNames[name]; opt != nil {
			option = opt
			group = grp
		}
	})

	if option != nil {
		return option, group
	}

	return nil, nil
}

func (p *Parser) parseShort(args []string, name rune, islast bool, argument *string, index int) (error, int, *Option) {
	names := make([]byte, utf8.RuneLen(name))
	utf8.EncodeRune(names, name)

	option, grp := p.getShort(name)

	if option != nil {
		if option.canArgument() && !islast && !option.OptionalArgument {
			return newError(ErrExpectedArgument,
					fmt.Sprintf("expected argument for flag `%s'", option)),
				index,
				option
		}

		return p.parseOption(grp, args, string(names), option, islast, argument, index)
	}

	return newError(ErrUnknownFlag,
			fmt.Sprintf("unknown flag `%s'", string(names))),
		index,
		nil
}

func (p *Parser) parseIni(ini Ini) error {
	for groupName, section := range ini {
		group := p.GroupsMap[strings.ToLower(groupName)]

		if group == nil {
			return newError(ErrUnknownGroup,
				fmt.Sprintf("could not find option group `%s'", groupName))
		}

		for _, inival := range section {
			opt, usedName := group.lookupByName(inival.Name, true)

			if opt == nil {
				if (p.Options & IgnoreUnknown) == None {
					return newError(ErrUnknownFlag,
						fmt.Sprintf("unknown option: %s", inival.Name))
				}

				continue
			}

			if opt.tag.Get("no-ini") != "" {
				continue
			}

			opt.iniUsedName = usedName

			pval := &inival.Value

			if opt.isBool() && len(inival.Value) == 0 {
				pval = nil
			}

			if err := opt.Set(pval); err != nil {
				return wrapError(err)
			}
		}
	}

	return nil
}

func (p *Parser) currentCommander() *Commander {
	if p.currentCommand != nil {
		return &p.currentCommand.Commander
	}

	return &p.Commander
}

func (p *Parser) eachTopLevelGroup(cb func(int, *Group)) {
	index := 0

	for _, group := range p.Groups {
		if !group.IsCommand {
			index = group.each(index, cb)
		}
	}
}
