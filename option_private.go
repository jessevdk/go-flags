package flags

import (
	"reflect"
)

func (option *Option) canArgument() bool {
	return !option.isBool()
}

func (option *Option) clear() {
	tp := option.Value.Type()

	switch tp.Kind() {
	case reflect.Func:
		// Skip
	case reflect.Map:
		// Empty the map
		option.Value.Set(reflect.MakeMap(tp))
	default:
		zeroval := reflect.Zero(tp)
		option.Value.Set(zeroval)
	}
}

func (option *Option) isBool() bool {
	tp := option.Value.Type()

	switch tp.Kind() {
	case reflect.Bool:
		return true
	case reflect.Slice:
		return (tp.Elem().Kind() == reflect.Bool)
	case reflect.Func:
		return tp.NumIn() == 0
	}

	return false
}

func (option *Option) isFunc() bool {
	return option.Value.Type().Kind() == reflect.Func
}

func (option *Option) iniName() string {
	if len(option.iniUsedName) != 0 {
		return option.iniUsedName
	}

	name := option.tag.Get("ini-name")

	if len(name) != 0 {
		return name
	}

	return option.Field.Name
}

func (option *Option) call(value *string) error {
	var retval []reflect.Value

	if value == nil {
		retval = option.Value.Call(nil)
	} else {
		tp := option.Value.Type().In(0)

		val := reflect.New(tp)
		val = reflect.Indirect(val)

		if err := convert(*value, val, option.tag); err != nil {
			return err
		}

		retval = option.Value.Call([]reflect.Value{val})
	}

	if len(retval) == 1 && retval[0].Type() == reflect.TypeOf((*error)(nil)).Elem() {
		if retval[0].Interface() == nil {
			return nil
		}

		return retval[0].Interface().(error)
	}

	return nil
}
