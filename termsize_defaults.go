package flags

// defaultTermSize is the terminal width assumed when a terminal is attached
// but its actual width could not be determined.
const defaultTermSize = 100

// terminalColumnsOverride, when positive, is returned by getTerminalColumns
// instead of querying the terminal. The test suite sets it so that the
// generated help output does not depend on the terminal the tests happen to
// run in.
var terminalColumnsOverride int

// getTerminalColumns returns the width of the terminal attached to stdout, or
// 0 when there is no terminal attached and the help message should therefore
// not be wrapped.
func getTerminalColumns() int {
	if terminalColumnsOverride > 0 {
		return terminalColumnsOverride
	}

	return terminalColumns()
}
