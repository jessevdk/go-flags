package flags

import (
	"reflect"
	"strings"
	"unicode/utf8"
	"unsafe"
)

func (g *Group) lookupByName(name string, ini bool) (*Option, string) {
	name = strings.ToLower(name)

	if ini {
		if ret := g.IniNames[name]; ret != nil {
			return ret, ret.options.Get("ini-name")
		}

		if ret := g.Names[name]; ret != nil {
			return ret, ret.fieldName
		}
	}

	if ret := g.LongNames[name]; ret != nil {
		return ret, ret.LongName
	}

	if utf8.RuneCountInString(name) == 1 {
		r, _ := utf8.DecodeRuneInString(name)

		if ret := g.ShortNames[r]; ret != nil {
			data := make([]byte, utf8.RuneLen(ret.ShortName))
			utf8.EncodeRune(data, ret.ShortName)

			return ret, string(data)
		}
	}

	return nil, ""
}

func (g *Group) storeDefaults() {
	for _, option := range g.Options {
		if !option.value.CanAddr() {
			continue
		}

		addr := option.value.UnsafeAddr()

		// Create a pointer to the current value
		ptr := reflect.NewAt(option.value.Type(), unsafe.Pointer(addr))

		// Indirect the pointer to the value to make a copy
		option.defaultValue = reflect.Indirect(ptr)
	}
}

func (g *Group) scan() error {
	// Get all the public fields in the data struct
	ptrval := reflect.ValueOf(g.data)

	if ptrval.Type().Kind() != reflect.Ptr {
		panic(ErrNotPointerToStruct)
	}

	stype := ptrval.Type().Elem()

	if stype.Kind() != reflect.Struct {
		panic(ErrNotPointerToStruct)
	}

	realval := reflect.Indirect(ptrval)

	for i := 0; i < stype.NumField(); i++ {
		field := stype.Field(i)

		// PkgName is set only for non-exported fields, which we ignore
		if field.PkgPath != "" {
			continue
		}

		// Skip anonymous fields.
		// TODO: would be useful to support
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
		required := (field.Tag.Get("required") != "")

		option := &Option{
			Description:      description,
			ShortName:        short,
			LongName:         longname,
			Default:          def,
			OptionalArgument: optional,
			Required:         required,
			value:            realval.Field(i),
			options:          field.Tag,
			fieldName:        field.Name,
		}

		g.Options = append(g.Options, option)

		if option.ShortName != 0 {
			g.ShortNames[option.ShortName] = option
		}

		if option.LongName != "" {
			g.LongNames[strings.ToLower(option.LongName)] = option
		}

		g.Names[strings.ToLower(field.Name)] = option

		ininame := field.Tag.Get("ini-name")

		if len(ininame) != 0 {
			g.IniNames[strings.ToLower(ininame)] = option
		}
	}

	return nil
}
