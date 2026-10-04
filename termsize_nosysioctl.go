//go:build plan9 || appengine || wasm || aix
// +build plan9 appengine wasm aix

package flags

func terminalColumns() int {
	return defaultTermSize
}
