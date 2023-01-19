package aho_corasick

import "testing"

func BenchmarkAhoCorasick_LeftmostInsensitiveWholeWord(b *testing.B) {
	for _, tc := range leftmostInsensitiveWholeWordTestCases {
		tt := tc
		builder := NewAhoCorasickBuilderBenchmark(Opts{
			AsciiCaseInsensitive: true,
			MatchOnlyWholeWords:  true,
			MatchKind:            LeftMostLongestMatch,
		})
		ac := builder.Build(tt.patterns)
		b.Run("", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ac.FindAll(tt.haystack)
			}
		})
	}
}
