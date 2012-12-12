// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

import (
	"errors"
)

// The provided container is not a pointer to a struct
var ErrNotPointerToStruct = errors.New("provided data is not a pointer to struct")

// The provided short name is longer than a single character
var ErrShortNameTooLong = errors.New("short names can only be 1 character")

// An option group. The option group has a name and a set of options.
type Group struct {
	Commander

	// The name of the group.
	Name string

	Names    map[string]*Option
	IniNames map[string]*Option

	// A map of long names to option option descriptions.
	LongNames map[string]*Option

	// A map of short names to option option descriptions.
	ShortNames map[rune]*Option

	// A list of all the options in the group.
	Options []*Option

	// An error which occurred when creating the group.
	Error error

	// Groups embedded in this group
	EmbeddedGroups []*Group

	IsCommand       bool
	LongDescription string

	data interface{}
}

type Command interface {
	Execute(args []string) error
}

// NewGroup creates a new option group with a given name and underlying data
// container. The data container is a pointer to a struct. The fields of the
// struct represent the command line options (using field tags) and their values
// will be set when their corresponding options appear in the command line
// arguments.
func NewGroup(name string, data interface{}) *Group {
	ret := &Group{
		Commander: Commander{
			Commands: make(map[string]*Group),
		},

		Name:       name,
		Names:      make(map[string]*Option),
		IniNames:   make(map[string]*Option),
		LongNames:  make(map[string]*Option),
		ShortNames: make(map[rune]*Option),
		IsCommand:  false,
		data:       data,
	}

	ret.Error = ret.scan()
	return ret
}
