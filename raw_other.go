//go:build !linux && !darwin && !freebsd && !windows

package tuichart

import "os"

// acquireInput is a no-op on platforms without raw termios support.
func acquireInput(*os.File) (restore func(), engaged bool) { return func() {}, false }

// swallowInput is a no-op on platforms without raw termios support.
func swallowInput(*os.File, <-chan struct{}) <-chan struct{} {
	done := make(chan struct{})
	close(done)
	return done
}

// drainStdin is a no-op on platforms without raw termios support.
func drainStdin(*os.File) {}
