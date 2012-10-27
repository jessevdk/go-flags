package flags

import (
	"reflect"
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

func (option *Option) iniName() string {
	if len(option.iniUsedName) != 0 {
		return option.iniUsedName
	}

	name := option.options.Get("ini-name")

	if len(name) != 0 {
		return name
	}

	return option.fieldName
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
		if retval[0].Interface() == nil {
			return nil
		}

		return retval[0].Interface().(error)
	}

	return nil
}
