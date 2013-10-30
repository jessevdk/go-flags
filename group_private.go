package flags

import (
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
	"unsafe"
)

func (g *Group) lookupByName(name string, ini bool) (*Option, string) {
	name = strings.ToLower(name)

	if ini {
		if ret := g.IniNames[name]; ret != nil {
			return ret, ret.tag.Get("ini-name")
		}

		if ret := g.Names[name]; ret != nil {
			return ret, ret.Field.Name
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
		// First. empty out the value
		if len(option.Default) > 0 {
			option.clear()
		}

		for _, d := range option.Default {
			option.Set(&d)
		}

		if !option.Value.CanSet() {
			continue
		}

		option.defaultValue = reflect.ValueOf(option.Value.Interface())
	}
}

func (g *Group) scanStruct(realval reflect.Value, sfield *reflect.StructField) error {
	stype := realval.Type()

	if sfield != nil {
		mtag := newMultiTag(string(sfield.Tag))

		groupName := mtag.Get("group")
		commandName := mtag.Get("command")
		name := mtag.Get("name")
		description := mtag.Get("description")

		iscommand := false

		if len(commandName) != 0 {
			iscommand = true

			if len(name) != 0 {
				groupName = name
			} else if len(commandName) != 0 {
				groupName = commandName
			}
		}

		if len(groupName) != 0 {
			ptrval := reflect.NewAt(realval.Type(), unsafe.Pointer(realval.UnsafeAddr()))
			newGroup := NewGroup(groupName, ptrval.Interface())

			if iscommand {
				newGroup.IsCommand = true
				g.Commands[commandName] = newGroup
			}

			newGroup.LongDescription = description

			g.EmbeddedGroups = append(g.EmbeddedGroups, newGroup)
			return g.Error
		}
	}

	for i := 0; i < stype.NumField(); i++ {
		field := stype.Field(i)

		// PkgName is set only for non-exported fields, which we ignore
		if field.PkgPath != "" {
			continue
		}

		mtag := newMultiTag(string(field.Tag))

		// Skip fields with the no-flag tag
		if mtag.Get("no-flag") != "" {
			continue
		}

		// Dive deep into structs or pointers to structs
		kind := field.Type.Kind()

		if kind == reflect.Struct {
			err := g.scanStruct(realval.Field(i), &field)

			if err != nil {
				return err
			}

		} else if kind == reflect.Ptr &&
			field.Type.Elem().Kind() == reflect.Struct &&
			!realval.Field(i).IsNil() {
			err := g.scanStruct(reflect.Indirect(realval.Field(i)),
				&field)

			if err != nil {
				return err
			}
		}

		longname := mtag.Get("long")
		shortname := mtag.Get("short")

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

		description := mtag.Get("description")
		def := mtag.GetMany("default")
		optionalValue := mtag.GetMany("optional-value")
		valueName := mtag.Get("value-name")
		defaultMask := mtag.Get("default-mask")

		optional := (mtag.Get("optional") != "")
		required := (mtag.Get("required") != "")

		option := &Option{
			Description:      description,
			ShortName:        short,
			LongName:         longname,
			Default:          def,
			OptionalArgument: optional,
			OptionalValue:    optionalValue,
			Required:         required,
			Field:            field,
			Value:            realval.Field(i),
			ValueName:        valueName,
			tag:              mtag,
			defaultMask:      defaultMask,
		}

		g.Options = append(g.Options, option)

		if option.ShortName != 0 {
			g.ShortNames[option.ShortName] = option
		}

		if option.LongName != "" {
			g.LongNames[strings.ToLower(option.LongName)] = option
		}

		g.Names[strings.ToLower(field.Name)] = option

		ininame := mtag.Get("ini-name")

		if len(ininame) != 0 {
			g.IniNames[strings.ToLower(ininame)] = option
		}
	}

	return nil
}

func (g *Group) each(index int, cb func(int, *Group)) int {
	cb(index, g)
	index++

	for _, group := range g.EmbeddedGroups {
		group.each(index, cb)
		index++
	}

	return index
}

func (g *Group) eachCommand(cb func(string, *Group)) {
	cmds := make([]string, len(g.Commands))
	i := 0

	for k, _ := range g.Commands {
		cmds[i] = k
	}

	sort.Strings(cmds)

	for _, k := range cmds {
		cb(k, g.Commands[k])
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
	return g.scanStruct(realval, nil)
}
