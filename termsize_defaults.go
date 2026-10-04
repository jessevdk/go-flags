package flags

const defaultTermSize = 80

// terminalColumnsOverride, when positive, is returned by getTerminalColumns
// instead of querying the terminal. The test suite sets it so that the
// generated help output does not depend on the terminal the tests happen to
// run in.
var terminalColumnsOverride int

func getTerminalColumns() int {
	if terminalColumnsOverride > 0 {
		return terminalColumnsOverride
	}

	return terminalColumns()
}
