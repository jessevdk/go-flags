// Copyright 2012 Jesse van den Kieboom. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

import (
	"errors"
	"strings"
)

// The provided container is not a pointer to a struct
var ErrNotPointerToStruct = errors.New("provided data is not a pointer to struct")

// An option group. The option group has a name and a set of options.
type Group struct {
	// A short description of the group
	ShortDescription string

	// A long description of the group
	LongDescription string

	// All the options in the group
	options []*Option

	// All the subgroups
	groups []*Group

	data interface{}
}

// AddGroup adds a new group to the command with the given name and data. The
// data needs to be a pointer to a struct from which the fields indicate which
// options are in the group.
func (g *Group) AddGroup(shortDescription string, longDescription string, data interface{}) (*Group, error) {
	group := newGroup(shortDescription, longDescription, data)

	if err := group.scan(); err != nil {
		return nil, err
	}

	g.groups = append(g.groups, group)
	return group, nil
}

// Get a list of subgroups in the group.
func (g *Group) Groups() []*Group {
	return g.groups
}

// Get a list of options in the group.
func (g *Group) Options() []*Option {
	return g.options
}

// Group locates the subgroup with the given short description and returns it.
// If no such group can be found Find will return nil. Note that the description
// is matched case insensitively.
func (g *Group) Find(shortDescription string) *Group {
	lshortDescription := strings.ToLower(shortDescription)

	var ret *Group

	g.eachGroup(func(gg *Group) {
		if gg != g && strings.ToLower(gg.ShortDescription) == lshortDescription {
			ret = gg
		}
	})

	return ret
}
