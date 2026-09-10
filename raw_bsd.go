//go:build darwin || freebsd

package tuichart

import (
	"syscall"
	"unsafe"
)

const (
	tcgetattrReq = syscall.TIOCGETA
	tcsetattrReq = syscall.TIOCSETA
)

// fdSetAdd marks fd as readable in a select(2) fd set, treating the whole
// set as a bitmap of 32-bit words so the concrete per-OS FdSet layout does
// not matter.
func fdSetAdd(set *syscall.FdSet, fd int) {
	//nolint:gosec // casting an fd set to a uint32 bitmap is positionally defined
	words := (*[32]uint32)(unsafe.Pointer(set))
	if word := fd / 32; word >= 0 && word < 32 {
		words[word] |= 1 << (uint(fd) % 32)
	}
}

// fdSetTest reports whether fd is set in a select(2) result set.
func fdSetTest(set *syscall.FdSet, fd int) bool {
	words := (*[32]uint32)(unsafe.Pointer(set))
	return words[fd/32]&(1<<(uint(fd)%32)) != 0
}

// selectReady reports whether fd has input available, polling only.
func selectReady(fd int) bool {
	var set syscall.FdSet
	fdSetAdd(&set, fd)
	var tv syscall.Timeval // zero => poll without blocking
	err := syscall.Select(fd+1, &set, nil, nil, &tv)
	if err == syscall.EINTR {
		return selectReady(fd)
	}
	if err != nil {
		return false
	}
	return fdSetTest(&set, fd)
}
