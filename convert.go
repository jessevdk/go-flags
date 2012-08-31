// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

import (
	"reflect"
	"strconv"
	"strings"
)

func getBase(options reflect.StructTag, base int) (int, error) {
	sbase := options.Get("base")

	var err error
	var ivbase int64

	if sbase != "" {
		ivbase, err = strconv.ParseInt(sbase, 10, 32)
		base = int(ivbase)
	}

	return base, err
}

func convert(val string, retval reflect.Value, options reflect.StructTag) error {
	tp := retval.Type()

	switch tp.Kind() {
	case reflect.String:
		retval.SetString(val)
	case reflect.Bool:
		retval.SetBool(true)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		base, err := getBase(options, 10)

		if err != nil {
			return err
		}

		parsed, err := strconv.ParseInt(val, base, tp.Bits())

		if err != nil {
			return err
		}

		retval.SetInt(parsed)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		base, err := getBase(options, 10)

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
		elemval := reflect.New(elemtp)

		if err := convert(val, elemval, options); err != nil {
			return err
		}

		retval.Set(reflect.Append(retval, reflect.Indirect(elemval)))
	case reflect.Map:
		parts := strings.SplitN(val, ":", 2)

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
	}

	return nil
}
