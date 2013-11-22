package flags_test

import (
	"github.com/jessevdk/go-flags"
	"testing"
	"bytes"
	"strings"
)

func TestWriteIni(t *testing.T) {
	var opts helpOptions

	p := flags.NewNamedParser("TestIni", flags.Default)
	p.AddGroup("Application Options", "The application options", &opts)

	p.ParseArgs([]string{"-vv", "--intmap=a:2", "--intmap", "b:3"})

	inip := flags.NewIniParser(p)

	var b bytes.Buffer
	inip.Write(&b, flags.IniDefault | flags.IniIncludeDefaults)

	got := b.String()
	expected := `[Application Options]
; Show verbose debug information
verbose = true
verbose = true

; A slice of pointers to string
; PtrSlice =

[Other Options]
; A slice of strings
; StringSlice =

; A map from string to int
int-map = a:2
int-map = b:3

`

	if got != expected {
		ret, err := helpDiff(got, expected)

		if err != nil {
			t.Errorf("Unexpected ini, expected:\n\n%s\n\nbut got\n\n%s", expected, got)
		} else {
			t.Errorf("Unexpected ini:\n\n%s", ret)
		}
	}
}

func TestReadIni(t *testing.T) {
	var opts helpOptions

	p := flags.NewNamedParser("TestIni", flags.Default)
	p.AddGroup("Application Options", "The application options", &opts)

	inip := flags.NewIniParser(p)

inic := `[Application Options]
; Show verbose debug information
verbose = true
verbose = true

; A slice of pointers to string
; PtrSlice =

[Other Options]
; A slice of strings
; StringSlice =

; A map from string to int
int-map = a:2
int-map = b:3

`

	b := strings.NewReader(inic)
	err := inip.Parse(b)

	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	assertBoolArray(t, opts.Verbose, []bool{true, true})

	if v, ok := opts.Other.IntMap["a"]; !ok {
		t.Errorf("Expected \"a\" in Other.IntMap")
	} else if v != 2 {
		t.Errorf("Expected Other.IntMap[\"a\"] = 2, but got %v", v)
	}

	if v, ok := opts.Other.IntMap["b"]; !ok {
		t.Errorf("Expected \"b\" in Other.IntMap")
	} else if v != 3 {
		t.Errorf("Expected Other.IntMap[\"b\"] = 3, but got %v", v)
	}
}
