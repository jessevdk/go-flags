package flags

import (
	"fmt"
	"reflect"
	"unicode/utf8"
)

// Option flag information. Contains a description of the option, short and
// long name as well as a default value and whether an argument for this
// flag is optional.
type Option struct {
	// The short name of the option (a single character). If not 0, the
	// option flag can be 'activated' using -<ShortName>. Either ShortName
	// or LongName needs to be non-empty.
	ShortName rune

	// The long name of the option. If not "", the option flag can be
	// activated using --<LongName>. Either ShortName or LongName needs
	// to be non-empty.
	LongName string

	// The description of the option flag. This description is shown
	// automatically in the builtin help.
	Description string

	// The default value of the option.
	Default []string

	// If true, specifies that the argument to an option flag is optional.
	// When no argument to the flag is specified on the command line, the
	// value of Default will be set in the field this option represents.
	// This is only valid for non-boolean options.
	OptionalArgument bool

	// The optional value of the option. The optional value is used when
	// the option flag is marked as having an OptionalArgument. This means
	// that when the flag is specified, but no option argument is given,
	// the value of the field this option represents will be set to
	// OptionalValue. This is only valid for non-boolean options.
	OptionalValue []string

	// If true, the option _must_ be specified on the command line. If the
	// option is not specified, the parser will generate an ErrRequired type
	// error.
	Required bool

	// A name for the value of an option shown in the Help as --flag [ValueName]
	ValueName string

	// The struct field which the option represents.
	Field reflect.StructField

	// The struct field value which the option represents.
	Value reflect.Value

	defaultValue reflect.Value
	defaultMask  string
	iniUsedName  string
	tag          multiTag
}

// Set the value of an option to the specified value. An error will be returned
// if the specified value could not be converted to the corresponding option
// value type.
func (option *Option) Set(value *string) error {
	if option.isFunc() {
		return option.call(value)
	} else if value != nil {
		return convert(*value, option.Value, option.tag)
	} else {
		return convert("", option.Value, option.tag)
	}

	return nil
}

// Convert an option to a human friendly readable string describing the option.
func (option *Option) String() string {
	var s string
	var short string

	if option.ShortName != 0 {
		data := make([]byte, utf8.RuneLen(option.ShortName))
		utf8.EncodeRune(data, option.ShortName)
		short = string(data)

		if len(option.LongName) != 0 {
			s = fmt.Sprintf("-%s, --%s", short, option.LongName)
		} else {
			s = fmt.Sprintf("-%s", short)
		}
	} else if len(option.LongName) != 0 {
		s = fmt.Sprintf("--%s", option.LongName)
	}

	return s
}
