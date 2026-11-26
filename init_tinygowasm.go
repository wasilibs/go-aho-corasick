//go:build tinygo.wasm

package aho_corasick

/*
#cgo LDFLAGS: -Linternal/wasm -laho_corasick
*/
import "C"
