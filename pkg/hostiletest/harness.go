package hostiletest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime/metrics"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shapestone/shape-core/pkg/data"
)

// ParseFunc parses in under the limits l. ctx is also l.Context. Profile is
// passed so a parser can follow the profile's policy, such as rejecting
// duplicate keys under Wire. Syntax errors must be *data.SyntaxError and limit
// violations *data.LimitError.
type ParseFunc func(ctx context.Context, in []byte, p Profile, l data.Limits) (data.Value, error)

// RefFunc is a reference implementation for the differential comparison,
// typically the standard library wrapped in test code. A non-nil error means
// the reference rejects the input.
type RefFunc func(in []byte) (data.Value, error)

// Option configures Run.
type Option func(*config)

type config struct {
	ref       RefFunc
	strict    bool
	maxSize   int64
	unknown   map[string]bool
	diverges  map[string]string
	childMode bool
}

// WithReference enables the differential comparison for cases with Diff set.
// Where the reference accepts, the parser must accept with an equal value;
// where it rejects, the parser must reject. The reference never sees an input
// over 1 MiB or an isolated case.
func WithReference(ref RefFunc) Option { return func(c *config) { c.ref = ref } }

// WithoutKnown drops the Known entries of the named cases: use it when a
// format has adopted pkg/data for those cases.
func WithoutKnown(names ...string) Option {
	return func(c *config) {
		for _, n := range names {
			c.unknown[n] = true
		}
	}
}

// Strict ignores every Known entry, as HOSTILETEST_STRICT=1 does.
func Strict() Option { return func(c *config) { c.strict = true } }

// MaxSize skips cases with an input larger than n bytes, as
// HOSTILETEST_MAXSIZE does.
func MaxSize(n int64) Option { return func(c *config) { c.maxSize = n } }

// Diverges documents that a parser knowingly differs from the reference on
// the named case. The difference is logged instead of failing.
func Diverges(name, reason string) Option {
	return func(c *config) { c.diverges[name] = reason }
}

const (
	envChild   = "HOSTILETEST_CHILD"
	envStrict  = "HOSTILETEST_STRICT"
	envMaxSize = "HOSTILETEST_MAXSIZE"
	childOK    = "HOSTILETEST_OK"
)

func newConfig(opts []Option) *config {
	c := &config{
		unknown:   map[string]bool{},
		diverges:  map[string]string{},
		strict:    os.Getenv(envStrict) == "1",
		childMode: os.Getenv(envChild) == "1",
	}
	if n, err := strconv.ParseInt(os.Getenv(envMaxSize), 10, 64); err == nil && n > 0 {
		c.maxSize = n
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Run runs every case of format f against parse, one subtest per case and
// profile. See the package documentation.
func Run(t *testing.T, f Format, parse ParseFunc, opts ...Option) {
	t.Helper()
	runCases(t, For(f), parse, opts...)
}

func runCases(t *testing.T, cases []Case, parse ParseFunc, opts ...Option) {
	t.Helper()
	cfg := newConfig(opts)
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			for _, p := range Profiles() {
				want, ok := c.Want[p]
				if !ok {
					continue
				}
				t.Run(string(p), func(t *testing.T) {
					if testing.Short() && c.Heavy() {
						t.Skip("heavy case skipped under -short")
					}
					execute(t, strings.Split(t.Name(), "/"), c, p, want, cfg, parse)
				})
			}
		})
	}
}

// reporter is the part of *testing.T that execute uses, so the harness can be
// tested with a recorder.
type reporter interface {
	Helper()
	Errorf(format string, args ...any)
	Skipf(format string, args ...any)
	Logf(format string, args ...any)
}

// execute runs one case under one profile and reports through r. path is the
// full subtest name split on "/", which a child process uses to select it.
func execute(r reporter, path []string, c Case, p Profile, want Expect, cfg *config, parse ParseFunc) {
	r.Helper()
	if cfg.maxSize > 0 && c.Size > cfg.maxSize {
		r.Skipf("input of %d bytes is over the %d byte cap", c.Size, cfg.maxSize)
		return
	}

	if cfg.childMode {
		// Raw verdict: the parent applies Known and interprets the exit.
		desc, err := runInProcess(c, p, want, cfg, parse)
		if err != nil {
			r.Errorf("%v", err)
			return
		}
		r.Logf("%s", desc)
		fmt.Println(childOK) // the parent looks for this marker
		return
	}

	var (
		desc string
		err  error
	)
	if c.Isolated() {
		err = runIsolated(path, c)
	} else {
		desc, err = runInProcess(c, p, want, cfg, parse)
	}

	known := c.Known[p]
	if cfg.strict || cfg.unknown[c.Name] {
		known = ""
	}
	switch {
	case known != "" && err != nil:
		r.Skipf("known gap: %s: %v", known, err)
	case known != "":
		r.Errorf("known gap now passes, remove its Known entry: %s", known)
	case err != nil:
		r.Errorf("%v", err)
	case desc != "":
		r.Logf("%s", desc)
	}
}

// memoryCap is the most the live heap may grow during one parse, by profile.
// It is deliberately looser than the Limits.MaxMemory budget: the Guard
// estimate ignores slice growth and garbage the collector has not yet freed,
// and a correct Guard-limited parse of 64 MiB of small ints under Wire
// measured a 2.1 GiB peak against the 1 GiB budget. The cap exists to catch
// allocation the Guard never sees, such as the 3 GiB of an unguarded parse.
func memoryCap(p Profile) uint64 {
	switch p {
	case Wire:
		return 5 << 29 // 2.5 GiB
	case Tight:
		return 256 << 20
	}
	return 2 << 30
}

func timeout(c Case) time.Duration {
	if c.Heavy() {
		return 60 * time.Second
	}
	return 10 * time.Second
}

// runInProcess parses the case input in this process and checks the verdict,
// the value, the heap growth and the differential. desc describes the verdict.
func runInProcess(c Case, p Profile, want Expect, cfg *config, parse ParseFunc) (desc string, err error) {
	in := c.Input()
	ctx, cancel := context.WithTimeout(context.Background(), timeout(c))
	defer cancel()
	limits := Limits(p)
	limits.Context = ctx

	peak := startPeakSampler()
	v, perr := parse(ctx, in, p, limits)
	grown := peak()

	if grown > memoryCap(p) {
		return "", fmt.Errorf("heap grew by %d MiB, over the %d MiB cap for %s", grown>>20, memoryCap(p)>>20, p)
	}
	desc, err = verdict(c, want, v, perr)
	if err != nil {
		return "", err
	}
	if cfg.ref != nil && c.Diff && p == Default && want.Outcome&(Valid|Syntax) != 0 && c.Size <= heavySize {
		if derr := differential(cfg, in, v, perr); derr != nil {
			if why, ok := cfg.diverges[c.Name]; ok {
				return "diverges from the reference (" + why + "): " + derr.Error(), nil
			}
			return "", derr
		}
	}
	return desc, nil
}

// verdict compares what the parser did with what the case expects.
func verdict(c Case, want Expect, v data.Value, err error) (string, error) {
	var le *data.LimitError
	var got string
	var ok bool
	switch {
	case err == nil:
		got, ok = "valid", want.Outcome&Valid != 0
		if ok && c.Check != nil {
			if cerr := c.Check(v); cerr != nil {
				return "", fmt.Errorf("value: %w", cerr)
			}
		}
	case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
		return "", fmt.Errorf("parse did not finish in time: %w", err)
	case errors.Is(err, data.ErrSyntax):
		got, ok = "syntax error", want.Outcome&Syntax != 0
	case errors.As(err, &le):
		got, ok = "limit error "+le.Limit, want.Outcome&Limit != 0
		if ok && len(want.Limits) > 0 {
			ok = false
			for _, name := range want.Limits {
				ok = ok || name == le.Limit
			}
		}
	default:
		got = fmt.Sprintf("untyped error (%T)", err)
	}
	if want.Outcome&Record != 0 {
		return "recorded: " + got, nil
	}
	if !ok {
		return "", fmt.Errorf("want %s, got %s: %v", describe(want), got, clip(fmt.Sprint(err)))
	}
	return got, nil
}

func describe(e Expect) string {
	var parts []string
	if e.Outcome&Valid != 0 {
		parts = append(parts, "valid")
	}
	if e.Outcome&Syntax != 0 {
		parts = append(parts, "syntax error")
	}
	if e.Outcome&Limit != 0 {
		s := "limit error"
		if len(e.Limits) > 0 {
			s += " " + strings.Join(e.Limits, "|")
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, " or ")
}

// differential compares the parser's result with the reference's.
func differential(cfg *config, in []byte, v data.Value, perr error) error {
	rv, rerr := cfg.ref(in)
	switch {
	case rerr == nil && perr == nil:
		if !v.Equal(rv) {
			return fmt.Errorf("value differs from reference: got %s, reference %s", clip(v.String()), clip(rv.String()))
		}
	case rerr == nil:
		return fmt.Errorf("parser rejects input the reference accepts: %v", clip(perr.Error()))
	case perr == nil:
		return fmt.Errorf("parser accepts input the reference rejects: %v", clip(rerr.Error()))
	}
	return nil
}

// startPeakSampler samples live heap bytes every few milliseconds and returns
// a func that stops sampling and reports the peak growth since the call.
func startPeakSampler() func() uint64 {
	const name = "/memory/classes/heap/objects:bytes"
	read := func() uint64 {
		s := []metrics.Sample{{Name: name}}
		metrics.Read(s)
		if s[0].Value.Kind() != metrics.KindUint64 {
			return 0
		}
		return s[0].Value.Uint64()
	}
	base := read()
	var peak atomic.Uint64
	stop, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		tick := time.NewTicker(2 * time.Millisecond)
		defer tick.Stop()
		for {
			if n := read(); n > peak.Load() {
				peak.Store(n)
			}
			select {
			case <-stop:
				return
			case <-tick.C:
			}
		}
	}()
	return func() uint64 {
		close(stop)
		<-done
		if n := read(); n > peak.Load() {
			peak.Store(n)
		}
		if p := peak.Load(); p > base {
			return p - base
		}
		return 0
	}
}

// runIsolated runs one case in a re-executed copy of the test binary, so a
// fatal crash fails this case alone. The child selects the case by its
// subtest path and runs it in-process, printing childOK when it passed.
func runIsolated(path []string, c Case) error {
	parts := make([]string, len(path))
	for i, s := range path {
		parts[i] = "^" + regexp.QuoteMeta(s) + "$"
	}
	limit := timeout(c) + 15*time.Second
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()

	//nolint:gosec // re-executes this test binary with fixed flags
	cmd := exec.CommandContext(ctx, os.Args[0],
		"-test.run="+strings.Join(parts, "/"),
		"-test.count=1",
		"-test.v",
		"-test.timeout="+limit.String(),
	)
	cmd.Env = append(os.Environ(), envChild+"=1")
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	text := out.String()

	switch {
	case err == nil && strings.Contains(text, childOK):
		return nil
	case err == nil:
		return fmt.Errorf("child process did not run the case (no %s marker): %s", childOK, tail(text))
	case ctx.Err() != nil:
		return fmt.Errorf("child process timed out after %s", limit)
	case strings.Contains(text, "stack overflow") || strings.Contains(text, "goroutine stack exceeds"):
		return errors.New("process crash: fatal stack overflow")
	case strings.Contains(text, "out of memory") || strings.Contains(err.Error(), "killed"):
		return fmt.Errorf("process killed, probably out of memory (%v)", err)
	}
	return fmt.Errorf("child process failed (%v): %s", err, tail(text))
}

func tail(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 1500 {
		s = "..." + s[len(s)-1500:]
	}
	return s
}

func clip(s string) string {
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}

func errf(format string, args ...any) error { return fmt.Errorf(format, args...) }
