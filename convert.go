// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

import (
	"encoding"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Marshaler is the interface implemented by types that can marshal themselves
// to a string representation of the flag.
type Marshaler interface {
	// MarshalFlag marshals a flag value to its string representation.
	MarshalFlag() (string, error)
}

// Unmarshaler is the interface implemented by types that can unmarshal a flag
// argument to themselves. The provided value is directly passed from the
// command line.
type Unmarshaler interface {
	// UnmarshalFlag unmarshals a string value representation to the flag
	// value (which therefore needs to be a pointer receiver).
	UnmarshalFlag(value string) error
}

// ValueValidator is the interface implemented by types that can validate a
// flag argument themselves. The provided value is directly passed from the
// command line.
type ValueValidator interface {
	// IsValidValue returns an error if the provided string value is valid for
	// the flag.
	IsValidValue(value string) error
}

func getBase(options multiTag, base int) (int, error) {
	sbase := options.Get("base")

	var err error
	var ivbase int64

	if sbase != "" {
		ivbase, err = strconv.ParseInt(sbase, 10, 32)
		base = int(ivbase)
	}

	return base, err
}

// keyValueDelimiter returns the delimiter separating the key from the value
// for map typed options, as configured by the key-value-delimiter tag.
func keyValueDelimiter(options multiTag) string {
	if delim := options.Get("key-value-delimiter"); delim != "" {
		return delim
	}

	return ":"
}

func convertMarshal(val reflect.Value) (bool, string, error) {
	// Check first for the Marshaler interface
	if val.IsValid() && val.Type().NumMethod() > 0 && val.CanInterface() {
		if marshaler, ok := val.Interface().(Marshaler); ok {
			ret, err := marshaler.MarshalFlag()
			return true, ret, err
		}

		if marshaler, ok := val.Interface().(encoding.TextMarshaler); ok {
			ret, err := marshaler.MarshalText()
			return true, string(ret), err
		}
	}

	if val.IsValid() && val.Type().Kind() != reflect.Pointer && val.CanAddr() {
		return convertMarshal(val.Addr())
	}

	return false, "", nil
}

func convertToString(val reflect.Value, options multiTag) (string, error) {
	if ok, ret, err := convertMarshal(val); ok {
		return ret, err
	}

	if !val.IsValid() {
		return "", nil
	}

	tp := val.Type()

	// Support for time.Duration
	if tp == reflect.TypeFor[time.Duration]() {
		stringer := val.Interface().(fmt.Stringer)
		return stringer.String(), nil
	}

	switch tp.Kind() {
	case reflect.String:
		return val.String(), nil
	case reflect.Bool:
		if val.Bool() {
			return "true", nil
		}

		return "false", nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		base, err := getBase(options, 10)

		if err != nil {
			return "", err
		}

		return strconv.FormatInt(val.Int(), base), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		base, err := getBase(options, 10)

		if err != nil {
			return "", err
		}

		return strconv.FormatUint(val.Uint(), base), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'g', -1, tp.Bits()), nil
	case reflect.Slice:
		if val.Len() == 0 {
			return "", nil
		}

		var ret strings.Builder

		ret.WriteString("[")

		for i := 0; i < val.Len(); i++ {
			if i != 0 {
				ret.WriteString(", ")
			}

			item, err := convertToString(val.Index(i), options)

			if err != nil {
				return "", err
			}

			ret.WriteString(item)
		}

		ret.WriteString("]")

		return ret.String(), nil
	case reflect.Map:
		delim := keyValueDelimiter(options)

		// Map iteration order is not stable, so the entries are sorted by
		// their stringified key to keep the result reproducible.
		type entry struct {
			key string
			str string
		}

		entries := make([]entry, 0, val.Len())

		for _, key := range val.MapKeys() {
			keyitem, err := convertToString(key, options)

			if err != nil {
				return "", err
			}

			item, err := convertToString(val.MapIndex(key), options)

			if err != nil {
				return "", err
			}

			entries = append(entries, entry{key: keyitem, str: keyitem + delim + item})
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].key < entries[j].key
		})

		strs := make([]string, len(entries))

		for i, e := range entries {
			strs[i] = e.str
		}

		return "{" + strings.Join(strs, ", ") + "}", nil
	case reflect.Pointer:
		return convertToString(reflect.Indirect(val), options)
	case reflect.Interface:
		if !val.IsNil() {
			return convertToString(val.Elem(), options)
		}
	}

	return "", nil
}

func convertUnmarshal(val string, retval reflect.Value) (bool, error) {
	if retval.Type().NumMethod() > 0 && retval.CanInterface() {
		if unmarshaler, ok := retval.Interface().(Unmarshaler); ok {
			if retval.Kind() == reflect.Pointer && retval.IsNil() {
				retval.Set(reflect.New(retval.Type().Elem()))

				// Re-assign from the new value
				unmarshaler = retval.Interface().(Unmarshaler)
			}

			return true, unmarshaler.UnmarshalFlag(val)
		}

		if unmarshaler, ok := retval.Interface().(encoding.TextUnmarshaler); ok {
			if retval.Kind() == reflect.Pointer && retval.IsNil() {
				retval.Set(reflect.New(retval.Type().Elem()))

				// Re-assign from the new value
				unmarshaler = retval.Interface().(encoding.TextUnmarshaler)
			}

			return true, unmarshaler.UnmarshalText([]byte(val))
		}
	}

	if retval.Type().Kind() != reflect.Pointer && retval.CanAddr() {
		return convertUnmarshal(val, retval.Addr())
	}

	if retval.Type().Kind() == reflect.Interface && !retval.IsNil() {
		return convertUnmarshal(val, retval.Elem())
	}

	return false, nil
}

func convert(val string, retval reflect.Value, options multiTag) error {
	if ok, err := convertUnmarshal(val, retval); ok {
		return err
	}

	tp := retval.Type()

	// Support for time.Duration
	if tp == reflect.TypeFor[time.Duration]() {
		parsed, err := time.ParseDuration(val)

		if err != nil {
			return err
		}

		retval.SetInt(int64(parsed))
		return nil
	}

	switch tp.Kind() {
	case reflect.String:
		retval.SetString(val)
	case reflect.Bool:
		if val == "" {
			retval.SetBool(true)
		} else {
			b, err := strconv.ParseBool(val)

			if err != nil {
				return err
			}

			retval.SetBool(b)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		base, err := getBase(options, 0)

		if err != nil {
			return err
		}

		parsed, err := strconv.ParseInt(val, base, tp.Bits())

		if err != nil {
			return err
		}

		retval.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		base, err := getBase(options, 0)

		if err != nil {
			return err
		}

		parsed, err := strconv.ParseUint(val, base, tp.Bits())

		if err != nil {
			return err
		}

		retval.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(val, tp.Bits())

		if err != nil {
			return err
		}

		retval.SetFloat(parsed)
	case reflect.Slice:
		elemtp := tp.Elem()

		elemvalptr := reflect.New(elemtp)
		elemval := reflect.Indirect(elemvalptr)

		if err := convert(val, elemval, options); err != nil {
			return err
		}

		retval.Set(reflect.Append(retval, elemval))
	case reflect.Map:
		parts := strings.SplitN(val, keyValueDelimiter(options), 2)

		key := parts[0]
		var value string

		if len(parts) == 2 {
			value = parts[1]
		}

		keytp := tp.Key()
		keyval := reflect.New(keytp)

		if err := convert(key, keyval, options); err != nil {
			return err
		}

		valuetp := tp.Elem()
		valueval := reflect.New(valuetp)

		if err := convert(value, valueval, options); err != nil {
			return err
		}

		if retval.IsNil() {
			retval.Set(reflect.MakeMap(tp))
		}

		retval.SetMapIndex(reflect.Indirect(keyval), reflect.Indirect(valueval))
	case reflect.Pointer:
		if retval.IsNil() {
			retval.Set(reflect.New(retval.Type().Elem()))
		}

		return convert(val, reflect.Indirect(retval), options)
	case reflect.Interface:
		if !retval.IsNil() {
			return convert(val, retval.Elem(), options)
		}
	}

	return nil
}

func isPrint(s string) bool {
	for _, c := range s {
		if !strconv.IsPrint(c) {
			return false
		}
	}

	return true
}

func quoteIfNeeded(s string) string {
	if !isPrint(s) {
		return strconv.Quote(s)
	}

	return s
}

func quoteV(s []string) []string {
	ret := make([]string, len(s))

	for i, v := range s {
		ret[i] = strconv.Quote(v)
	}

	return ret
}

func unquoteIfPossible(s string) (string, error) {
	if len(s) == 0 || s[0] != '"' {
		return s, nil
	}

	return strconv.Unquote(s)
}
