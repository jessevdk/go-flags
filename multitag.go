package flags

import (
	"strconv"
)

type multiTag struct {
	value string
	cache map[string][]string
}

func newMultiTag(v string) multiTag {
	return multiTag{
		value: v,
	}
}

func (x *multiTag) scan() map[string][]string {
	v := x.value

	ret := make(map[string][]string)

	// This is mostly copied from reflect.StructTAg.Get
	for v != "" {
		i := 0

		// Skip whitespace
		for i < len(v) && v[i] == ' ' {
			i++
		}

		v = v[i:]

		if v == "" {
			break
		}

		// Scan to colon to find key
		i = 0

		for i < len(v) && v[i] != ' ' && v[i] != ':' && v[i] != '"' {
			i++
		}

		if i+1 >= len(v) || v[i] != ':' || v[i+1] != '"' {
			break
		}

		name := v[:i]
		v = v[i+1:]

		// Scan quoted string to find value
		i = 1

		for i < len(v) && v[i] != '"' {
			if v[i] == '\\' {
				i++
			}
			i++
		}

		if i >= len(v) {
			break
		}

		val, _ := strconv.Unquote(v[:i+1])
		v = v[i+1:]

		ret[name] = append(ret[name], val)
	}

	return ret
}

func (x *multiTag) cached() map[string][]string {
	if x.cache == nil {
		x.cache = x.scan()
	}

	return x.cache
}

func (x *multiTag) Get(key string) string {
	c := x.cached()

	if v, ok := c[key]; ok {
		return v[len(v)-1]
	}

	return ""
}

func (x *multiTag) GetMany(key string) []string {
	c := x.cached()
	return c[key]
}

func (x *multiTag) Set(key string, value string) {
	c := x.cached()
	c[key] = []string{value}
}

func (x *multiTag) SetMany(key string, value []string) {
	c := x.cached()
	c[key] = value
}
