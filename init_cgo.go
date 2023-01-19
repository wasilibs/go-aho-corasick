//go:build aho_corasick_cgo

package aho_corasick

/*
#cgo LDFLAGS: -L${SRCDIR}/buildtools/aho-corasick/target/release -laho_corasick
*/
import "C"
