package flags

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
)

// Completer is an interface which can be implemented by types
// to provide custom command line argument completion.
type Completer interface {
	// Complete receives a prefix representing a (partial) value
	// for its type and should provide a list of possible valid
	// completions.
	Complete(match string) []string
}

type completion struct {
	parser *Parser
}

// Filename is a string alias which provides filename completion.
type Filename string

// Complete returns a list of existing files with the given
// prefix.
func (f *Filename) Complete(match string) []string {
	ret, _ := filepath.Glob(match + "*")
	return ret
}

func (c *completion) skipPositional(s *parseState, n int) {
	if n >= len(s.positional) {
		s.positional = nil
	} else {
		s.positional = s.positional[n:]
	}
}

func (c *completion) completeOptionNames(names map[string]*Option, prefix string, match string) []string {
	n := make([]string, 0, len(names))

	for k, _ := range names {
		if strings.HasPrefix(k, match) {
			n = append(n, prefix+k)
		}
	}

	return n
}

func (c *completion) completeLongNames(s *parseState, prefix string, match string) []string {
	return c.completeOptionNames(s.lookup.longNames, prefix, match)
}

func (c *completion) completeShortNames(s *parseState, prefix string, match string) []string {
	if len(match) != 0 {
		return []string{prefix + match}
	}

	return c.completeOptionNames(s.lookup.shortNames, prefix, match)
}

func (c *completion) completeCommands(s *parseState, match string) []string {
	n := make([]string, 0, len(s.command.commands))

	for _, cmd := range s.command.commands {
		if cmd.data != c && strings.HasPrefix(cmd.Name, match) {
			n = append(n, cmd.Name)
		}
	}

	return n
}

func (c *completion) completeValue(value reflect.Value, prefix string, match string) []string {
	i := value.Interface()

	var ret []string

	if cmp, ok := i.(Completer); ok {
		ret = cmp.Complete(match)
	} else if value.CanAddr() {
		if cmp, ok = value.Addr().Interface().(Completer); ok {
			ret = cmp.Complete(match)
		}
	}

	for i, v := range ret {
		ret[i] = prefix + v
	}

	return ret
}

func (c *completion) complete(args []string) []string {
	if len(args) == 0 {
		args = []string{""}
	}

	s := &parseState{
		args: args,
	}

	c.parser.fillParseState(s)

	var opt *Option

	for len(s.args) > 1 {
		arg := s.pop()

		if (c.parser.Options&PassDoubleDash) != None && arg == "--" {
			opt = nil
			c.skipPositional(s, len(s.args)-1)

			break
		}

		if argumentIsOption(arg) {
			prefix, optname, islong := stripOptionPrefix(arg)
			optname, _, argument := splitOption(prefix, optname, islong)

			if argument == nil {
				var o *Option
				canarg := true

				if islong {
					o = s.lookup.longNames[optname]
				} else {
					for i, r := range optname {
						sname := string(r)
						o = s.lookup.shortNames[sname]

						if o == nil {
							break
						}

						if i == 0 && o.canArgument() && len(optname) != len(sname) {
							canarg = false
							break
						}
					}
				}

				if o == nil && (c.parser.Options&PassAfterNonOption) != None {
					opt = nil
					c.skipPositional(s, len(s.args)-1)

					break
				} else if o != nil && o.canArgument() && !o.OptionalArgument && canarg {
					if len(s.args) > 1 {
						s.pop()
					} else {
						opt = o
					}
				}
			}
		} else {
			if len(s.positional) > 0 {
				s.positional = s.positional[1:]
			} else if cmd, ok := s.lookup.commands[arg]; ok {
				cmd.fillParseState(s)
			}

			opt = nil
		}
	}

	lastarg := s.args[len(s.args)-1]
	var ret []string

	if opt != nil {
		// Completion for the argument of 'opt'
		ret = c.completeValue(opt.value, "", lastarg)
	} else if argumentIsOption(lastarg) {
		// Complete the option
		prefix, optname, islong := stripOptionPrefix(lastarg)
		optname, split, argument := splitOption(prefix, optname, islong)

		if argument == nil && !islong {
			rname, n := utf8.DecodeRuneInString(optname)
			sname := string(rname)

			if opt := s.lookup.shortNames[sname]; opt != nil && opt.canArgument() {
				ret = c.completeValue(opt.value, prefix+sname, optname[n:])
			} else {
				ret = c.completeShortNames(s, prefix, optname)
			}
		} else if argument != nil {
			if islong {
				opt = s.lookup.longNames[optname]
			} else {
				opt = s.lookup.shortNames[optname]
			}

			if opt != nil {
				ret = c.completeValue(opt.value, prefix+optname+split, *argument)
			}
		} else if islong {
			ret = c.completeLongNames(s, prefix, optname)
		} else {
			ret = c.completeShortNames(s, prefix, optname)
		}
	} else if len(s.positional) > 0 {
		// Complete for positional argument
		ret = c.completeValue(s.positional[0].value, "", lastarg)
	} else if len(s.command.commands) > 0 {
		// Complete for command
		ret = c.completeCommands(s, lastarg)
	}

	sort.Strings(ret)
	return ret
}

func (c *completion) Execute(args []string) error {
	ret := c.complete(args)

	for _, v := range ret {
		fmt.Println(v)
	}

	return nil
}
