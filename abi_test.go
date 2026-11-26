package aho_corasick

import "testing"

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
