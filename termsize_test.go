package flags

import (
	"testing"
)

func init() {
	// Keep the generated help output independent of the terminal the tests
	// happen to run in.
	terminalColumnsOverride = defaultTermSize
}

func TestTerminalColumnsOverride(t *testing.T) {
	saved := terminalColumnsOverride
	defer func() { terminalColumnsOverride = saved }()

	terminalColumnsOverride = 123

	if cols := getTerminalColumns(); cols != 123 {
		t.Errorf("Expected the override to be honoured, but got %d", cols)
	}

	// A non-positive override falls through to querying the terminal, which
	// reports the default when there is none.
	terminalColumnsOverride = 0

	if cols := getTerminalColumns(); cols != terminalColumns() {
		t.Errorf("Expected the terminal to be queried, but got %d", cols)
	}
}
