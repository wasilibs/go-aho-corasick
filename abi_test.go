package aho_corasick

import (
	"testing"
	"unsafe"
)

// Package-level so the ABI objects escape to the heap, where a zero-size type
// would share one address, rather than staying on the stack.
var abiSinkA, abiSinkB *ahoCorasickABI

// Build frees each Rust matcher from a finalizer on its ABI, so every matcher
// needs its own ABI object. A zero-size ABI would make every newABI call return
// the same address, whose finalizers never run, and leak every matcher.
func TestNewABIDistinct(t *testing.T) {
	abiSinkA, abiSinkB = newABI(), newABI()
	if abiSinkA == abiSinkB {
		t.Fatal("newABI returned the same object twice; finalizers on it would never run")
	}
}

// Go's tiny allocator packs pointer-free objects under 16 bytes into shared
// blocks, and a finalizer on one only runs once the whole block is dead. A
// long-lived matcher would then keep the matchers sharing its block alive.
func TestABINotTinyAllocated(t *testing.T) {
	if size := unsafe.Sizeof(ahoCorasickABI{}); size < 16 {
		t.Fatalf("ahoCorasickABI is %d bytes; it must be at least 16 so its finalizer runs", size)
	}
}
