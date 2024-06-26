//go:build !aho_corasick_bench_stdlib

package wafbench

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
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
	if !tx.Capturing() {
		// Not capturing so just one match is enough.
		return len(matcher.FindN(value, 1)) > 0
	}

	var numMatches int
	for _, m := range matcher.FindN(value, 10) {
		tx.CaptureField(numMatches, value[m.Start():m.End()])
		numMatches++
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

var errEmptyPaths = errors.New("empty paths")

func loadFromFile(filepath string, paths []string, root fs.FS) ([]byte, error) {
	if path.IsAbs(filepath) {
		return fs.ReadFile(root, filepath)
	}

	if len(paths) == 0 {
		return nil, errEmptyPaths
	}

	// handling files by operators is hard because we must know the paths where we can
	// search, for example, the policy path or the binary path...
	// CRS stores the .data files in the same directory as the directives
	var (
		content []byte
		err     error
	)

	for _, p := range paths {
		absFilepath := path.Join(p, filepath)
		content, err = fs.ReadFile(root, absFilepath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				return nil, err
			}
		}

		return content, nil
	}

	return nil, err
}

func init() {
	operators.Register("pm", newPM)
	operators.Register("pmFromDataset", newPMFromDataset)
	operators.Register("pmFromFile", newPMFromFile)
}
