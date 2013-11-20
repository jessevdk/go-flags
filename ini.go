package flags

import (
	"io"
	"os"
)

// ParseIni is a convenience function to parse command line options with default
// settings from an ini file. The provided data is a pointer to a struct
// representing the default option group (named "Application Options"). For
// more control, use flags.NewParser.
func ParseIni(filename string, data interface{}) error {
	return NewParser(data, Default).ParseIniFile(filename)
}

// ParseIniFile parses flags from an ini formatted file. See ParseIni for more
// information on the ini file foramt. The returned errors can be of the type
// flags.Error or flags.IniError.
func (p *Parser) ParseIniFile(filename string) error {
	p.storeDefaults()

	ini, err := readIniFromFile(filename)

	if err != nil {
		return err
	}

	return p.parseIni(ini)
}

// ParseIni parses flags from an ini format. You can use ParseIniFile as a
// convenience function to parse from a filename instead of a general
// io.Reader.
//
// The format of the ini file is as follows:
//
// [Option group name]
// option = value
//
// Each section in the ini file represents an option group or command in the
// flags parser. The default flags parser option group (i.e. when using
// flags.Parse) is named 'Application Options'. The ini option name is matched
// in the following order:
//
// 1. Compared to the ini-name tag on the option struct field (if present)
// 2. Compared to the struct field name
// 3. Compared to the option long name (if present)
// 4. Compared to the option short name (if present)
//
// Sections for nested groups and commands can be addressed using a dot `.'
// namespacing notation (i.e [subcommand.Options]). Group section names are
// matched case insensitive.
//
// The returned errors can be of the type flags.Error or
// flags.IniError.
func (p *Parser) ParseIni(reader io.Reader) error {
	p.storeDefaults()

	ini, err := readIni(reader, "")

	if err != nil {
		return err
	}

	return p.parseIni(ini)
}

// WriteIniToFile writes the flags as ini format into a file. See WriteIni
// for more information. The returned error occurs when the specified file
// could not be opened for writing.
func (p *Parser) WriteIniToFile(filename string, options IniOptions) error {
	file, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer file.Close()
	p.WriteIni(file, options)

	return nil
}

// WriteIni writes the current values of all the flags to an ini format.
// See ParseIni for more information on the ini file format. You typically
// call this only after settings have been parsed since the default values of each
// option are stored just before parsing the flags (this is only relevant when
// IniIncludeDefaults is _not_ set in options).
func (p *Parser) WriteIni(writer io.Writer, options IniOptions) {
	writeIni(p, writer, options)
}
