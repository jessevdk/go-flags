package flags

import (
	"sort"
)

type Commander struct {
	Commands map[string]*Group
}

func (x *Commander) sortedNames() []string {
	ret := make([]string, 0, len(x.Commands))

	for k, _ := range x.Commands {
		ret = append(ret, k)
	}

	sort.Strings(ret)
	return ret
}

func (x *Commander) EachCommand(cb func(command string, grp *Group)) {
	for k, v := range x.Commands {
		cb(k, v)

		// Recurse
		v.Commander.EachCommand(cb)
	}
}
