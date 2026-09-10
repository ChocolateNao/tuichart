//go:build linux

package tuichart

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

const (
	linuxTCGETS     = 0x5401
	linuxTCSETS     = 0x5402
	linuxTIOCGPTN   = 0x80045430
	linuxTIOCSPTLCK = 0x40045431
	linuxTIOCSWINSZ = 0x5414
)

// openPTY allocates a pseudo-terminal and forces its canonical+echo modes on
// so the test can verify that Live quiets them.
func openPTY(t *testing.T) (master, slave *os.File) {
	t.Helper()
	ptmx, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("no /dev/ptmx: %v", err)
	}
	var n uint32
	if _, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		ptmx.Fd(),
		linuxTIOCGPTN,
		uintptr(unsafe.Pointer(&n)),
	); errno != 0 {
		ptmx.Close()
		t.Skipf("TIOCGPTN failed: %v", errno)
	}
	var unlock int
	if _, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		ptmx.Fd(),
		linuxTIOCSPTLCK,
		uintptr(unsafe.Pointer(&unlock)),
	); errno != 0 {
		ptmx.Close()
		t.Skipf("TIOCSPTLCK failed: %v", errno)
	}
	sl, err := os.OpenFile(fmt.Sprintf("/dev/pts/%d", n), os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		ptmx.Close()
		t.Skipf("open /dev/pts/%d: %v", n, err)
	}
	return ptmx, sl
}

func ptyGetAttr(fd uintptr, t *syscall.Termios) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		uintptr(linuxTCGETS),
		uintptr(unsafe.Pointer(t)),
	) //nolint:gosec // pty ioctl requires unsafe
	if errno != 0 {
		return errno
	}
	return nil
}

func ptySetAttr(fd uintptr, t *syscall.Termios) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		uintptr(linuxTCSETS),
		uintptr(unsafe.Pointer(t)),
	) //nolint:gosec // pty ioctl requires unsafe
	if errno != 0 {
		return errno
	}
	return nil
}

func ptyReadAll(m *os.File) string {
	fd := int(m.Fd())
	_ = syscall.SetNonblock(fd, true)
	defer func() { _ = syscall.SetNonblock(fd, false) }()
	deadline := time.Now().Add(300 * time.Millisecond)
	var b []byte
	tmp := make([]byte, 1024)
	for time.Now().Before(deadline) {
		n, err := syscall.Read(fd, tmp)
		if n > 0 {
			b = append(b, tmp[:n]...)
			continue
		}
		if err == syscall.EAGAIN || err == nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		return string(b) // EOF / EIO / error
	}
	return string(b)
}

// ptySetSize sets the pseudo-terminal's window size (rows, cols); it
// propagates from master to slave.
func ptySetSize(master *os.File, rows, cols uint16) error {
	ws := winsize{Row: rows, Col: cols}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		master.Fd(),
		uintptr(linuxTIOCSWINSZ),
		uintptr(unsafe.Pointer(&ws)),
	) //nolint:gosec // pty ioctl requires unsafe
	if errno != 0 {
		return errno
	}
	return nil
}

// TestLiveResizeForcesFullRepaintAndClips drives a live renderer whose output
// is a real PTY and verifies the terminal-aware painting contract: frames
// taller than the visible height are clipped to it, unchanged frames still
// diff (no screen resets), and a window resize forces a full repaint.
func TestLiveResizeForcesFullRepaintAndClips(t *testing.T) {
	master, slave := openPTY(t)
	defer master.Close()
	defer slave.Close()
	if err := ptySetSize(master, 10, 60); err != nil {
		t.Skipf("set size 60x10: %v", err)
	}

	g := New(WithNoColor(), WithUnicode(true))
	for _, title := range []string{"a", "b", "c"} {
		g.Add(NewPlot().Title(title))
	}
	lv := NewLive(g, WithLiveOutput(slave))

	lv.Repaint()
	first := ptyReadAll(master)
	if strings.Count(first, seqHome) != 1 {
		t.Fatalf("first frame must be exactly one full paint, got %d: %q",
			strings.Count(first, seqHome), first)
	}
	if !strings.HasPrefix(first, seqHome+seqClearAll) {
		t.Fatalf("terminal paint must home then clear the screen, got %q", first)
	}
	if strings.Contains(first, seqClearBelow) {
		t.Fatalf("terminal paint must not use legacy clear-below, got %q", first)
	}
	if n := strings.Count(first, "\n"); n != 10 {
		t.Fatalf("taller-than-window frame must be clipped to %d rows, painted %d: %q",
			10, n, first)
	}

	// An unchanged frame on the same window still takes the diff path.
	lv.Repaint()
	if second := ptyReadAll(master); strings.Contains(second, seqHome) ||
		strings.Contains(second, seqClearAll) {
		t.Fatalf("unchanged frame must diff, not repaint: %q", second)
	}

	// Resizing the window must blank and repaint everything at the new size.
	if err := ptySetSize(master, 12, 60); err != nil {
		t.Skipf("set size 60x12: %v", err)
	}
	lv.Repaint()
	third := ptyReadAll(master)
	if strings.Count(third, seqHome) != 1 || !strings.Contains(third, seqClearAll) {
		t.Fatalf("resize must force a full repaint: %q", third)
	}
	if n := strings.Count(third, "\n"); n != 12 {
		t.Fatalf("post-resize frame must be clipped to %d rows, painted %d: %q", 12, n, third)
	}
}

// TestLiveSwallowsTerminalInput verifies that typed/scroll bytes arriving on
// the terminal while Live.Run is active never echo onto the alternate screen
// and are swallowed before the program exits.
func TestLiveSwallowsTerminalInput(t *testing.T) {
	master, slave := openPTY(t)
	defer master.Close()
	defer slave.Close()

	var tm syscall.Termios
	if err := ptyGetAttr(slave.Fd(), &tm); err != nil {
		t.Fatalf("get termios: %v", err)
	}
	tm.Lflag |= syscall.ICANON | syscall.ECHO
	if err := ptySetAttr(slave.Fd(), &tm); err != nil {
		t.Fatalf("set termios: %v", err)
	}

	oldStdin := os.Stdin
	os.Stdin = slave
	defer func() { os.Stdin = oldStdin }()

	g := New(WithWidth(60))
	g.Add(NewGauge(50, 100))
	lv := NewLive(g, WithInterval(10*time.Millisecond), WithLiveOutput(io.Discard))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- lv.Run(ctx) }()

	time.Sleep(50 * time.Millisecond) // let it enter and paint once
	if _, err := master.WriteString("\x1b[B\x1b[Aq"); err != nil {
		t.Fatalf("write input: %v", err)
	}
	time.Sleep(70 * time.Millisecond) // reaper window
	cancel()
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after cancel")
	}

	if leak := ptyReadAll(master); leak != "" {
		t.Fatalf("terminal input leaked/echoed during live run: %q", leak)
	}

	if _, err := master.WriteString("z"); err != nil {
		t.Fatalf("write post-check: %v", err)
	}
	if out := ptyReadAll(master); !strings.Contains(out, "z") {
		t.Fatalf("termios was not restored: no echo after Run: %q", out)
	}
}
