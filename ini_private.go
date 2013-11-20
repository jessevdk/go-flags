package flags

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

type IniValue struct {
	Name  string
	Value string
}

type IniSection []IniValue
type Ini map[string]IniSection

func readFullLine(reader *bufio.Reader) (string, error) {
	var line []byte

	for {
		l, more, err := reader.ReadLine()

		if err != nil {
			return "", err
		}

		if line == nil && !more {
			return string(l), nil
		}

		line = append(line, l...)

		if !more {
			break
		}
	}

	return string(line), nil
}

type IniError struct {
	Message    string
	File       string
	LineNumber uint
}

func (x *IniError) Error() string {
	return fmt.Sprintf("%s:%d: %s",
		x.File,
		x.LineNumber,
		x.Message)
}

type IniOptions uint

const (
	IniNone            IniOptions = 0
	IniIncludeDefaults            = 1 << iota
	IniIncludeComments
	IniDefault = IniIncludeComments
)

func optionIniName(option *Option) string {
	name := option.tag.Get("_read-ini-name")

	if len(name) != 0 {
		return name
	}

	name = option.tag.Get("ini-name")

	if len(name) != 0 {
		return name
	}

	return option.field.Name
}

func writeGroupIni(group *Group, namespace string, writer io.Writer, options IniOptions) {
	var sname string

	if len(namespace) != 0 {
		sname = namespace + "." + group.ShortDescription
	} else {
		sname = group.ShortDescription
	}

	fmt.Fprintf(writer, "[%s]\n", sname)

	for _, option := range group.options {
		if option.isFunc() {
			continue
		}

		if len(option.tag.Get("no-ini")) != 0 {
			continue
		}

		val := option.value

		if (options&IniIncludeDefaults) == IniNone &&
			reflect.DeepEqual(val, option.defaultValue) {
			continue
		}

		if (options & IniIncludeComments) != IniNone {
			fmt.Fprintf(writer, "; %s\n", option.Description)
		}

		switch val.Type().Kind() {
		case reflect.Slice:
			for idx := 0; idx < val.Len(); idx++ {
				v, _ := convertToString(val.Index(idx), option.tag)
				fmt.Fprintf(writer, "%s = %s\n", optionIniName(option), v)
			}
		case reflect.Map:
			for _, key := range val.MapKeys() {
				k, _ := convertToString(key, option.tag)
				v, _ := convertToString(val.MapIndex(key), option.tag)

				fmt.Fprintf(writer, "%s = %s:%s\n", optionIniName(option), k, v)
			}
		default:
			v, _ := convertToString(val, option.tag)
			fmt.Fprintf(writer, "%s = %s\n", optionIniName(option), v)
		}
	}

	fmt.Fprintln(writer)
}

func writeCommandIni(command *Command, namespace string, writer io.Writer, options IniOptions) {
	command.eachGroup(func(group *Group) {
		writeGroupIni(group, namespace, writer, options)
	}, true)

	for _, c := range command.commands {
		var nns string

		if len(namespace) != 0 {
			nns = c.Name + "." + nns
		} else {
			nns = c.Name
		}

		writeCommandIni(c, nns, writer, options)
	}
}

func writeIni(parser *Parser, writer io.Writer, options IniOptions) {
	writeCommandIni(parser.Command, "", writer, options)
}

func readIniFromFile(filename string) (Ini, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	return readIni(file, filename)
}

func readIni(contents io.Reader, filename string) (Ini, error) {
	ret := make(Ini)

	reader := bufio.NewReader(contents)

	var section IniSection
	var sectionname string

	var lineno uint

	for {
		line, err := readFullLine(reader)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		lineno++
		line = strings.TrimSpace(line)

		// Skip empty lines and lines starting with ; (comments)
		if len(line) == 0 || line[0] == ';' {
			continue
		}

		if section == nil {
			if line[0] != '[' || line[len(line)-1] != ']' {
				return nil, &IniError{
					Message:    "malformed section header",
					File:       filename,
					LineNumber: lineno,
				}
			}

			name := strings.TrimSpace(line[1 : len(line)-1])

			if len(name) == 0 {
				return nil, &IniError{
					Message:    "empty section name",
					File:       filename,
					LineNumber: lineno,
				}
			}

			sectionname = name
			section = ret[name]

			if section == nil {
				section = make(IniSection, 0, 10)
				ret[name] = section
			}

			continue
		}

		// Parse option here
		keyval := strings.SplitN(line, "=", 2)

		if len(keyval) != 2 {
			return nil, &IniError{
				Message:    "malformed key=value",
				File:       filename,
				LineNumber: lineno,
			}
		}

		name := strings.TrimSpace(keyval[0])
		value := strings.TrimSpace(keyval[1])

		section = append(section, IniValue{
			Name:  name,
			Value: value,
		})

		ret[sectionname] = section
	}

	return ret, nil
}

func (p *Parser) parseIni(ini Ini) error {
	for name, section := range ini {
		group := p.groupByName(name)

		if group == nil {
			return newError(ErrUnknownGroup,
				fmt.Sprintf("could not find option group `%s'", name))
		}

		for _, inival := range section {
			opt := group.optionByName(inival.Name, func(o *Option, n string) bool {
				return strings.ToLower(o.tag.Get("ini-name")) == strings.ToLower(n)
			})

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

			pval := &inival.Value

			if opt.isBool() && len(inival.Value) == 0 {
				pval = nil
			}

			if err := opt.set(pval); err != nil {
				return wrapError(err)
			}

			opt.tag.Set("_read-ini-name", inival.Name)
		}
	}

	return nil
}
