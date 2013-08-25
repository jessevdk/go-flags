// +build !windows

package flags

import (
	"syscall"
	"unsafe"
)

type winsize struct {
	ws_row, ws_col       uint16
	ws_xpixel, ws_ypixel uint16
}

func getTerminalColumns() int {
	ws := winsize{}

	if TIOCGWINSZ != 0 {
		syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(0),
			uintptr(TIOCGWINSZ),
			uintptr(unsafe.Pointer(&ws)))

		return int(ws.ws_col)
	}

	return 80
}
