//go:build linux || darwin || freebsd

package tuichart

import (
	"os"
	"syscall"
	"time"
	"unsafe"
)

func tcgetattr(fd uintptr, t *syscall.Termios) error {
	//nolint:gosec // ioctl requires unsafe.Pointer for syscall
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		uintptr(tcgetattrReq),
		uintptr(unsafe.Pointer(t)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func tcsetattr(fd uintptr, t *syscall.Termios) error {
	//nolint:gosec // ioctl requires unsafe.Pointer for syscall
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		uintptr(tcsetattrReq),
		uintptr(unsafe.Pointer(t)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// acquireInput quiets a terminal for the duration of a Live run: canonical
// mode and echo are disabled so stray input (typed keys, mouse scroll bytes,
// arrow escape sequences) is never reflected onto the screen. The returned
// restore function puts the terminal back exactly as it was before; the
// second return value reports whether the mode was actually engaged.
func acquireInput(f *os.File) (restore func(), engaged bool) {
	noop := func() {}
	if !platformIsTTY(f) {
		return noop, false
	}
	var old syscall.Termios
	if err := tcgetattr(f.Fd(), &old); err != nil {
		return noop, false
	}
	quiet := old
	quiet.Lflag &^= syscall.ICANON | syscall.ECHO
	if err := tcsetattr(f.Fd(), &quiet); err != nil {
		return noop, false
	}
	restore = func() { tcsetattr(f.Fd(), &old) }
	return restore, true
}

// swallowInput discards everything that arrives on f until stop is closed.
// The reaper polls the descriptor with select(2) and only reads once the
// kernel reports input available, so it never depends on O_NONBLOCK and a
// blocking read can never park it. Bytes disappear here so they can never be
// echoed onto the screen or replayed after the terminal is restored.
func swallowInput(f *os.File, stop <-chan struct{}) <-chan struct{} {
	fd := int(f.Fd())
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 1024)
		for {
			if selectReady(fd) {
				n, err := syscall.Read(fd, buf)
				if n > 0 {
					continue
				}
				if err != syscall.EINTR {
					return // EOF / EIO / terminal gone
				}
				continue
			}
			select {
			case <-stop:
				return
			case <-time.After(10 * time.Millisecond):
			}
		}
	}()
	return done
}

// drainStdin drops anything still queued, polled the same way as
// swallowInput. It is called once after the reaper stops, before termios is
// restored, so bytes that arrived in the final instant are never replayed.
func drainStdin(f *os.File) {
	fd := int(f.Fd())
	buf := make([]byte, 4096)
	for {
		if !selectReady(fd) {
			return
		}
		n, err := syscall.Read(fd, buf)
		if n <= 0 {
			if err == syscall.EINTR {
				continue
			}
			return
		}
	}
}
