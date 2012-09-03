package flags

import (
	"reflect"
	"unicode/utf8"
)

func (info *Info) canArgument() bool {
	if info.isBool() {
		return false
	}

	if info.isFunc() {
		return (info.value.Type().NumIn() > 0)
	}

	return true
}

func (info *Info) isBool() bool {
	tp := info.value.Type()

	switch tp.Kind() {
	case reflect.Bool:
		return true
	case reflect.Slice:
		return (tp.Elem().Kind() == reflect.Bool)
	}

	return false
}

func (info *Info) isFunc() bool {
	return info.value.Type().Kind() == reflect.Func
}

func (info *Info) call(value *string) error {
	var retval []reflect.Value

	if value == nil {
		retval = info.value.Call(nil)
	} else {
		tp := info.value.Type().In(0)

		val := reflect.New(tp)
		val = reflect.Indirect(val)

		if err := convert(*value, val, info.options); err != nil {
			return err
		}

		retval = info.value.Call([]reflect.Value{val})
	}

	if len(retval) == 1 && retval[0].Type() == reflect.TypeOf((*error)(nil)).Elem() {
		return retval[0].Interface().(error)
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

		info := &Info{
			Description:      description,
			ShortName:        short,
			LongName:         longname,
			Default:          def,
			OptionalArgument: optional,
			value:            realval.Field(i),
			options:          field.Tag,
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
