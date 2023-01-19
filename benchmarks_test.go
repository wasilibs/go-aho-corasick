package aho_corasick

import "testing"

var benchmarkReplacerDFA []Replacer

func init() {
	benchmarkReplacerDFA = make([]Replacer, len(testCasesReplace))
	for i, t2 := range testCasesReplace {
		builder := NewAhoCorasickBuilderBenchmark(Opts{
			AsciiCaseInsensitive: true,
			MatchOnlyWholeWords:  true,
			MatchKind:            LeftMostLongestMatch,
			DFA:                  true,
		})
		ac := builder.Build(t2.patterns)
		benchmarkReplacerDFA[i] = NewReplacer(ac)
	}
}

func BenchmarkAhoCorasick_ReplaceAllDFA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for i, ac := range benchmarkReplacerDFA {
			_ = ac.ReplaceAll(testCasesReplace[i].haystack, testCasesReplace[i].replaceWith)
		}
	}
}

func BenchmarkAhoCorasick_ReplaceAllNFA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for i, ac := range acsNFA {
			_ = ac.ReplaceAll(testCasesReplace[i].haystack, testCasesReplace[i].replaceWith)
		}
	}
}

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
