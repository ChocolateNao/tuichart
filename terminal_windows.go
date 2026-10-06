//go:build windows

package tuichart

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle = kernel32.NewProc("GetStdHandle")
	procGetCSBI      = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

type coord struct{ X, Y int16 }

type smallRect struct{ Left, Top, Right, Bottom int16 }

// consoleScreenBufferInfo mirrors the Win32 CONSOLE_SCREEN_BUFFER_INFO
// record. Only Size and Window are read, but the unread fields are kept as
// byte pads so every field stays at its documented offset and the buffer
// remains large enough for GetConsoleScreenBufferInfo to fill.
type consoleScreenBufferInfo struct {
	Size   coord     // dwSize, offset 0
	_      [6]byte   // dwCursorPosition (4) + wAttributes (2), offset 4
	Window smallRect // srWindow, offset 10
	_      [4]byte   // dwMaximumWindowSize, offset 18
}

// The pads above are only correct if the struct is exactly the size Win32
// writes. Fail the Windows build rather than corrupting memory at runtime.
var _ [22]byte = [unsafe.Sizeof(consoleScreenBufferInfo{})]byte{}

func stdoutHandle() uintptr {
	h, _, _ := procGetStdHandle.Call(^uintptr(10))
	return h
}

func platformIsTTY(f *os.File) bool {
	if f == os.Stdout || f == os.Stderr || f == os.Stdin {
		var info consoleScreenBufferInfo
		r, _, _ := procGetCSBI.Call(stdoutHandle(), uintptr(unsafe.Pointer(&info)))
		return r != 0
	}
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

func termSize(f *os.File) (int, int, bool) {
	if f != os.Stdout && f != os.Stderr && f != os.Stdin {
		return DefaultWidth, DefaultHeight, false
	}
	var info consoleScreenBufferInfo
	r, _, _ := procGetCSBI.Call(stdoutHandle(), uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		return DefaultWidth, DefaultHeight, false
	}
	w := int(info.Window.Right-info.Window.Left) + 1
	h := int(info.Window.Bottom-info.Window.Top) + 1
	if w < 1 || h < 1 {
		return DefaultWidth, DefaultHeight, false
	}
	return w, h, true
}
