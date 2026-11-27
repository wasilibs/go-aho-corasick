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

// Rust's standard library calls posix_memalign for aligned allocations on
// wasm32-wasip1.
//
//export posix_memalign
func posix_memalign(memptr *unsafe.Pointer, align uint32, size uint32) int32 {
	// Ignore alignment as in aligned_alloc.
	p := C.malloc(C.ulong(size))
	if p == nil {
		return 12 // ENOMEM
	}
	*memptr = p
	return 0
}
