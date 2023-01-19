//go:build tinygo.wasm && !gc.custom && !custommalloc

package aho_corasick

import (
	"unsafe"
)

/*
void* malloc(unsigned long size);
*/
import "C"

// TinyGo currently only includes a subset of malloc functions by default, so we
// reimplement the remaining here.

//export aligned_alloc
func aligned_alloc(align uint32, size uint32) unsafe.Pointer {
	// Ignore alignment and hope for best, TinyGo by default does not
	// provide a way to allocate aligned memory.
	return C.malloc(C.ulong(size))
}
