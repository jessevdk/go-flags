package flags

import (
	"reflect"
	"errors"
	"unicode/utf8"
	"fmt"
)

var _ = fmt.Printf

var ErrNotPointer = errors.New("provided data is not a pointer")
var ErrNotPointerToStruct = errors.New("provided data is not a pointer to struct")
var ErrShortNameTooLong = errors.New("short names can only be 1 character")

type Info struct {
	ShortName rune
	LongName string
	Description string
	Default string
	OptionalArgument bool

	Value reflect.Value
	Options reflect.StructTag
}

type Group struct {
	Name string
	LongNames map[string]*Info
	ShortNames map[rune]*Info
	Options []*Info

	Data interface{}
	Error error
}

func (info *Info) IsBool() bool {
	tp := info.Value.Type()

	switch tp.Kind() {
	case reflect.Bool:
		return true
	case reflect.Slice:
		return (tp.Elem().Kind() == reflect.Bool)
	}

	return false
}

func (info *Info) Set(value string) error {
	return convert(value, info.Value, info.Options)
}

func (info *Info) String() string {
	var s string
	var short string

	if info.ShortName != 0 {
		data := make([]byte, utf8.RuneLen(info.ShortName))
		utf8.EncodeRune(data, info.ShortName)
		short = string(data)

		if len(info.LongName) != 0 {
			s = fmt.Sprintf("-%s, --%s", short, info.LongName)
		} else {
			s = fmt.Sprintf("-%s", short)
		}
	} else if len(info.LongName) != 0 {
		s = fmt.Sprintf("--%s", info.LongName)
	}

	if len(info.Description) != 0 {
		return fmt.Sprintf("%s (%s)", s, info.Description)
	}

	return s
}

func NewGroup(name string, s interface{}) *Group {
	ret := &Group {
		Name: name,
		LongNames: make(map[string]*Info),
		ShortNames: make(map[rune]*Info),
		Data: s,
	}

	ret.Error = ret.scan()
	return ret
}

func (g *Group) scan() error {
	// Get all the public fields in the data struct
	ptrval := reflect.ValueOf(g.Data)

	if ptrval.Type().Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	stype := ptrval.Type().Elem()

	if stype.Kind() != reflect.Struct {
		return ErrNotPointerToStruct
	}

	realval := reflect.Indirect(ptrval)

	for i := 0; i < stype.NumField(); i++ {
		field := stype.Field(i)

		// PkgName is set only for non-exported fields, which we ignore
		if field.PkgPath != "" {
			continue
		}

		// Skip anonymous fields
		if field.Anonymous {
			continue
		}

		// Skip fields with the no-flag tag
		if field.Tag.Get("no-flag") != "" {
			continue
		}

		longname := field.Tag.Get("long")
		shortname := field.Tag.Get("short")

		if longname == "" && shortname == "" {
			continue
		}

		short := rune(0)
		rc := utf8.RuneCountInString(shortname)

		if rc > 1 {
			return ErrShortNameTooLong
		} else if rc == 1 {
			short, _ = utf8.DecodeRuneInString(shortname)
		}

		description := field.Tag.Get("description")
		def := field.Tag.Get("default")

		optional := (field.Tag.Get("optional") != "")

		info := &Info {
			Description: description,
			ShortName: short,
			LongName: longname,
			Default: def,
			OptionalArgument: optional,
			Value: realval.Field(i),
			Options: field.Tag,
		}

		g.Options = append(g.Options, info)

		if info.ShortName != 0 {
			g.ShortNames[info.ShortName] = info
		}

		if info.LongName != "" {
			g.LongNames[info.LongName] = info
		}
	}

	return nil
}
