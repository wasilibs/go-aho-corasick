// Benchmarks taken from wrapped Rust library https://github.com/BurntSushi/aho-corasick

package aho_corasick

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"
)

//go:embed testdata/random.txt
var random string

//go:embed testdata/random10x.txt
var random10x string

//go:embed testdata/sherlock.txt
var sherlock string

//go:embed testdata/words-100
var words100Txt string
var words100 = strings.Split(strings.TrimSpace(words100Txt), "\n")

//go:embed testdata/words-5000
var words5000Txt string
var words5000 = strings.Split(strings.TrimSpace(words5000Txt), "\n")

func BenchmarkBurntSushi(b *testing.B) {
	tests := []struct {
		groupName string
		benchName string
		corpus    string
		count     int
		patterns  []string
	}{
		{
			groupName: "random",
			benchName: "onebyte/match",
			corpus:    random,
			count:     352,
			patterns:  []string{"a"},
		},
		{
			groupName: "random",
			benchName: "onebyte/nomatch",
			corpus:    random,
			count:     0,
			patterns:  []string{"\x00"},
		},
		{
			groupName: "random",
			benchName: "twobytes/match",
			corpus:    random,
			count:     352,
			patterns:  []string{"a", "\x00"},
		},
		{
			groupName: "random",
			benchName: "twobytes/nomatch",
			corpus:    random,
			count:     0,
			patterns:  []string{"\x00", "\x01"},
		},
		{
			groupName: "random",
			benchName: "threebytes/match",
			corpus:    random,
			count:     352,
			patterns:  []string{"a", "\x00", "\x01"},
		},
		{
			groupName: "random",
			benchName: "threebytes/nomatch",
			corpus:    random,
			count:     0,
			patterns:  []string{"\x00", "\x01", "\x02"},
		},
		{
			groupName: "random",
			benchName: "fourbytes/match",
			corpus:    random,
			count:     352,
			patterns:  []string{"a", "\x00", "\x01", "\x02"},
		},
		{
			groupName: "random",
			benchName: "fourbytes/nomatch",
			corpus:    random,
			count:     0,
			patterns:  []string{"\x00", "\x01", "\x02", "\x03"},
		},
		{
			groupName: "random",
			benchName: "fivebytes/match",
			corpus:    random,
			count:     352,
			patterns:  []string{"a", "\x00", "\x01", "\x02", "\x03"},
		},
		{
			groupName: "random",
			benchName: "fivebytes/nomatch",
			corpus:    random,
			count:     0,
			patterns:  []string{"\x00", "\x01", "\x02", "\x03", "\x04"},
		},
		{
			groupName: "random",
			benchName: "ten-one-prefix",
			corpus:    random,
			count:     0,
			patterns: []string{
				"zacdef", "zbcdef", "zccdef", "zdcdef", "zecdef", "zfcdef",
				"zgcdef", "zhcdef", "zicdef", "zjcdef",
			},
		},
		{
			groupName: "random",
			benchName: "ten-diff-prefix",
			corpus:    random,
			count:     0,
			patterns: []string{
				"abcdef", "bcdefg", "cdefgh", "defghi", "efghij", "fghijk",
				"ghijkl", "hijklm", "ijklmn", "jklmno",
			},
		},
		{
			groupName: "random10x/leftmost-first",
			benchName: "5000words",
			corpus:    random10x,
			count:     0,
			patterns:  words5000,
		},
		{
			groupName: "random10x/leftmost-first",
			benchName: "100words",
			corpus:    random10x,
			count:     0,
			patterns:  words100,
		},
		{
			groupName: "sherlock",
			benchName: "name/alt1",
			corpus:    sherlock,
			count:     158,
			patterns:  []string{"Sherlock", "Street"},
		},
		{
			groupName: "sherlock",
			benchName: "name/alt2",
			corpus:    sherlock,
			count:     558,
			patterns:  []string{"Sherlock", "Holmes"},
		},
		{
			groupName: "sherlock",
			benchName: "name/alt3",
			corpus:    sherlock,
			count:     740,
			patterns:  []string{"Sherlock", "Holmes", "Watson", "Irene", "Adler", "John", "Baker"},
		},
		{
			groupName: "sherlock",
			benchName: "name/alt4",
			corpus:    sherlock,
			count:     582,
			patterns:  []string{"Sher", "Hol"},
		},
		{
			groupName: "sherlock",
			benchName: "name/alt5",
			corpus:    sherlock,
			count:     639,
			patterns:  []string{"Sherlock", "Holmes", "Watson"},
		},
		{
			groupName: "sherlock",
			benchName: "name/alt6",
			corpus:    sherlock,
			count:     0,
			patterns:  []string{"SherlockZ", "HolmesZ", "WatsonZ", "IreneZ", "MoriartyZ"},
		},
		{
			groupName: "sherlock",
			benchName: "name/alt7",
			corpus:    sherlock,
			count:     0,
			patterns:  []string{"Шерлок Холмс", "Джон Уотсон"},
		},
		{
			groupName: "sherlock",
			benchName: "name/nocase1",
			corpus:    sherlock,
			count:     1764,
			patterns: []string{
				"ADL", "ADl", "AdL", "Adl", "BAK", "BAk", "BAK", "BaK", "Bak",
				"BaK", "HOL", "HOl", "HoL", "Hol", "IRE", "IRe", "IrE", "Ire",
				"JOH", "JOh", "JoH", "Joh", "SHE", "SHe", "ShE", "She", "WAT",
				"WAt", "WaT", "Wat", "aDL", "aDl", "adL", "adl", "bAK", "bAk",
				"bAK", "baK", "bak", "baK", "hOL", "hOl", "hoL", "hol", "iRE",
				"iRe", "irE", "ire", "jOH", "jOh", "joH", "joh", "sHE", "sHe",
				"shE", "she", "wAT", "wAt", "waT", "wat", "ſHE", "ſHe", "ſhE",
				"ſhe",
			},
		},
		{
			groupName: "sherlock",
			benchName: "name/nocase2",
			corpus:    sherlock,
			count:     1307,
			patterns: []string{
				"HOL", "HOl", "HoL", "Hol", "SHE", "SHe", "ShE", "She", "hOL",
				"hOl", "hoL", "hol", "sHE", "sHe", "shE", "she", "ſHE", "ſHe",
				"ſhE", "ſhe",
			},
		},
		{
			groupName: "sherlock",
			benchName: "name/nocase3",
			corpus:    sherlock,
			count:     1442,
			patterns: []string{
				"HOL", "HOl", "HoL", "Hol", "SHE", "SHe", "ShE", "She", "WAT",
				"WAt", "WaT", "Wat", "hOL", "hOl", "hoL", "hol", "sHE", "sHe",
				"shE", "she", "wAT", "wAt", "waT", "wat", "ſHE", "ſHe", "ſhE",
				"ſhe",
			},
		},
		{
			groupName: "sherlock",
			benchName: "5000words",
			corpus:    sherlock,
			count:     567,
			patterns:  words5000,
		},
	}

	for _, tc := range tests {
		tt := tc
		b.Run(fmt.Sprintf("%s/%s", tt.groupName, tt.benchName), func(b *testing.B) {
			for _, dfa := range []bool{false, true} {
				dfaStr := "default"
				if dfa {
					dfaStr = "dfa"
				}
				ac := NewAhoCorasickBuilderBenchmark(Opts{
					MatchKind: LeftMostFirstMatch,
					DFA:       dfa,
				}).Build(tt.patterns)
				b.Run(dfaStr, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						cnt := 0
						iter := ac.Iter(tt.corpus)
						for iter.Next() != nil {
							cnt++
						}
						if cnt != tt.count {
							b.Errorf("expected %d matches, got %d", tt.count, cnt)
						}
					}
				})
			}
		})
	}
}
