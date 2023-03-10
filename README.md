# go-aho-corasick

go-aho-corasick is a library for aho-corasick which wraps the Rust [BurntSushi/aho-corasick][2] library.
It provides the same API as [petar-dambovaliev/aho-corasick][1] and is a drop-in replacement with full
API and behavior compatibility. By default, aho-corasick is packaged as a WebAssembly module and accessed
with the pure Go runtime,  [wazero][3]. This means that it is compatible with any Go application, regardless
of availability of cgo.

For TinyGo applications being built for WASM, this library will perform significantly better. For Go applications,
the performance difference varies by case. [petar-dambovaliev/aho-corasick][1] performs quite well, but this
library seems to perform better with 4+ patterns, and larger ones. The API is a drop-in replacement, so it
is best to try it and benchmark to see the effect. Notably, for unknown input where a decision on using DFA or
not cannot be made beforehand, the automatic detection in this library will often perform much better.

## Behavior differences

This library will automatically pick an implementations among NFA, contiguous NFA, and DFA unless the `DFA` flag
is explicitly passed when constructing. This generally results in better performance than always using one or the
other.

## Usage

go-aho-corasick is a standard Go library package and can be added to a go.mod file. It will work fine in
Go or TinyGo projects.

```
go get github.com/wasilibs/go-aho-corasick
```

Because the library is a drop-in replacement for [petar-dambovaliev/aho-corasick][1], an import rewrite can
make migrating code to use it simple.

```go
import "github.com/peter-dambovaliev/aho-corasick"
```

can be changed to

```go
import "github.com/wasilibs/go-aho-corasick"
```

### cgo

This library also supports opting into using cgo to wrap [BurntSushi/aho-corasick][2] instead
of using WebAssembly. This requires having a built version of the library available - because
Rust libraries are not easily installed for use with `pkg-config`, environment variables like
`CGO_LDFLAGS` and `LD_LIBRARY_PATH` may be needed to find the library at build and runtime.
The build tag `aho_corasick_cgo` can be used to enable cgo support.

## Performance

Benchmarks are run against every commit in the [bench][4] workflow. GitHub action runners are highly
virtualized and do not have stable performance across runs, but the relative numbers within a run
should still be somewhat, though not precisely, informative.

### Microbenchmarks

Microbenchmarks are the same as included in the reference Rust and Go libraries. Full results can be
viewed in the workflow, a sample of results for one run looks like this

```
BurntSushi/random/onebyte/match/default-2                           51.3??s ?? 3%     147.5??s ?? 0%          74.7??s ?? 1%
BurntSushi/random/onebyte/match/dfa-2                               64.4??s ?? 0%     148.8??s ?? 0%          76.9??s ?? 4%
BurntSushi/random/onebyte/nomatch/default-2                         5.67??s ?? 2%      7.33??s ??12%          1.33??s ??19%
BurntSushi/random/onebyte/nomatch/dfa-2                             5.79??s ?? 1%      6.98??s ?? 5%          1.50??s ?? 1%
BurntSushi/random/twobytes/match/default-2                          53.9??s ?? 0%     156.3??s ?? 0%          78.2??s ?? 0%
BurntSushi/random/twobytes/match/dfa-2                              66.6??s ?? 0%     156.9??s ?? 1%          79.9??s ?? 3%
BurntSushi/random/twobytes/nomatch/default-2                        10.0??s ?? 0%      10.9??s ?? 0%           1.6??s ?? 1%
BurntSushi/random/twobytes/nomatch/dfa-2                            10.0??s ?? 1%      11.0??s ?? 1%           1.6??s ?? 1%
BurntSushi/random/threebytes/match/default-2                        55.7??s ?? 1%     162.1??s ?? 0%          81.0??s ?? 2%
BurntSushi/random/threebytes/match/dfa-2                            67.3??s ?? 0%     162.5??s ?? 0%          81.4??s ?? 1%
BurntSushi/random/threebytes/nomatch/default-2                      10.0??s ?? 0%      13.0??s ?? 0%           1.7??s ?? 0%
BurntSushi/random/threebytes/nomatch/dfa-2                          10.0??s ?? 0%      12.9??s ?? 1%           1.7??s ?? 0%
BurntSushi/random/fourbytes/match/default-2                          188??s ?? 0%       174??s ?? 0%            86??s ?? 1%
BurntSushi/random/fourbytes/match/dfa-2                              108??s ?? 0%       180??s ?? 0%            88??s ?? 2%
BurntSushi/random/fourbytes/nomatch/default-2                        158??s ?? 0%        56??s ?? 1%             2??s ?? 1%
BurntSushi/random/fourbytes/nomatch/dfa-2                           64.5??s ?? 0%      58.3??s ?? 0%           1.9??s ?? 1%
BurntSushi/random/fivebytes/match/default-2                          188??s ?? 0%       177??s ?? 0%            85??s ?? 1%
BurntSushi/random/fivebytes/match/dfa-2                              119??s ?? 0%       181??s ?? 0%            85??s ?? 1%
BurntSushi/random/fivebytes/nomatch/default-2                        158??s ?? 0%        56??s ?? 1%             2??s ?? 1%
BurntSushi/random/fivebytes/nomatch/dfa-2                           64.5??s ?? 0%      58.4??s ?? 0%           1.9??s ?? 1%
BurntSushi/random/ten-one-prefix/default-2                          33.1??s ?? 1%      36.2??s ?? 0%           6.0??s ?? 1%
BurntSushi/random/ten-one-prefix/dfa-2                              26.9??s ?? 0%      40.6??s ?? 0%           5.7??s ?? 1%
BurntSushi/random/ten-diff-prefix/default-2                          228??s ?? 0%       253??s ?? 0%            35??s ?? 1%
BurntSushi/random/ten-diff-prefix/dfa-2                             64.5??s ?? 0%     281.5??s ?? 0%          38.9??s ?? 1%
BurntSushi/random10x/leftmost-first/5000words/default-2             3.19ms ?? 0%      2.84ms ?? 1%          0.92ms ?? 0%
BurntSushi/random10x/leftmost-first/5000words/dfa-2                  646??s ?? 0%       652??s ?? 0%           332??s ?? 0%
BurntSushi/random10x/leftmost-first/100words/default-2              2.61ms ?? 0%      0.59ms ?? 0%          0.25ms ?? 0%
BurntSushi/random10x/leftmost-first/100words/dfa-2                   647??s ?? 0%       573??s ?? 0%           251??s ?? 0%
BurntSushi/sherlock/name/alt1/default-2                              443??s ?? 1%       469??s ?? 0%            64??s ?? 0%
BurntSushi/sherlock/name/alt1/dfa-2                                  425??s ?? 2%       481??s ?? 0%            64??s ?? 1%
BurntSushi/sherlock/name/alt2/default-2                              813??s ?? 1%      1002??s ?? 0%           209??s ?? 0%
BurntSushi/sherlock/name/alt2/dfa-2                                  793??s ?? 0%      1022??s ?? 0%           212??s ?? 1%
BurntSushi/sherlock/name/alt3/default-2                             9.61ms ?? 0%      3.56ms ?? 1%          0.26ms ?? 0%
BurntSushi/sherlock/name/alt3/dfa-2                                 3.95ms ?? 0%      3.65ms ?? 0%          0.26ms ?? 1%
BurntSushi/sherlock/name/alt4/default-2                              772??s ?? 0%       999??s ?? 0%           208??s ?? 1%
BurntSushi/sherlock/name/alt4/dfa-2                                  763??s ?? 0%      1020??s ?? 0%           213??s ?? 1%
BurntSushi/sherlock/name/alt5/default-2                              891??s ?? 1%      1253??s ?? 0%           255??s ?? 0%
BurntSushi/sherlock/name/alt5/dfa-2                                  881??s ?? 1%      1284??s ?? 0%           258??s ?? 1%
BurntSushi/sherlock/name/alt6/default-2                             9.54ms ?? 0%      0.30ms ?? 0%          0.01ms ?? 0%
BurntSushi/sherlock/name/alt6/dfa-2                                 3.86ms ?? 0%      0.30ms ?? 0%          0.01ms ?? 0%
BurntSushi/sherlock/name/alt7/default-2                              545??s ?? 0%       559??s ?? 0%            15??s ?? 0%
BurntSushi/sherlock/name/alt7/dfa-2                                  542??s ?? 0%       561??s ?? 0%            15??s ?? 0%
BurntSushi/sherlock/name/nocase1/default-2                          12.7ms ?? 0%       4.1ms ?? 0%           0.9ms ?? 1%
BurntSushi/sherlock/name/nocase1/dfa-2                              4.08ms ?? 0%      4.04ms ?? 0%          0.86ms ?? 1%
BurntSushi/sherlock/name/nocase2/default-2                          10.7ms ?? 0%       3.8ms ?? 0%           0.4ms ?? 1%
BurntSushi/sherlock/name/nocase2/dfa-2                              4.03ms ?? 0%      3.85ms ?? 0%          0.42ms ?? 1%
BurntSushi/sherlock/name/nocase3/default-2                          10.9ms ?? 0%       3.9ms ?? 0%           0.8ms ?? 0%
BurntSushi/sherlock/name/nocase3/dfa-2                              4.05ms ?? 0%      3.90ms ?? 0%          0.79ms ?? 0%
BurntSushi/sherlock/5000words/default-2                             21.6ms ?? 0%      18.6ms ?? 0%           8.4ms ?? 0%
BurntSushi/sherlock/5000words/dfa-2                                 4.35ms ?? 0%      4.93ms ?? 0%          2.65ms ?? 0%
```

Random gives us a look at various different patterns against a random corpus. We see that the four-byte case, with
four patterns, seems to be a threshold where this library performs better with auto-detection of the matcher.

Sherlock is a more real-world example, with the literary text for Sherlock Holmes. We again see that cases with
4+ patterns seem to perform significantly better.

### wafbench

wafbench tests the performance of replacing the pm operator of the OWASP [CoreRuleSet][5] and
[Coraza][6] implementation with this library. This benchmark is considered a real world performance
test, though the bulk of processing is in regex, not pm. The pm matchers themselves have a large
number of patterns and are considered complex.

```
WAF/FTW-2                         28.2s ?? 0%          28.3s ?? 0%              28.3s ?? 0%
WAF/POST/1-2                     3.08ms ?? 2%         3.09ms ?? 2%             3.10ms ?? 2%
WAF/POST/1000-2                  20.0ms ?? 1%         19.9ms ?? 1%             20.2ms ?? 1%
WAF/POST/10000-2                  190ms ?? 0%          191ms ?? 1%              191ms ?? 1%
WAF/POST/100000-2                 1.88s ?? 0%          1.87s ?? 1%              1.88s ?? 1%
```

In all cases, the libraries perform the same, within noise. This likely reflects that the test
is mostly spending time in regex, not pm. It's presented mostly as a case study that wrapping
with WebAssembly can provide the same performance without needing to rewrite an entire library.

[1]: https://github.com/petar-dambovaliev/aho-corasick
[2]: https://github.com/BurntSushi/aho-corasick
[3]: https://wazero.io
[4]: https://github.com/wasilibs/go-aho-corasick/actions/workflows/bench.yaml
[5]: https://github.com/coreruleset/coreruleset
[6]: https://github.com/corazawaf/coraza
