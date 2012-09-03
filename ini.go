package flags

import (
	"bufio"
	"fmt"
	"io"
	"os"
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

func readIni(filename string) (Ini, error) {
	ret := make(Ini)

	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	reader := bufio.NewReader(file)
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

			name := strings.TrimSpace(line[1 : len(line)-2])

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
