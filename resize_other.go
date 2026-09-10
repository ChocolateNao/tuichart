//go:build !unix && !windows

package tuichart

import "os"

// watchResize has nothing to watch on platforms without resize signals or
// console dimension queries; it is a no-op.
func watchResize(_ *os.File, _ func()) (stop func()) {
	return func() {}
}
