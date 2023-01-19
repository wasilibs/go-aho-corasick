//go:build aho_corasick_cgo

package aho_corasick

/*
#cgo LDFLAGS: -Lbuildtools/aho-corasick/target/release -laho_corasick
*/
import "C"
