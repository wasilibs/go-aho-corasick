//go:build !aho_corasick_bench_stdlib

package aho_corasick

type ReplacerBenchmark = Replacer

func NewAhoCorasickBuilderBenchmark(o Opts) AhoCorasickBuilder {
	return NewAhoCorasickBuilder(o)
}

func NewReplacerBenchmark(a AhoCorasick) Replacer {
	return NewReplacer(a)
}
