//go:build windows

package tuichart

import (
	"os"
	"unsafe"
)

const (
	enableLineInput uint32 = 0x0002
	enableEchoInput uint32 = 0x0004
)

var (
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

// acquireInput disables line/echo input on a Windows console so stray input
// is never reflected onto the alternate screen.
func acquireInput(f *os.File) (restore func(), engaged bool) {
	noop := func() {}
	if !platformIsTTY(f) {
		return noop, false
	}
	var mode uint32
	if r, _, _ := procGetConsoleMode.Call(f.Fd(), uintptr(unsafe.Pointer(&mode))); r == 0 {
		return noop, false
	}
	quiet := mode &^ (enableLineInput | enableEchoInput)
	if r, _, _ := procSetConsoleMode.Call(f.Fd(), uintptr(quiet)); r == 0 {
		return noop, false
	}
	return func() {
		procSetConsoleMode.Call(f.Fd(), uintptr(mode))
	}, true
}

// swallowInput is a no-op on Windows; echo is already disabled so nothing
// reaches the screen, and the stdlib offers no non-blocking console drain.
func swallowInput(*os.File, <-chan struct{}) <-chan struct{} {
	done := make(chan struct{})
	close(done)
	return done
}

// drainStdin is a no-op on Windows.
func drainStdin(*os.File) {}
