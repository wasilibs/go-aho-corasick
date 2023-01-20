//go:build aho_corasick_bench_stdlib

package aho_corasick

import ac "github.com/petar-dambovaliev/aho-corasick"

func NewAhoCorasickBuilderBenchmark(o Opts) *ac.AhoCorasickBuilder {
	opts := ac.Opts{
		AsciiCaseInsensitive: o.AsciiCaseInsensitive,
		MatchOnlyWholeWords:  o.MatchOnlyWholeWords,
		DFA:                  o.DFA,
	}
	switch o.MatchKind {
	case LeftMostLongestMatch:
		opts.MatchKind = ac.LeftMostLongestMatch
	case StandardMatch:
		opts.MatchKind = ac.StandardMatch
	case LeftMostFirstMatch:
		opts.MatchKind = ac.LeftMostFirstMatch
	}
	b := ac.NewAhoCorasickBuilder(opts)
	return &b
}

func NewReplacerBenchmark(a ac.AhoCorasick) ac.Replacer {
	return ac.NewReplacer(a)
}
