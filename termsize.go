//go:build !windows && !plan9 && !appengine && !wasm && !aix
// +build !windows,!plan9,!appengine,!wasm,!aix

package flags

import (
	"golang.org/x/sys/unix"
)

func terminalColumns() int {
	ws, err := unix.IoctlGetWinsize(0, unix.TIOCGWINSZ)
	if err != nil {
		return defaultTermSize
	}
	return int(ws.Col)
}
