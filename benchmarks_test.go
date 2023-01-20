package aho_corasick

import "testing"

func BenchmarkReplaceAll(b *testing.B) {
	for _, tc := range testCasesReplace {
		tt := tc
		b.Run(tt.name, func(b *testing.B) {
			for _, dfa := range []bool{false, true} {
				builder := NewAhoCorasickBuilderBenchmark(Opts{
					AsciiCaseInsensitive: true,
					MatchOnlyWholeWords:  true,
					MatchKind:            LeftMostLongestMatch,
					DFA:                  dfa,
				})
				ac := NewReplacerBenchmark(builder.Build(tt.patterns))
				dfaStr := "nfa"
				if dfa {
					dfaStr = "dfa"
				}
				b.Run(dfaStr, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_ = ac.ReplaceAll(tt.haystack, tt.replaceWith)
					}
				})
			}
		})
	}
}

func BenchmarkLeftmostInsensitiveWholeWord(b *testing.B) {
	for _, tc := range leftmostInsensitiveWholeWordTestCases {
		tt := tc
		b.Run(tt.name, func(b *testing.B) {
			for _, dfa := range []bool{false, true} {
				builder := NewAhoCorasickBuilderBenchmark(Opts{
					AsciiCaseInsensitive: true,
					MatchOnlyWholeWords:  true,
					MatchKind:            LeftMostLongestMatch,
					DFA:                  dfa,
				})
				ac := builder.Build(tt.patterns)
				dfaStr := "nfa"
				if dfa {
					dfaStr = "dfa"
				}
				b.Run(dfaStr, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_ = ac.FindAll(tt.haystack)
					}
				})
			}
		})
	}
}
