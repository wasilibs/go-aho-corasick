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

	"github.com/corazawaf/coraza/v3/experimental/plugins"
	"github.com/corazawaf/coraza/v3/experimental/plugins/plugintypes"

	ahocorasick "github.com/wasilibs/go-aho-corasick"
)

// The operators below mirror Coraza's built-in pm, pmFromDataset and
// pmFromFile, replacing petar-dambovaliev/aho-corasick with this library.

type pm struct {
	matcher ahocorasick.AhoCorasick
	// minLen is the length of the shortest pattern. If the input is shorter
	// than this, no pattern can match and the automaton is skipped.
	minLen int
}

var _ plugintypes.Operator = (*pm)(nil)

func newPM(options plugintypes.OperatorOptions) (plugintypes.Operator, error) {
	data := strings.ToLower(options.Arguments)
	dict := strings.Split(data, " ")

	// TODO this operator is supposed to support snort data syntax: "@pm A|42|C|44|F"
	return buildPM(options.Memoizer, data, dict, true)
}

func newPMFromDataset(options plugintypes.OperatorOptions) (plugintypes.Operator, error) {
	data := options.Arguments
	dataset, ok := options.Datasets[data]
	if !ok {
		return nil, fmt.Errorf("dataset %q not found", data)
	}

	return buildPM(options.Memoizer, data, dataset, true)
}

func newPMFromFile(options plugintypes.OperatorOptions) (plugintypes.Operator, error) {
	filepath := options.Arguments

	data, err := loadFromFile(filepath, options.Path, options.Root)
	if err != nil {
		return nil, err
	}

	var lines []string
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if len(l) == 0 || l[0] == '#' {
			continue
		}
		lines = append(lines, strings.ToLower(l))
	}

	return buildPM(options.Memoizer, strings.Join(options.Path, ",")+filepath, lines, false)
}

func buildPM(m plugintypes.Memoizer, key string, patterns []string, dfa bool) (plugintypes.Operator, error) {
	build := func() (any, error) {
		return ahocorasick.NewAhoCorasickBuilder(ahocorasick.Opts{
			AsciiCaseInsensitive: true,
			MatchOnlyWholeWords:  false,
			MatchKind:            ahocorasick.LeftMostLongestMatch,
			DFA:                  dfa,
		}).Build(patterns), nil
	}

	var matcher any
	if m != nil {
		var err error
		if matcher, err = m.Do(key, build); err != nil {
			return nil, fmt.Errorf("wafbench: build matcher: %w", err)
		}
	} else {
		matcher, _ = build()
	}

	ac, ok := matcher.(ahocorasick.AhoCorasick)
	if !ok {
		return nil, fmt.Errorf("wafbench: unexpected cached matcher type %T", matcher)
	}
	return &pm{matcher: ac, minLen: minPatternLen(patterns)}, nil
}

func (o *pm) Evaluate(tx plugintypes.TransactionState, value string) bool {
	if len(value) < o.minLen {
		return false
	}

	if !tx.Capturing() {
		// Not capturing so just one match is enough.
		return len(o.matcher.FindN(value, 1)) > 0
	}

	matches := o.matcher.FindN(value, 10)
	for i, m := range matches {
		tx.CaptureField(i, value[m.Start():m.End()])
	}
	return len(matches) > 0
}

// minPatternLen returns the length of the shortest pattern, or 0 if any pattern
// is empty, since an empty pattern matches every input.
func minPatternLen(patterns []string) int {
	minLen := 0
	for _, p := range patterns {
		if len(p) == 0 {
			return 0
		}
		if minLen == 0 || len(p) < minLen {
			minLen = len(p)
		}
	}
	return minLen
}

var errEmptyPaths = errors.New("empty paths")

func loadFromFile(filepath string, paths []string, root fs.FS) ([]byte, error) {
	if path.IsAbs(filepath) {
		content, err := fs.ReadFile(root, filepath)
		if err != nil {
			return nil, fmt.Errorf("wafbench: read %s: %w", filepath, err)
		}
		return content, nil
	}

	if len(paths) == 0 {
		return nil, errEmptyPaths
	}

	// CRS stores the .data files in the same directory as the directives.
	var err error
	for _, p := range paths {
		absFilepath := path.Join(p, filepath)
		var content []byte
		content, err = fs.ReadFile(root, absFilepath)
		if err == nil {
			return content, nil
		}
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("wafbench: read %s: %w", absFilepath, err)
		}
	}

	return nil, fmt.Errorf("wafbench: %s not found in %v: %w", filepath, paths, err)
}

func init() {
	plugins.RegisterOperator("pm", newPM)
	plugins.RegisterOperator("pmFromDataset", newPMFromDataset)
	plugins.RegisterOperator("pmFromFile", newPMFromFile)
}
