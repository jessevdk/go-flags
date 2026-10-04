//go:build !windows && !plan9 && !appengine && !wasm && !aix

package flags

import (
	"os"

	"golang.org/x/sys/unix"
)

func terminalColumns() int {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)

	// The ioctl fails when stdout is not a terminal (e.g. when the output is
	// redirected to a file or piped into another program), in which case we
	// report an unknown width so that the help message is not wrapped.
	if err != nil {
		return 0
	}

	if ws.Col == 0 {
		return defaultTermSize
	}

	return int(ws.Col)
}
