// +build darwin freebsd linux netbsd openbsd

package flags

import (
	"syscall"
	"unsafe"
)

// #include <sys/ioctl.h>
// enum { _GO_TIOCGWINSZ = TIOCGWINSZ };
import "C"

type winsize struct {
	ws_row, ws_col       uint16
	ws_xpixel, ws_ypixel uint16
}

func getTerminalColumns() int {
	ws := winsize{}

	syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(0),
		uintptr(C._GO_TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)))

	return int(ws.ws_col)
}
