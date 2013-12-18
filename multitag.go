package flags

import (
	"strings"
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

func (x *multiTag) scan() (map[string][]string, error) {
	v := x.value

	ret := make(map[string][]string)

	// This is mostly copied from reflect.StructTag.Get
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

		if v[i] != ':' {
			return nil, newErrorf(ErrTag, "Expected `:' after key name, but got %v (in %v column %v)", v[i], v, i)
		}

		if i+1 >= len(v) {
			return nil, newErrorf(ErrTag, "Expected `\"' to start tag value at end of tag (in %v)", v)
		}

		if v[i+1] != '"' {
			return nil, newErrorf(ErrTag, "Expected `\"' to start tag value, but got %v (in %v column %v)", v[i+1], v, i+1)
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
			return nil, newErrorf(ErrTag, "Expected end of tag value `\"' at end of tag (in %v)", v)
		}

		fixed := strings.Replace(v[:i+1], "\n", "", -1)
		val, _ := strconv.Unquote(fixed)
		v = v[i+1:]

		ret[name] = append(ret[name], val)
	}

	return ret, nil
}

func (x *multiTag) Parse() error {
	vals, err := x.scan()
	x.cache = vals

	return err
}

func (x *multiTag) cached() map[string][]string {
	if x.cache == nil {
		cache, _ := x.scan()

		if cache == nil {
			cache = make(map[string][]string)
		}

		x.cache = cache
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
