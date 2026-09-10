//go:build windows

package tuichart

import (
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// watchResize notifies onResize when the terminal window size changes.
// Windows has no SIGWINCH equivalent, so the console dimensions are polled
// in a background goroutine while f is attached to a real TTY.
func watchResize(f *os.File, onResize func()) (stop func()) {
	if !platformIsTTY(f) {
		return func() {}
	}
	var done int32
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		lastW, lastH, _ := termSize(f)
		for atomic.LoadInt32(&done) == 0 {
			time.Sleep(250 * time.Millisecond)
			w, h, ok := termSize(f)
			if ok && (w != lastW || h != lastH) {
				lastW, lastH = w, h
				onResize()
			}
		}
	}()
	return func() {
		atomic.StoreInt32(&done, 1)
		wg.Wait()
	}
}
