package flags

import (
	"reflect"
)

// Arg represents a positional argument on the command line.
type Arg struct {
	// The name of the positional argument (used in the help)
	Name string

	// A description of the positional argument (used in the help)
	Description string

	// The minimal number of required positional arguments
	Required int

	// The maximum number of required positional arguments
	RequiredMaximum int

	value reflect.Value
	tag   multiTag
}

func (a *Arg) isRemaining() bool {
	return a.value.Type().Kind() == reflect.Slice
}

// isValidValue checks arg with the ValueValidator implementation of the
// positional argument type (the element type for remaining arguments), if
// there is one.
func (a *Arg) isValidValue(arg string) error {
	tp := a.value.Type()

	if tp.Kind() == reflect.Slice {
		tp = tp.Elem()
	}

	// Validate against a fresh value so that slices, nil pointers and
	// already populated arguments are all handled the same way.
	v := reflect.New(tp).Elem()

	if v.Kind() == reflect.Pointer {
		v.Set(reflect.New(tp.Elem()))
	}

	if validator := valueValidator(v); validator != nil {
		return validator.IsValidValue(arg)
	}

	return nil
}
