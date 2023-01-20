//go:build !aho_corasick_bench_stdlib

package aho_corasick

func NewAhoCorasickBuilderBenchmark(o Opts) AhoCorasickBuilder {
	return NewAhoCorasickBuilder(o)
}
