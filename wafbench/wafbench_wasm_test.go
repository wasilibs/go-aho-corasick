//go:build !aho_corasick_bench_stdlib

package wafbench

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/corazawaf/coraza/v3/operators"
	"github.com/corazawaf/coraza/v3/rules"

	ahocorasick "github.com/wasilibs/go-aho-corasick"
)

type pm struct {
	matcher ahocorasick.AhoCorasick
}

var _ rules.Operator = (*pm)(nil)

func newPM(options rules.OperatorOptions) (rules.Operator, error) {
	data := options.Arguments

	data = strings.ToLower(data)
	dict := strings.Split(data, " ")
	builder := ahocorasick.NewAhoCorasickBuilder(ahocorasick.Opts{
		AsciiCaseInsensitive: true,
		MatchOnlyWholeWords:  false,
		MatchKind:            ahocorasick.LeftMostLongestMatch,
		DFA:                  true,
	})

	// TODO this operator is supposed to support snort data syntax: "@pm A|42|C|44|F"
	return &pm{matcher: builder.Build(dict)}, nil
}

func (o *pm) Evaluate(tx rules.TransactionState, value string) bool {
	return pmEvaluate(o.matcher, tx, value)
}

func pmEvaluate(matcher ahocorasick.AhoCorasick, tx rules.TransactionState, value string) bool {
	iter := matcher.Iter(value)

	if !tx.Capturing() {
		// Not capturing so just one match is enough.
		return iter.Next() != nil
	}

	var numMatches int
	for {
		m := iter.Next()
		if m == nil {
			break
		}

		tx.CaptureField(numMatches, value[m.Start():m.End()])

		numMatches++
		if numMatches == 10 {
			return true
		}
	}

	return numMatches > 0
}

func newPMFromDataset(options rules.OperatorOptions) (rules.Operator, error) {
	data := options.Arguments
	dataset, ok := options.Datasets[data]
	if !ok {
		return nil, fmt.Errorf("dataset %q not found", data)
	}
	builder := ahocorasick.NewAhoCorasickBuilder(ahocorasick.Opts{
		AsciiCaseInsensitive: true,
		MatchOnlyWholeWords:  false,
		MatchKind:            ahocorasick.LeftMostLongestMatch,
		DFA:                  true,
	})

	return &pm{matcher: builder.Build(dataset)}, nil
}

func newPMFromFile(options rules.OperatorOptions) (rules.Operator, error) {
	path := options.Arguments

	data, err := loadFromFile(path, options.Path, options.Root)
	if err != nil {
		return nil, err
	}

	var lines []string
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		l := sc.Text()
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		if l[0] == '#' {
			continue
		}
		lines = append(lines, strings.ToLower(l))
	}

	builder := ahocorasick.NewAhoCorasickBuilder(ahocorasick.Opts{
		AsciiCaseInsensitive: true,
		MatchOnlyWholeWords:  false,
		MatchKind:            ahocorasick.LeftMostLongestMatch,
		DFA:                  false,
	})

	return &pm{matcher: builder.Build(lines)}, nil
}

func init() {
	operators.Register("pm", newPM)
	operators.Register("pmFromFile", newPMFromFile)
}
