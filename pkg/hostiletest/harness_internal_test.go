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
	return func(context.Context, []byte, Profile, data.Limits) (data.Value, error) { return v, err }
}

var (
	errSyntax = &data.SyntaxError{Format: "x", Msg: "bad"}
	errDepth  = &data.LimitError{Limit: data.LimitDepth, Max: 1, Got: 2}
)

func testCase(want Expect) Case {
	return Case{Name: "c", Format: JSON, Size: 1, Input: func() []byte { return []byte("x") },
		Want: map[Profile]Expect{Default: want}}
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
		{"timeout fails even when recording", record, data.Value{}, context.DeadlineExceeded, false},
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

func TestVerdictRunsCheck(t *testing.T) {
	c := testCase(valid)
	c.Check = func(data.Value) error { return errors.New("wrong shape") }
	if _, err := verdict(c, valid, data.NullValue(), nil); err == nil || !strings.Contains(err.Error(), "wrong shape") {
		t.Errorf("Check error not reported: %v", err)
	}
}

// newTestConfig is newConfig without the environment: tests must not change
// behaviour because HOSTILETEST_STRICT is set when they run.
func newTestConfig(opts ...Option) *config {
	cfg := &config{unknown: map[string]bool{}, diverges: map[string]string{}}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
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
	t.Run("unknown failure is an error", func(t *testing.T) {
		plain := testCase(valid)
		r := &recorder{}
		execute(r, path, plain, Default, valid, newTestConfig(), fixed(data.Value{}, errSyntax))
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

func TestExecuteMemoryCap(t *testing.T) {
	c := testCase(valid)
	r := &recorder{}
	var keep []byte
	hog := func(context.Context, []byte, Profile, data.Limits) (data.Value, error) {
		keep = make([]byte, 300<<20) // over Tight's 256 MiB cap
		for i := 0; i < len(keep); i += 4096 {
			keep[i] = 1
		}
		return data.NullValue(), nil
	}
	c.Want = map[Profile]Expect{Tight: valid}
	execute(r, []string{"T"}, c, Tight, valid, newTestConfig(), hog)
	if len(r.errs) != 1 || !strings.Contains(r.errs[0], "heap grew") {
		t.Errorf("errs=%q logs=%q", r.errs, r.logs)
	}
	runtime.KeepAlive(keep)
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

func isolatedCases() map[string]struct {
	parse ParseFunc
	want  Expect
} {
	return map[string]struct {
		parse ParseFunc
		want  Expect
	}{
		"passes":   {fixed(data.NullValue(), nil), valid},
		"mismatch": {fixed(data.Value{}, errSyntax), valid},
		"stack-overflow": {func(context.Context, []byte, Profile, data.Limits) (data.Value, error) {
			debug.SetMaxStack(1 << 20) // fail fast instead of at 1 GB
			return data.IntValue(int64(overflow(0))), nil
		}, valid},
		"exits": {func(context.Context, []byte, Profile, data.Limits) (data.Value, error) {
			os.Exit(3)
			return data.Value{}, nil
		}, valid},
	}
}

func TestIsolation(t *testing.T) {
	for name, tc := range isolatedCases() {
		c := Case{Name: name, Format: JSON, Tags: []string{TagCrash}, Size: 1,
			Input: func() []byte { return []byte("x") },
			Want:  map[Profile]Expect{Default: tc.want}}
		t.Run(name, func(t *testing.T) {
			t.Run("default", func(t *testing.T) {
				cfg := newConfig(nil)
				cfg.strict = false
				rec := &recorder{}
				execute(rec, strings.Split(t.Name(), "/"), c, Default, tc.want, cfg, tc.parse)

				if cfg.childMode {
					// Child: report the raw verdict by failing like Run would.
					for _, e := range rec.errs {
						t.Error(e)
					}
					return
				}
				switch name {
				case "passes":
					if len(rec.errs) != 0 || len(rec.skips) != 0 {
						t.Errorf("errs=%q skips=%q", rec.errs, rec.skips)
					}
				case "mismatch":
					if len(rec.errs) != 1 || !strings.Contains(rec.errs[0], "want valid, got syntax error") {
						t.Errorf("errs=%q", rec.errs)
					}
				case "stack-overflow":
					if len(rec.errs) != 1 || !strings.Contains(rec.errs[0], "fatal stack overflow") {
						t.Errorf("errs=%q", rec.errs)
					}
				case "exits":
					if len(rec.errs) != 1 || !strings.Contains(rec.errs[0], "child process failed") {
						t.Errorf("errs=%q", rec.errs)
					}
				}
			})
		})
	}
}

// TestIsolationKnownCrash: a known crash is skipped, and the test process
// itself survives.
func TestIsolationKnownCrash(t *testing.T) {
	tc := isolatedCases()["stack-overflow"]
	c := Case{Name: "stack-overflow", Format: JSON, Tags: []string{TagCrash}, Size: 1,
		Input: func() []byte { return []byte("x") },
		Want:  map[Profile]Expect{Default: tc.want},
		Known: map[Profile]string{Default: gapStack}}
	t.Run("known", func(t *testing.T) {
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
			if len(rec.skips) != 1 || len(rec.errs) != 0 {
				t.Errorf("skips=%q errs=%q", rec.skips, rec.errs)
			}
		})
	})
}

func TestStartPeakSamplerReportsGrowth(t *testing.T) {
	stop := startPeakSampler()
	buf := make([]byte, 64<<20)
	for i := 0; i < len(buf); i += 4096 {
		buf[i] = 1
	}
	got := stop()
	if got < 32<<20 {
		t.Errorf("peak growth = %d, want at least 32 MiB", got)
	}
	_ = buf[0]
}
