package flags

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

type IniSection map[string]string
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

func writeIni(parser *Parser, writer io.Writer, options IniOptions) {
	parser.EachGroup(func(i int, group *Group) {
		if i != 0 {
			io.WriteString(writer, "\n")
		}

		fmt.Fprintf(writer, "[%s]\n", group.Name)

		for _, option := range group.Options {
			if option.isFunc() {
				continue
			}

			if len(option.tag.Get("no-ini")) != 0 {
				continue
			}

			val := option.Value

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
					fmt.Fprintf(writer,
						"%s = %s\n",
						option.iniName(),
						convertToString(val.Index(idx),
							option.tag))
				}
			case reflect.Map:
				for _, key := range val.MapKeys() {
					fmt.Fprintf(writer,
						"%s = %s:%s\n",
						option.iniName(),
						convertToString(key,
							option.tag),
						convertToString(val.MapIndex(key),
							option.tag))
				}
			default:
				fmt.Fprintf(writer,
					"%s = %s\n",
					option.iniName(),
					convertToString(val,
						option.tag))
			}
		}
	})
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

			section = ret[name]

			if section == nil {
				section = make(IniSection)
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

		section[strings.TrimSpace(keyval[0])] = strings.TrimSpace(keyval[1])
	}

	return ret, nil
}
