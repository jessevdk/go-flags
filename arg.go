package flags

import (
	"reflect"
)

type Arg struct {
	Name        string
	Description string

	value reflect.Value
	tag   multiTag
}

func (a *Arg) isRemaining() bool {
	return a.value.Type().Kind() == reflect.Slice
}
