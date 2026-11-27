package wafbench

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	coreruleset "github.com/corazawaf/coraza-coreruleset/v4"
	crstests "github.com/corazawaf/coraza-coreruleset/v4/tests"
	"github.com/corazawaf/coraza/v3"
	txhttp "github.com/corazawaf/coraza/v3/http"
	"github.com/corazawaf/coraza/v3/types"
	albedo "github.com/coreruleset/albedo/server"
	"github.com/coreruleset/go-ftw/v2/config"
	"github.com/coreruleset/go-ftw/v2/output"
	"github.com/coreruleset/go-ftw/v2/runner"
	"github.com/coreruleset/go-ftw/v2/test"
	"github.com/rs/zerolog"
)

//go:embed coraza.conf-recommended
var confRecommended string

func BenchmarkWAF(b *testing.B) {
	errorPath, s := setupWAF(b)
	defer s.Close()

	b.Run("FTW", func(b *testing.B) {
		for b.Loop() {
			runFTW(b, errorPath, s)
		}
	})

	for _, size := range []int{1, 1000, 10000, 100000} {
		payload := strings.Repeat("a", size)
		b.Run(fmt.Sprintf("POST/%d", size), func(b *testing.B) {
			for b.Loop() {
				res, err := http.Post(s.URL+"/anything", "text/plain", strings.NewReader(payload))
				if err != nil {
					b.Error(err)
				}
				if res != nil {
					res.Body.Close()
				}
			}
		})
	}
}

// runFTW runs the CRS regression suite against the WAF fronting server.
func runFTW(tb testing.TB, errorPath string, s *httptest.Server) {
	tb.Helper()

	var tests []*test.FTWTest
	err := doublestar.GlobWalk(crstests.FS, "**/*.yaml", func(path string, _ os.DirEntry) error {
		yaml, err := fs.ReadFile(crstests.FS, path)
		if err != nil {
			return fmt.Errorf("wafbench: read file during walk: %w", err)
		}
		t, err := test.GetTestFromYaml(yaml, path)
		if err != nil {
			return fmt.Errorf("wafbench: parse test during walk: %w", err)
		}
		tests = append(tests, t)
		return nil
	})
	if err != nil {
		tb.Fatal(err)
	}
	if len(tests) == 0 {
		tb.Fatal("no tests found")
	}

	u, _ := url.Parse(s.URL)
	host := u.Hostname()
	port, _ := strconv.Atoi(u.Port())
	// TODO(anuraaga): Don't use global config for FTW for better support of programmatic.
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	ftwConf, err := config.NewConfigFromFile(".ftw.yml")
	if err != nil {
		tb.Fatal(err)
	}
	ftwConf.LogFile = errorPath
	ftwConf.TestOverride.Overrides.DestAddr = &host
	ftwConf.TestOverride.Overrides.Port = &port

	runnerConf := config.NewRunnerConfiguration(ftwConf)
	runnerConf.ShowTime = false
	// The 1s default is not always enough when running in CI.
	runnerConf.ReadTimeout = 3 * time.Second

	res, err := runner.Run(runnerConf, tests, output.NewOutput("quiet", os.Stdout))
	if err != nil {
		tb.Fatal(err)
	}

	if len(res.Stats.Failed) > 0 {
		tb.Errorf("failed tests: %v", res.Stats.Failed)
	}
}

// setupWAF starts a WAF-fronted origin server and returns the path of the error
// log the FTW runner scans, together with the server itself.
func setupWAF(tb testing.TB) (string, *httptest.Server) {
	tb.Helper()

	conf := coraza.NewWAFConfig()
	customTestingConfig := `
SecResponseBodyMimeType text/plain
SecDefaultAction "phase:3,log,auditlog,pass"
SecDefaultAction "phase:4,log,auditlog,pass"
SecDefaultAction "phase:5,log,auditlog,pass"
# Rule 900005 from https://github.com/coreruleset/coreruleset/blob/v4.0/dev/tests/regression/README.md#requirements
SecAction "id:900005,\
  phase:1,\
  nolog,\
  pass,\
  ctl:ruleEngine=DetectionOnly,\
  ctl:ruleRemoveById=910000,\
  setvar:tx.blocking_paranoia_level=4,\
  setvar:tx.crs_validate_utf8_encoding=1,\
  setvar:tx.arg_name_length=100,\
  setvar:tx.arg_length=400,\
  setvar:tx.total_arg_length=64000,\
  setvar:tx.max_num_args=255,\
  setvar:tx.max_file_size=64100,\
  setvar:tx.combined_file_sizes=65535"
# Write the value from the X-CRS-Test header as a marker to the log
# Requests with X-CRS-Test header will not be matched by any rule. See https://github.com/coreruleset/go-ftw/pull/133
SecRule REQUEST_HEADERS:X-CRS-Test "@rx ^.*$" \
  "id:999999,\
  phase:1,\
  pass,\
  t:none,\
  log,\
  msg:'X-CRS-Test %{MATCHED_VAR}',\
  ctl:ruleRemoveById=1-999999"
`
	// Configs are loaded with a precise order:
	// 1. Coraza config
	// 2. Custom Rules for testing and eventually overrides of the basic Coraza config
	// 3. CRS basic config
	// 4. CRS rules (on top of which are applied the previously defined SecDefaultAction)
	conf = conf.
		WithRootFS(coreruleset.FS).
		WithDirectives(confRecommended).
		WithDirectives(customTestingConfig).
		WithDirectives("Include @crs-setup.conf.example").
		WithDirectives("Include @owasp_crs/*.conf")

	errorPath := filepath.Join(tb.TempDir(), "error.log")
	errorFile, err := os.Create(errorPath)
	if err != nil {
		tb.Fatalf("failed to create error log: %v", err)
	}
	errorWriter := bufio.NewWriter(errorFile)
	conf = conf.WithErrorCallback(func(rule types.MatchedRule) {
		msg := rule.ErrorLog() + "\n"
		if _, err := io.WriteString(errorWriter, msg); err != nil {
			tb.Fatal(err)
		}
		if err := errorWriter.Flush(); err != nil {
			tb.Fatal(err)
		}
	})

	waf, err := coraza.NewWAF(conf)
	if err != nil {
		tb.Fatal(err)
	}

	// The CRS regression tests expect albedo as the backend server, the same one
	// coraza itself runs them against. See https://github.com/coreruleset/albedo.
	s := httptest.NewServer(txhttp.WrapHandler(waf, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		// Response body tests only work when the mime type matches SecResponseBodyMimeType above.
		w.Header().Set("Content-Type", "text/plain")
		albedo.Handler().ServeHTTP(w, r)
	})))

	return errorPath, s
}
