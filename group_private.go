package flags

import (
	"reflect"
	"strings"
	"unicode/utf8"
)

func (option *Option) canArgument() bool {
	if option.isBool() {
		return false
	}

	if option.isFunc() {
		return (option.value.Type().NumIn() > 0)
	}

	return true
}

func (option *Option) isBool() bool {
	tp := option.value.Type()

	switch tp.Kind() {
	case reflect.Bool:
		return true
	case reflect.Slice:
		return (tp.Elem().Kind() == reflect.Bool)
	}

	return false
}

func (option *Option) isFunc() bool {
	return option.value.Type().Kind() == reflect.Func
}

func (option *Option) call(value *string) error {
	var retval []reflect.Value

	if value == nil {
		retval = option.value.Call(nil)
	} else {
		tp := option.value.Type().In(0)

		val := reflect.New(tp)
		val = reflect.Indirect(val)

		if err := convert(*value, val, option.options); err != nil {
			return err
		}

		retval = option.value.Call([]reflect.Value{val})
	}

	if len(retval) == 1 && retval[0].Type() == reflect.TypeOf((*error)(nil)).Elem() {
		return retval[0].Interface().(error)
	}

	return nil
}

func (g *Group) lookupByName(name string, ini bool) *Option {
	name = strings.ToLower(name)

	if ini {
		if ret := g.IniNames[name]; ret != nil {
			return ret
		}

		if ret := g.Names[name]; ret != nil {
			return ret
		}
	}

	if ret := g.LongNames[name]; ret != nil {
		return ret
	}

	if utf8.RuneCountInString(name) == 1 {
		r, _ := utf8.DecodeRuneInString(name)

		if ret := g.ShortNames[r]; ret != nil {
			return ret
		}
	}

	return nil
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

		option := &Option{
			Description:      description,
			ShortName:        short,
			LongName:         longname,
			Default:          def,
			OptionalArgument: optional,
			value:            realval.Field(i),
			options:          field.Tag,
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
