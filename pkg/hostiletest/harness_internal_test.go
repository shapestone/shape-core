package hostiletest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/shapestone/shape-core/pkg/data"
)

// recorder is a reporter that remembers what execute reported.
type recorder struct {
	errs, skips, logs []string
}

func (r *recorder) Helper() {}
func (r *recorder) Errorf(f string, a ...any) {
	r.errs = append(r.errs, fmt.Sprintf(f, a...))
}
func (r *recorder) Skipf(f string, a ...any) {
	r.skips = append(r.skips, fmt.Sprintf(f, a...))
}
func (r *recorder) Logf(f string, a ...any) {
	r.logs = append(r.logs, fmt.Sprintf(f, a...))
}

func fixed(v data.Value, err error) ParseFunc {
	return func([]byte, Request) (data.Value, error) { return v, err }
}

var (
	errSyntax = &data.SyntaxError{Format: "x", Msg: "bad"}
	errDepth  = &data.LimitError{Limit: data.LimitDepth, Max: 1, Got: 2}
)

func testCase(want Expect) Case {
	return Case{Name: "c", Format: JSON, Size: 1, Input: func() []byte { return []byte("x") },
		Want: map[Profile]Expect{Default: want}}
}

// newTestConfig is newConfig without the environment: tests must not change
// behaviour because HOSTILETEST_STRICT is set when they run.
func newTestConfig(opts ...Option) *config {
	cfg := &config{
		unknown:    map[string]bool{},
		unknownFor: map[Profile]map[string]bool{},
		adopted:    map[string]bool{},
		diverges:   map[string]string{},
	}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

// shortTimes makes the timeout and grace period short for a test.
func shortTimes(t *testing.T) {
	t.Helper()
	caseTimeout, graceTime = 100*time.Millisecond, 100*time.Millisecond
	t.Cleanup(func() { caseTimeout, graceTime = 0, 5*time.Second })
}

func TestVerdict(t *testing.T) {
	tests := []struct {
		name string
		want Expect
		v    data.Value
		err  error
		ok   bool
	}{
		{"valid", valid, data.NullValue(), nil, true},
		{"valid but syntax error", valid, data.Value{}, errSyntax, false},
		{"syntax", syntax, data.Value{}, errSyntax, true},
		{"syntax but valid", syntax, data.NullValue(), nil, false},
		{"syntax but limit", syntax, data.Value{}, errDepth, false},
		{"limit any name", limit(), data.Value{}, errDepth, true},
		{"limit right name", limit(data.LimitDepth), data.Value{}, errDepth, true},
		{"limit wrong name", limit(data.LimitBytes), data.Value{}, errDepth, false},
		{"limit one of names", limit(data.LimitBytes, data.LimitDepth), data.Value{}, errDepth, true},
		{"limit but valid", limit(data.LimitDepth), data.NullValue(), nil, false},
		{"untyped error", syntax, data.Value{}, errors.New("boom"), false},
		{"wrapped syntax", syntax, data.Value{}, fmt.Errorf("ctx: %w", errSyntax), true},
		{"syntax or limit: syntax", Expect{Outcome: Syntax | Limit}, data.Value{}, errSyntax, true},
		{"syntax or limit: limit", Expect{Outcome: Syntax | Limit}, data.Value{}, errDepth, true},
		{"syntax or limit: valid", Expect{Outcome: Syntax | Limit}, data.NullValue(), nil, false},
		{"record accepts anything", record, data.Value{}, errors.New("boom"), true},
		{"timeout fails even when recording", record, data.Value{}, fmt.Errorf("parse: %w", context.DeadlineExceeded), false},
	}
	for _, tt := range tests {
		t.Run(strings.ReplaceAll(tt.name, " ", "_"), func(t *testing.T) {
			_, err := verdict(testCase(tt.want), tt.want, tt.v, tt.err)
			if (err == nil) != tt.ok {
				t.Errorf("verdict error = %v, want ok=%v", err, tt.ok)
			}
		})
	}
}

func TestVerdictKinds(t *testing.T) {
	_, err := verdict(testCase(valid), valid, data.Value{}, errSyntax)
	if kindOf(err) != kVerdict {
		t.Errorf("mismatch kind = %q", kindOf(err))
	}
	c := testCase(valid)
	c.Check = func(data.Value) error { return errors.New("wrong shape") }
	_, err = verdict(c, valid, data.NullValue(), nil)
	if kindOf(err) != kCheck || !strings.Contains(err.Error(), "wrong shape") {
		t.Errorf("check error = %v (kind %q)", err, kindOf(err))
	}
}

func TestExecuteKnownGaps(t *testing.T) {
	c := testCase(limit(data.LimitDepth))
	c.Known = map[Profile]string{Default: "not adopted"}
	path := []string{"T", "c", "default"}
	want := c.Want[Default]

	t.Run("failing known case is skipped", func(t *testing.T) {
		r := &recorder{}
		execute(r, path, c, Default, want, newTestConfig(), fixed(data.NullValue(), nil))
		if len(r.skips) != 1 || len(r.errs) != 0 || !strings.Contains(r.skips[0], "not adopted") {
			t.Errorf("skips=%q errs=%q", r.skips, r.errs)
		}
	})
	t.Run("passing known case is stale", func(t *testing.T) {
		r := &recorder{}
		execute(r, path, c, Default, want, newTestConfig(), fixed(data.Value{}, errDepth))
		if len(r.errs) != 1 || !strings.Contains(r.errs[0], "remove its Known entry") {
			t.Errorf("skips=%q errs=%q", r.skips, r.errs)
		}
	})
	t.Run("a panic is not excused by a known gap", func(t *testing.T) {
		r := &recorder{}
		boom := func([]byte, Request) (data.Value, error) { panic("index out of range") }
		execute(r, path, c, Default, want, newTestConfig(), boom)
		if len(r.skips) != 0 || len(r.errs) != 1 || !strings.Contains(r.errs[0], "different failure") {
			t.Errorf("skips=%q errs=%q", r.skips, r.errs)
		}
	})
	t.Run("strict ignores known", func(t *testing.T) {
		r := &recorder{}
		execute(r, path, c, Default, want, newTestConfig(Strict()), fixed(data.NullValue(), nil))
		if len(r.errs) != 1 || len(r.skips) != 0 {
			t.Errorf("skips=%q errs=%q", r.skips, r.errs)
		}
	})
	t.Run("WithoutKnown drops one case", func(t *testing.T) {
		r := &recorder{}
		execute(r, path, c, Default, want, newTestConfig(WithoutKnown("c")), fixed(data.NullValue(), nil))
		if len(r.errs) != 1 {
			t.Errorf("skips=%q errs=%q", r.skips, r.errs)
		}
	})
	t.Run("WithoutKnownFor drops one profile", func(t *testing.T) {
		both := c
		both.Want = map[Profile]Expect{Default: want, Wire: want}
		both.Known = map[Profile]string{Default: "x", Wire: "y"}
		cfg := newTestConfig(WithoutKnownFor(Wire, "c"))
		r := &recorder{}
		execute(r, path, both, Default, want, cfg, fixed(data.NullValue(), nil))
		if len(r.skips) != 1 || len(r.errs) != 0 {
			t.Errorf("default must stay known: skips=%q errs=%q", r.skips, r.errs)
		}
		r = &recorder{}
		execute(r, path, both, Wire, want, cfg, fixed(data.NullValue(), nil))
		if len(r.errs) != 1 || len(r.skips) != 0 {
			t.Errorf("wire must be strict: skips=%q errs=%q", r.skips, r.errs)
		}
	})
	t.Run("Adopted needs every allowed limit", func(t *testing.T) {
		multi := testCase(limit(data.LimitNodes, data.LimitMemory))
		multi.Known = map[Profile]string{Default: "not adopted"}
		w := multi.Want[Default]

		r := &recorder{}
		execute(r, path, multi, Default, w, newTestConfig(Adopted(data.LimitNodes)), fixed(data.NullValue(), nil))
		if len(r.skips) != 1 {
			t.Errorf("one of two adopted must stay known: skips=%q errs=%q", r.skips, r.errs)
		}
		r = &recorder{}
		execute(r, path, multi, Default, w, newTestConfig(Adopted(data.LimitNodes, data.LimitMemory)), fixed(data.NullValue(), nil))
		if len(r.errs) != 1 || len(r.skips) != 0 {
			t.Errorf("both adopted must be strict: skips=%q errs=%q", r.skips, r.errs)
		}
	})
	t.Run("SkipHeavyKnown does not run heavy known cases", func(t *testing.T) {
		heavy := c
		heavy.Tags = []string{TagHeavy}
		ran := false
		parse := func([]byte, Request) (data.Value, error) { ran = true; return data.Value{}, errDepth }
		r := &recorder{}
		execute(r, path, heavy, Default, want, newTestConfig(SkipHeavyKnown()), parse)
		if ran || len(r.skips) != 1 {
			t.Errorf("ran=%v skips=%q errs=%q", ran, r.skips, r.errs)
		}
	})
	t.Run("unknown failure is an error", func(t *testing.T) {
		r := &recorder{}
		execute(r, path, testCase(valid), Default, valid, newTestConfig(), fixed(data.Value{}, errSyntax))
		if len(r.errs) != 1 {
			t.Errorf("errs=%q", r.errs)
		}
	})
	t.Run("max size skips", func(t *testing.T) {
		big := testCase(valid)
		big.Size = 100
		r := &recorder{}
		execute(r, path, big, Default, valid, newTestConfig(MaxSize(10)), fixed(data.Value{}, errSyntax))
		if len(r.skips) != 1 || len(r.errs) != 0 {
			t.Errorf("skips=%q errs=%q", r.skips, r.errs)
		}
	})
}

func TestExecuteRequest(t *testing.T) {
	var got Request
	parse := func(_ []byte, req Request) (data.Value, error) {
		got = req
		return data.NullValue(), nil
	}
	c := testCase(valid)
	c.Want = map[Profile]Expect{Wire: valid}
	r := &recorder{}
	execute(r, []string{"T"}, c, Wire, valid, newTestConfig(), parse)
	if got.Profile != Wire || !got.Policy.RejectDuplicateKeys || got.Limits.MaxBytes != 64<<20 {
		t.Errorf("request = %+v", got)
	}
	if got.Limits.Context == nil {
		t.Error("harness must set Limits.Context")
	}
}

func TestExecutePanicIsCaught(t *testing.T) {
	r := &recorder{}
	boom := func([]byte, Request) (data.Value, error) { panic("index out of range [5] with length 3") }
	execute(r, []string{"T"}, testCase(valid), Default, valid, newTestConfig(), boom)
	if len(r.errs) != 1 || !strings.Contains(r.errs[0], "panicked") || !strings.Contains(r.errs[0], "index out of range") {
		t.Errorf("errs=%q", r.errs)
	}
}

func TestExecuteHangIsFailed(t *testing.T) {
	shortTimes(t)
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	hang := func([]byte, Request) (data.Value, error) {
		<-release // ignores Limits.Context
		return data.NullValue(), nil
	}
	r := &recorder{}
	start := time.Now()
	execute(r, []string{"T"}, testCase(valid), Default, valid, newTestConfig(), hang)
	if len(r.errs) != 1 || !strings.Contains(r.errs[0], "did not return") {
		t.Errorf("errs=%q", r.errs)
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("took %s", d)
	}
}

func TestExecuteMemoryCaps(t *testing.T) {
	if testing.Short() {
		t.Skip("allocates 300 MiB")
	}
	tight := func(c Case) Case {
		c.Want = map[Profile]Expect{Tight: valid}
		return c
	}
	t.Run("peak over 3x MaxMemory fails", func(t *testing.T) {
		var keep []byte
		hog := func([]byte, Request) (data.Value, error) {
			keep = make([]byte, 300<<20) // over the 256 MiB floor of the cap
			for i := 0; i < len(keep); i += 4096 {
				keep[i] = 1
			}
			return data.NullValue(), nil
		}
		r := &recorder{}
		execute(r, []string{"T"}, tight(testCase(valid)), Tight, valid, newTestConfig(), hog)
		runtime.KeepAlive(keep)
		if len(r.errs) != 1 || !strings.Contains(r.errs[0], "over the 256 MiB cap") {
			t.Errorf("errs=%q", r.errs)
		}
	})
	t.Run("live heap over MaxMemory fails", func(t *testing.T) {
		// Keeps 20 MiB in the result, far over Tight's 1 MiB budget but
		// well under the peak cap: only the post-GC check can see it.
		big := func([]byte, Request) (data.Value, error) {
			return data.StringValue(string(make([]byte, 20<<20))), nil
		}
		r := &recorder{}
		execute(r, []string{"T"}, tight(testCase(valid)), Tight, valid, newTestConfig(), big)
		if len(r.errs) != 1 || !strings.Contains(r.errs[0], "live heap") {
			t.Errorf("errs=%q", r.errs)
		}
	})
	t.Run("garbage that is not kept passes", func(t *testing.T) {
		churn := func([]byte, Request) (data.Value, error) {
			for i := 0; i < 20; i++ {
				b := make([]byte, 4<<20)
				b[0] = 1
			}
			return data.NullValue(), nil
		}
		r := &recorder{}
		execute(r, []string{"T"}, tight(testCase(valid)), Tight, valid, newTestConfig(), churn)
		if len(r.errs) != 0 {
			t.Errorf("errs=%q", r.errs)
		}
	})
}

func TestPeakCap(t *testing.T) {
	if got := peakCap(Limits(Wire)); got != 3<<30 {
		t.Errorf("wire cap = %d, want 3 GiB", got)
	}
	if got := peakCap(Limits(Tight)); got != minPeak {
		t.Errorf("tight cap = %d, want the %d floor", got, minPeak)
	}
	if got := peakCap(data.Limits{}); got != 2<<30 {
		t.Errorf("default cap = %d, want 2 GiB", got)
	}
	if got := peakCap(data.Limits{MaxMemory: 1 << 30}); got != 3<<30 {
		t.Errorf("cap for 1 GiB = %d", got)
	}
}

func TestClipKeepsRunesWhole(t *testing.T) {
	for n := 118; n <= 122; n++ {
		in := strings.Repeat("a", n-118) + strings.Repeat("é", 100)
		if out := clip(in); !strings.HasSuffix(out, "...") || strings.ContainsRune(out, '�') {
			t.Errorf("clip split a rune: %q", out)
		}
	}
	if clip("short") != "short" {
		t.Error("short strings pass through")
	}
}

func TestDifferential(t *testing.T) {
	one, two := data.IntValue(1), data.IntValue(2)
	tests := []struct {
		name  string
		v     data.Value
		perr  error
		ref   data.Value
		rerr  error
		fails bool
	}{
		{"equal", one, nil, one, nil, false},
		{"values differ", one, nil, two, nil, true},
		{"parser rejects, reference accepts", data.Value{}, errSyntax, one, nil, true},
		{"parser accepts, reference rejects", one, nil, data.Value{}, errors.New("no"), true},
		{"both reject", data.Value{}, errSyntax, data.Value{}, errors.New("no"), false},
	}
	for _, tt := range tests {
		t.Run(strings.ReplaceAll(tt.name, " ", "_"), func(t *testing.T) {
			cfg := newTestConfig(WithReference(func([]byte) (data.Value, error) { return tt.ref, tt.rerr }))
			err := differential(cfg, []byte("x"), tt.v, tt.perr)
			if (err != nil) != tt.fails {
				t.Errorf("err = %v, want failure=%v", err, tt.fails)
			}
		})
	}
}

func TestExecuteDifferentialDivergence(t *testing.T) {
	c := testCase(valid)
	c.Diff = true
	ref := WithReference(func([]byte) (data.Value, error) { return data.IntValue(2), nil })
	parse := fixed(data.IntValue(1), nil)

	r := &recorder{}
	execute(r, []string{"T"}, c, Default, valid, newTestConfig(ref), parse)
	if len(r.errs) != 1 {
		t.Errorf("a difference must fail: errs=%q", r.errs)
	}

	r = &recorder{}
	execute(r, []string{"T"}, c, Default, valid, newTestConfig(ref, Diverges("c", "policy")), parse)
	if len(r.errs) != 0 || len(r.logs) != 1 || !strings.Contains(r.logs[0], "policy") {
		t.Errorf("a documented divergence is logged: errs=%q logs=%q", r.errs, r.logs)
	}

	// Without Diff the reference is not consulted.
	c.Diff = false
	r = &recorder{}
	execute(r, []string{"T"}, c, Default, valid, newTestConfig(ref), parse)
	if len(r.errs) != 0 {
		t.Errorf("reference used on a case without Diff: errs=%q", r.errs)
	}
}

// The isolation tests re-execute this test binary. Each parse function here
// only ever runs in the child: the parent never calls it for an isolated case.

func overflow(depth int) int {
	var pad [64]byte
	pad[depth%64] = byte(depth)
	return overflow(depth+1) + int(pad[0]) //nolint:staticcheck // deliberate stack overflow
}

type isolatedCase struct {
	parse   ParseFunc
	want    Expect
	profile Profile
}

func isolatedCases() map[string]isolatedCase {
	return map[string]isolatedCase{
		"passes":   {fixed(data.NullValue(), nil), valid, Default},
		"mismatch": {fixed(data.Value{}, errSyntax), valid, Default},
		"panics": {func([]byte, Request) (data.Value, error) {
			panic("nil map write")
		}, valid, Default},
		"stack-overflow": {func([]byte, Request) (data.Value, error) {
			debug.SetMaxStack(1 << 20) // fail fast instead of at 1 GB
			return data.IntValue(int64(overflow(0))), nil
		}, valid, Default},
		"exits": {func([]byte, Request) (data.Value, error) {
			os.Exit(3)
			return data.Value{}, nil
		}, valid, Default},
		// Allocates until the child's watchdog stops it at Tight's 256 MiB
		// cap, long before the 2 GiB it would take otherwise.
		"runaway": {func([]byte, Request) (data.Value, error) {
			var keep [][]byte
			for i := 0; i < 100; i++ {
				b := make([]byte, 32<<20)
				for j := 0; j < len(b); j += 4096 {
					b[j] = 1
				}
				keep = append(keep, b)
				time.Sleep(5 * time.Millisecond)
			}
			runtime.KeepAlive(keep)
			return data.NullValue(), nil
		}, valid, Tight},
	}
}

func isolatedCaseOf(name string, tc isolatedCase) Case {
	return Case{Name: name, Format: JSON, Tags: []string{TagCrash}, Size: 1,
		Input: func() []byte { return []byte("x") },
		Want:  map[Profile]Expect{tc.profile: tc.want}}
}

func TestIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("re-executes the test binary")
	}
	for name, tc := range isolatedCases() {
		c := isolatedCaseOf(name, tc)
		t.Run(name, func(t *testing.T) {
			t.Run(string(tc.profile), func(t *testing.T) {
				cfg := newConfig(nil)
				cfg.strict = false
				rec := &recorder{}
				execute(rec, strings.Split(t.Name(), "/"), c, tc.profile, tc.want, cfg, tc.parse)

				if cfg.childMode {
					// Child: report the raw verdict by failing like Run would.
					for _, e := range rec.errs {
						t.Error(e)
					}
					return
				}
				wantErr := map[string]string{
					"mismatch":       "want valid, got syntax error",
					"panics":         "panicked",
					"stack-overflow": "fatal stack overflow",
					"exits":          "child process failed",
					"runaway":        "child process stopped",
				}[name]
				switch {
				case wantErr == "" && (len(rec.errs) != 0 || len(rec.skips) != 0):
					t.Errorf("errs=%q skips=%q", rec.errs, rec.skips)
				case wantErr != "" && (len(rec.errs) != 1 || !strings.Contains(rec.errs[0], wantErr)):
					t.Errorf("want an error containing %q, got errs=%q skips=%q", wantErr, rec.errs, rec.skips)
				}
			})
		})
	}
}

// TestIsolationKnown: a known crash or verdict failure is skipped and the
// test process survives; a known case whose child panics is not excused.
func TestIsolationKnown(t *testing.T) {
	if testing.Short() {
		t.Skip("re-executes the test binary")
	}
	skipped := map[string]bool{"stack-overflow": true, "mismatch": true, "runaway": true, "exits": true}
	for name := range skipped {
		tc := isolatedCases()[name]
		c := isolatedCaseOf(name, tc)
		c.Known = map[Profile]string{tc.profile: gapStack}
		t.Run("known-"+name, func(t *testing.T) {
			t.Run(string(tc.profile), func(t *testing.T) {
				cfg := newConfig(nil)
				cfg.strict = false
				rec := &recorder{}
				execute(rec, strings.Split(t.Name(), "/"), c, tc.profile, tc.want, cfg, tc.parse)
				if cfg.childMode {
					for _, e := range rec.errs {
						t.Error(e)
					}
					return
				}
				if len(rec.skips) != 1 || len(rec.errs) != 0 {
					t.Errorf("skips=%q errs=%q", rec.skips, rec.errs)
				}
			})
		})
	}

	tc := isolatedCases()["panics"]
	c := isolatedCaseOf("panics", tc)
	c.Known = map[Profile]string{Default: "x"}
	t.Run("known-panics", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			cfg := newConfig(nil)
			cfg.strict = false
			rec := &recorder{}
			execute(rec, strings.Split(t.Name(), "/"), c, Default, tc.want, cfg, tc.parse)
			if cfg.childMode {
				for _, e := range rec.errs {
					t.Error(e)
				}
				return
			}
			if len(rec.skips) != 0 || len(rec.errs) != 1 || !strings.Contains(rec.errs[0], "different failure") {
				t.Errorf("a panic must not be excused: skips=%q errs=%q", rec.skips, rec.errs)
			}
		})
	})
}

func TestStartPeakSamplerReportsGrowth(t *testing.T) {
	if testing.Short() {
		t.Skip("allocates 64 MiB")
	}
	runtime.GC()
	stop := startPeakSampler(heapBytes(), 0, nil)
	buf := make([]byte, 64<<20)
	for i := 0; i < len(buf); i += 4096 {
		buf[i] = 1
	}
	got := stop()
	if got < 32<<20 {
		t.Errorf("peak growth = %d, want at least 32 MiB", got)
	}
	runtime.KeepAlive(buf)
}

func TestStartPeakSamplerCallsOnExceed(t *testing.T) {
	if testing.Short() {
		t.Skip("allocates 64 MiB")
	}
	runtime.GC()
	called := make(chan struct{}, 1)
	stop := startPeakSampler(heapBytes(), 16<<20, func() { called <- struct{}{} })
	buf := make([]byte, 64<<20)
	for i := 0; i < len(buf); i += 4096 {
		buf[i] = 1
	}
	time.Sleep(20 * time.Millisecond)
	stop()
	runtime.KeepAlive(buf)
	select {
	case <-called:
	default:
		t.Error("onExceed was not called")
	}
}
