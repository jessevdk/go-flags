package flags

import (
	"fmt"
	"path"
	"runtime"
	"testing"
)

func assertCallerInfo() (string, int) {
	ptr := make([]uintptr, 15)
	n := runtime.Callers(1, ptr)

	if n == 0 {
		return "", 0
	}

	mef := runtime.FuncForPC(ptr[0])
	mefile, meline := mef.FileLine(ptr[0])

	for i := 2; i < n; i++ {
		f := runtime.FuncForPC(ptr[i])
		file, line := f.FileLine(ptr[i])

		if file != mefile {
			return file, line
		}
	}

	return mefile, meline
}

func assertErrorf(t *testing.T, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	file, line := assertCallerInfo()

	t.Errorf("%s:%d: %s", path.Base(file), line, msg)
}

func assertFatalf(t *testing.T, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	file, line := assertCallerInfo()

	t.Fatalf("%s:%d: %s", path.Base(file), line, msg)
}

func assertString(t *testing.T, a string, b string) {
	if a != b {
		assertErrorf(t, "Expected %#v, but got %#v", b, a)
	}
}

func assertStringArray(t *testing.T, a []string, b []string) {
	if len(a) != len(b) {
		assertErrorf(t, "Expected %#v, but got %#v", b, a)
		return
	}

	for i, v := range a {
		if b[i] != v {
			assertErrorf(t, "Expected %#v, but got %#v", b, a)
			return
		}
	}
}

func assertBoolArray(t *testing.T, a []bool, b []bool) {
	if len(a) != len(b) {
		assertErrorf(t, "Expected %#v, but got %#v", b, a)
		return
	}

	for i, v := range a {
		if b[i] != v {
			assertErrorf(t, "Expected %#v, but got %#v", b, a)
			return
		}
	}
}

func assertParserSuccess(t *testing.T, data interface{}, args ...string) (*Parser, []string) {
	parser := NewParser(data, Default&^PrintErrors)
	ret, err := parser.ParseArgs(args)

	if err != nil {
		t.Fatalf("Unexpected parse error: %s", err)
		return nil, nil
	}

	return parser, ret
}

func assertParseSuccess(t *testing.T, data interface{}, args ...string) []string {
	_, ret := assertParserSuccess(t, data, args...)
	return ret
}

func assertError(t *testing.T, err error, typ ErrorType, msg string) {
	if err == nil {
		assertFatalf(t, "Expected error: %s", msg)
		return
	}

	if e, ok := err.(*Error); !ok {
		assertFatalf(t, "Expected Error type, but got %#v", err)
	} else {
		if e.Type != typ {
			assertErrorf(t, "Expected error type {%s}, but got {%s}", typ, e.Type)
		}

		if e.Message != msg {
			assertErrorf(t, "Expected error message %#v, but got %#v", msg, e.Message)
		}
	}
}

func assertParseFail(t *testing.T, typ ErrorType, msg string, data interface{}, args ...string) []string {
	parser := NewParser(data, Default&^PrintErrors)
	ret, err := parser.ParseArgs(args)

	assertError(t, err, typ, msg)
	return ret
}
