//go:build aho_corasick_cgo

package aho_corasick

/*
#cgo LDFLAGS: ${SRCDIR}/buildtools/aho-corasick/target/release/libaho_corasick.a
*/
import "C"
