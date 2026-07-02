//go:build cgo

// This file owns LinaPro's reusable output buffer for the cgo watermark adapter.

package watermark

import "sync"

// bufferPool reuses FFmpeg output buffers to avoid repeated large allocations.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 2*1024*1024)
	},
}
