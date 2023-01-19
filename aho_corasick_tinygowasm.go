//go:build tinygo.wasm || aho_corasick_cgo

package aho_corasick

/*
void* new_matcher(void* patterns, int patterns_len, int ascii_case_insensitive, int dfa, int match_kind);
void* find_iter(void* ac, void* value, int value_len);
int find_iter_next(void* iter, int* patternOut, int* startOut, int* endOut);
void find_iter_delete(void* iter);

void* overlapping_iter(void* ac, void* value, int value_len);
int overlapping_iter_next(void* iter, int* patternOut, int* startOut, int* endOut);
void overlapping_iter_delete(void* iter);
*/
import "C"

import (
	"reflect"
	"runtime"
	"unsafe"
)

type ahoCorasickABI struct{}

func newABI() *ahoCorasickABI {
	return &ahoCorasickABI{}
}

func (abi *ahoCorasickABI) startOperation(memorySize int) {
}

func (abi *ahoCorasickABI) endOperation() {
}

func (abi *ahoCorasickABI) newMatcher(patterns []byte, asciiCaseInsensitive bool, dfa bool, matchKind int) uintptr {
	patternsSh := (*reflect.SliceHeader)(unsafe.Pointer(&patterns))
	aci := 0
	if asciiCaseInsensitive {
		aci = 1
	}

	d := 0
	if dfa {
		d = 1
	}
	ptr := C.new_matcher(unsafe.Pointer(patternsSh.Data), C.int(patternsSh.Len), C.int(aci), C.int(d), C.int(matchKind))
	runtime.KeepAlive(patterns)
	return uintptr(ptr)
}

func (abi *ahoCorasickABI) findIter(acPtr uintptr, value cString) uintptr {
	ptr := C.find_iter(unsafe.Pointer(acPtr), unsafe.Pointer(value.ptr), C.int(value.length))
	return uintptr(ptr)
}

func (abi *ahoCorasickABI) findIterNext(iterPtr uintptr) (pattern int, start int, end int, ok bool) {
	var patternC, startC, endC C.int
	okC := C.find_iter_next(unsafe.Pointer(iterPtr), &patternC, &startC, &endC)
	if okC > 0 {
		ok = true
	}
	return int(patternC), int(startC), int(endC), ok
}

func (abi *ahoCorasickABI) findIterDelete(iterPtr uintptr) {
	C.find_iter_delete(unsafe.Pointer(iterPtr))
}

func (abi *ahoCorasickABI) overlappingIter(acPtr uintptr, value cString) uintptr {
	ptr := C.overlapping_iter(unsafe.Pointer(acPtr), unsafe.Pointer(value.ptr), C.int(value.length))
	return uintptr(ptr)
}

func (abi *ahoCorasickABI) overlappingIterNext(iterPtr uintptr) (pattern int, start int, end int, ok bool) {
	var patternC, startC, endC C.int
	okC := C.overlapping_iter_next(unsafe.Pointer(iterPtr), &patternC, &startC, &endC)
	if okC > 0 {
		ok = true
	}
	return int(patternC), int(startC), int(endC), ok
}

func (abi *ahoCorasickABI) overlappingIterDelete(iterPtr uintptr) {
	C.overlapping_iter_delete(unsafe.Pointer(iterPtr))
}

type cString struct {
	ptr    uintptr
	length int
}

func (abi *ahoCorasickABI) newCString(s string) cString {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return cString{
		ptr:    sh.Data,
		length: int(sh.Len),
	}
}

func (abi *ahoCorasickABI) newOwnedCString(s string) cString {
	return abi.newCString(s)
}

func (abi *ahoCorasickABI) freeOwnedCStringPtr(ptr uintptr) {
}
