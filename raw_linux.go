//go:build linux

package tuichart

import (
	"syscall"
	"unsafe"
)

const (
	tcgetattrReq = syscall.TCGETS
	tcsetattrReq = syscall.TCSETS
)

// fdSetAdd marks fd as readable in a select(2) fd set.
func fdSetAdd(set *syscall.FdSet, fd int) {
	//nolint:gosec // casting an fd set to a uint32 bitmap is positionally defined
	words := (*[32]uint32)(unsafe.Pointer(&set.Bits[0]))
	if word := fd / 32; word >= 0 && word < 32 {
		words[word] |= 1 << (uint(fd) % 32)
	}
}

// selectReady reports whether fd has input available, polling only.
func selectReady(fd int) bool {
	var set syscall.FdSet
	fdSetAdd(&set, fd)
	var tv syscall.Timeval // zero => poll without blocking
	ready, err := syscall.Select(fd+1, &set, nil, nil, &tv)
	if err == syscall.EINTR {
		return selectReady(fd)
	}
	return err == nil && ready > 0
}
