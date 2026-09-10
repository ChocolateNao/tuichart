//go:build unix

package tuichart

import (
	"os"
	"os/signal"
	"syscall"
)

// watchResize notifies onResize whenever the terminal window is resized
// while f is attached to a real TTY. It installs a SIGWINCH handler in a
// background goroutine and returns a stop function. Callers that pass a
// non-TTY writer should skip this entirely.
func watchResize(f *os.File, onResize func()) (stop func()) {
	if !platformIsTTY(f) {
		return func() {}
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			onResize()
		}
	}()
	return func() {
		signal.Stop(ch)
		close(ch)
	}
}
