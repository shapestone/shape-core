package hostiletest_test

import (
	"bytes"
	"errors"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/shapestone/shape-core/pkg/data"
	"github.com/shapestone/shape-core/pkg/hostiletest"
)

// TestJSONCorpus runs the JSON corpus through the Guard-based reference
// parser, with encoding/json as the differential reference. Inputs over
// 8 MiB are skipped unless HOSTILETEST_HEAVY=1, because the 64 MiB cases
// allocate close to 1 GiB, which does not fit under the race detector on a
// small runner.
func TestJSONCorpus(t *testing.T) {
	opts := []hostiletest.Option{hostiletest.WithReference(stdlibJSON)}
	if os.Getenv("HOSTILETEST_HEAVY") != "1" {
		opts = append(opts, hostiletest.MaxSize(8<<20))
	}
	// The reference parser is the first adopter: nothing is a known gap.
	opts = append(opts, hostiletest.Strict())
	hostiletest.Run(t, hostiletest.JSON, parseJSON, opts...)
}

// TestLiveHeapWithinGuardEstimate is the check the harness makes under Wire,
// run near the budget: a parse that uses most of a 64 MiB MaxMemory must keep
// no more live heap, after a collection, than the Guard estimated. The harness
// check itself only runs on small results under Tight, far from its budget.
func TestLiveHeapWithinGuardEstimate(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a 1.4 MB document and a 40 MB result")
	}
	const n = 700_000
	in := make([]byte, 0, 2*n+2)
	in = append(in, '[')
	for i := 0; i < n; i++ {
		in = append(in, '0', ',')
	}
	in = append(in, '0', ']')

	req := hostiletest.RequestFor(hostiletest.Default)
	req.Limits.MaxMemory = 64 << 20

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	// Grow is off: it refuses a parse whose growth peak passes the budget,
	// which stops this one before it gets near 64 MiB. The check here is the
	// live heap of a result that does fill the budget, not the peak.
	v, g, err := parseJSONWith(in, req, true)
	if err != nil {
		t.Fatal(err)
	}
	runtime.GC()
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(v)

	est := g.MemoryUsed()
	live := int64(after.HeapAlloc) - int64(before.HeapAlloc)
	t.Logf("estimate %d MiB, live heap %d MiB, budget %d MiB", est>>20, live>>20, req.Limits.MaxMemory>>20)
	if est < req.Limits.MaxMemory*3/4 {
		t.Fatalf("estimate %d MiB is not near the %d MiB budget; the test no longer exercises it", est>>20, req.Limits.MaxMemory>>20)
	}
	if live > est {
		t.Errorf("result holds %d MiB of live heap, over the Guard estimate of %d MiB", live>>20, est>>20)
	}
}

func TestRequestFor(t *testing.T) {
	w := hostiletest.RequestFor(hostiletest.Wire)
	if w.Profile != hostiletest.Wire || w.Limits != hostiletest.Limits(hostiletest.Wire) || !w.Policy.RejectDuplicateKeys {
		t.Errorf("wire request = %+v", w)
	}
	for _, p := range []hostiletest.Profile{hostiletest.Default, hostiletest.Tight, "some-future-profile"} {
		if hostiletest.RequestFor(p).Policy.RejectDuplicateKeys {
			t.Errorf("%s must not reject duplicate keys", p)
		}
	}
	// Profile is an open set: an unknown one gets the defaults.
	if hostiletest.Limits("some-future-profile") != (data.Limits{}) {
		t.Error("unknown profile must get data.Limits{}")
	}
}

func TestCorpusIntegrity(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range hostiletest.All() {
		key := string(c.Format) + "/" + c.Name
		if seen[key] {
			t.Errorf("duplicate case %s", key)
		}
		seen[key] = true

		if strings.ContainsAny(c.Name, " /") || c.Name == "" {
			t.Errorf("%s: name must be non-empty with no spaces or slashes", key)
		}
		if c.Input == nil || len(c.Want) == 0 {
			t.Errorf("%s: needs an Input and at least one expectation", key)
		}
		for p, w := range c.Want {
			if w.Outcome == 0 {
				t.Errorf("%s: empty expectation for %s", key, p)
			}
			if w.Outcome&hostiletest.Limit == 0 && len(w.Limits) > 0 {
				t.Errorf("%s: Limits set without the Limit outcome", key)
			}
			if w.Outcome&hostiletest.Record != 0 && w.Outcome != hostiletest.Record {
				t.Errorf("%s: Record must stand alone", key)
			}
		}
		for p := range c.Known {
			if _, ok := c.Want[p]; !ok {
				t.Errorf("%s: Known entry for %s, which is not run", key, p)
			}
		}
		if c.Diff && (c.Size > 1<<20 || c.Isolated()) {
			t.Errorf("%s: Diff on a case the reference cannot safely take", key)
		}
		if c.Isolated() && !c.Heavy() {
			t.Errorf("%s: isolated cases should be heavy", key)
		}
	}
	for _, f := range hostiletest.Formats() {
		if len(hostiletest.For(f)) == 0 {
			t.Errorf("no cases for %s", f)
		}
	}
}

// TestCorpusSizesAndDeterminism checks that Size matches the generated input
// and that generators are repeatable. Large inputs are checked once.
func TestCorpusSizesAndDeterminism(t *testing.T) {
	for _, c := range hostiletest.All() {
		if testing.Short() && c.Heavy() {
			continue
		}
		a := c.Input()
		if int64(len(a)) != c.Size {
			t.Errorf("%s/%s: Size %d but input is %d bytes", c.Format, c.Name, c.Size, len(a))
			continue
		}
		if c.Size <= 1<<20 && !bytes.Equal(a, c.Input()) {
			t.Errorf("%s/%s: generator is not deterministic", c.Format, c.Name)
		}
	}
}

// TestCorpusKnownCoversUnadoptedLimits: with no format on pkg/data yet, every
// case that demands only a LimitError is a known gap, and every crash-tagged
// case names its reason.
func TestCorpusKnownCoversUnadoptedLimits(t *testing.T) {
	for _, c := range hostiletest.All() {
		for p, w := range c.Want {
			if w.Outcome == hostiletest.Limit && c.Known[p] == "" {
				t.Errorf("%s/%s (%s): demands a limit error but is not Known", c.Format, c.Name, p)
			}
		}
	}
}

func TestProfiles(t *testing.T) {
	w := hostiletest.Limits(hostiletest.Wire)
	if w.MaxBytes != 64<<20 || w.MaxMemory != 1<<30 || w.MaxDepth != data.DefaultMaxDepth {
		t.Errorf("wire limits = %+v", w)
	}
	if w.MaxNodes != 0 || w.MaxStringBytes != 0 || w.MaxKeysPerMap != 0 {
		t.Errorf("wire must leave nodes, string bytes and keys unlimited: %+v", w)
	}
	if hostiletest.Limits(hostiletest.Default) != (data.Limits{}) {
		t.Error("default profile must be data.Limits{}")
	}
}

// TestCasesNameEveryLimit makes sure the corpus can provoke each limit name
// with the reference parser, so the names in Expect are real.
func TestCasesNameEveryLimit(t *testing.T) {
	want := map[string]bool{
		data.LimitDepth: false, data.LimitBytes: false, data.LimitNodes: false,
		data.LimitStringBytes: false, data.LimitKeysPerMap: false,
		data.LimitAliasExpansion: false, data.LimitMemory: false,
	}
	for _, c := range hostiletest.All() {
		for _, w := range c.Want {
			for _, n := range w.Limits {
				if _, ok := want[n]; !ok {
					t.Errorf("%s/%s: unknown limit name %q", c.Format, c.Name, n)
				}
				want[n] = true
			}
		}
	}
	for n, used := range want {
		if !used {
			t.Errorf("no case expects limit %q", n)
		}
	}
}

// TestNumberGrammarAgainstStrconv compares data.ParseNumber with strconv and a
// grammar written out as a regular expression, over a table of awkward
// numbers and over the numbers the corpus carries.
func TestNumberGrammarAgainstStrconv(t *testing.T) {
	grammar := regexp.MustCompile(`^-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?$`)
	inputs := []string{
		"0", "-0", "1", "-1", "10", "01", "00", "-01", "1.", ".5", "-.5", "1.5", "-1.5", "0.5",
		"1e5", "1E5", "1e+5", "1e-5", "1e", "1e+", "e5", "1.e5", "1.5e5", "+1", "--1", "-", "",
		" 1", "1 ", "0x10", "1_000", "NaN", "Inf", "Infinity", "-Infinity", "1e999999999",
		"1e-999999999", "9223372036854775807", "9223372036854775808", "-9223372036854775808",
		"-9223372036854775809", "1.7976931348623157e308", "1e309", "5e-324", "1e-400",
		strings.Repeat("1234567890", 10), strings.Repeat("9", 5000), "1" + strings.Repeat("0", 400) + "e-400",
	}
	for _, c := range hostiletest.For(hostiletest.JSON) {
		if c.Tags != nil && c.HasTag("number") && c.Size <= 1<<20 {
			in := string(c.Input())
			inputs = append(inputs, strings.TrimSuffix(strings.TrimPrefix(in, "["), "]"))
		}
	}

	for _, s := range inputs {
		n, err := data.ParseNumber(s)
		ok := err == nil
		if want := grammar.MatchString(s); ok != want {
			t.Errorf("ParseNumber(%.40q) valid=%v, grammar says %v", s, ok, want)
			continue
		}
		if !ok {
			if !errors.Is(err, data.ErrNumberSyntax) {
				t.Errorf("ParseNumber(%.40q): error %v is not ErrNumberSyntax", s, err)
			}
			continue
		}
		if n.Text != s {
			t.Errorf("ParseNumber(%.40q) changed the text to %.40q", s, n.Text)
		}

		// Integers: Int64 agrees with strconv.ParseInt, error for error.
		if n.IsInteger() {
			got, gerr := n.Int64()
			want, werr := strconv.ParseInt(s, 10, 64)
			switch {
			case werr == nil && (gerr != nil || got != want):
				t.Errorf("Int64(%.40q) = %d, %v; strconv says %d", s, got, gerr, want)
			case werr != nil && !errors.Is(gerr, data.ErrNumberRange):
				t.Errorf("Int64(%.40q) = %d, %v; strconv says out of range", s, got, gerr)
			}
		}

		// Float64 agrees with strconv.ParseFloat when that is finite; when
		// strconv overflows, so must Float64, with ErrNumberRange.
		got, gerr := n.Float64()
		want, werr := strconv.ParseFloat(s, 64)
		switch {
		case werr == nil && (gerr != nil || got != want):
			t.Errorf("Float64(%.40q) = %v, %v; strconv says %v", s, got, gerr, want)
		case werr != nil && !math.IsInf(want, 0) && !math.IsInf(got, 0) && gerr == nil:
			t.Errorf("Float64(%.40q) = %v; strconv says %v", s, got, werr)
		case werr != nil && math.IsInf(want, 0) && !errors.Is(gerr, data.ErrNumberRange):
			t.Errorf("Float64(%.40q) = %v, %v; want ErrNumberRange on overflow", s, got, gerr)
		}
	}
}

// TestRuntimeDoesNotImportHarness: pkg/hostiletest is test support. No other
// non-test package in the module may import it, and the package itself must
// not import encoding/json or encoding/xml. The check is scoped to this
// package: other packages of the module (validator, ast) import
// encoding/json for their own reasons, and the format modules enforce the
// stdlib-only rule in their own repositories.
func TestRuntimeDoesNotImportHarness(t *testing.T) {
	const self = "github.com/shapestone/shape-core/pkg/hostiletest"
	root := filepath.Join("..", "..")
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			return nil // not this test's business
		}
		inSelf := strings.Contains(filepath.ToSlash(path), "pkg/hostiletest/")
		isTest := strings.HasSuffix(path, "_test.go")
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			switch {
			case p == self && !isTest:
				t.Errorf("%s imports %s outside tests", path, self)
			case inSelf && !isTest && (p == "encoding/json" || p == "encoding/xml"):
				t.Errorf("%s imports %s", path, p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
