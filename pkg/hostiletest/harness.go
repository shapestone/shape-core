package hostiletest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"runtime/debug"
	"runtime/metrics"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/shapestone/shape-core/pkg/data"
)

// Policy holds the parsing choices a profile makes that data.Limits cannot
// express. An adapter maps these to the format's own options; it should not
// switch on Request.Profile, because Profile is an open set that can gain
// values.
type Policy struct {
	// RejectDuplicateKeys means a repeated key in one map must fail the
	// parse with a *data.SyntaxError. Wire sets it.
	RejectDuplicateKeys bool
}

// PolicyFor returns the Policy of p. Unknown profiles get the zero Policy.
func PolicyFor(p Profile) Policy { return Policy{RejectDuplicateKeys: p == Wire} }

// Request is everything a parse needs besides the input. New fields can be
// added without breaking ParseFuncs.
type Request struct {
	Profile Profile
	// Limits are the limits to enforce. Limits.Context is set by the
	// harness; the parser must honour it.
	Limits data.Limits
	Policy Policy
}

// RequestFor returns the Request the harness builds for p.
func RequestFor(p Profile) Request {
	return Request{Profile: p, Limits: Limits(p), Policy: PolicyFor(p)}
}

// ParseFunc parses in as req describes. It must report malformed input as
// *data.SyntaxError and limit violations as *data.LimitError, and must stop
// when req.Limits.Context is done: a parse that ignores it is failed after a
// grace period, but its goroutine cannot be stopped.
type ParseFunc func(in []byte, req Request) (data.Value, error)

// RefFunc is a reference implementation for the differential comparison,
// typically the standard library wrapped in test code. A non-nil error means
// the reference rejects the input.
type RefFunc func(in []byte) (data.Value, error)

// Option configures Run.
type Option func(*config)

type config struct {
	ref            RefFunc
	strict         bool
	maxSize        int64
	skipKnownHeavy bool
	unknown        map[string]bool
	unknownFor     map[Profile]map[string]bool
	adopted        map[string]bool
	diverges       map[string]string
	childMode      bool
}

// WithReference enables the differential comparison for cases with Diff set.
// Where the reference accepts, the parser must accept with an equal value;
// where it rejects, the parser must reject. The reference never sees an input
// over 1 MiB or an isolated case.
func WithReference(ref RefFunc) Option { return func(c *config) { c.ref = ref } }

// WithoutKnown drops the Known entries of the named cases under every
// profile.
func WithoutKnown(names ...string) Option {
	return func(c *config) {
		for _, n := range names {
			c.unknown[n] = true
		}
	}
}

// WithoutKnownFor drops the Known entries of the named cases under one
// profile only.
func WithoutKnownFor(p Profile, names ...string) Option {
	return func(c *config) {
		if c.unknownFor[p] == nil {
			c.unknownFor[p] = map[string]bool{}
		}
		for _, n := range names {
			c.unknownFor[p][n] = true
		}
	}
}

// Adopted declares that the format now enforces the named limits
// (data.LimitDepth and friends). The Known entry of any case that expects
// only a limit error drops as soon as any limit name the case allows is
// adopted: a case that allows several names (for example nodes, keys-per-map
// or memory) accepts whichever limit fires first, so one adopted limit can
// satisfy it. A format that has adopted only some of those limits and still
// fails such a case should use WithoutKnownFor to opt in case by case.
func Adopted(limits ...string) Option {
	return func(c *config) {
		for _, l := range limits {
			c.adopted[l] = true
		}
	}
}

// Strict ignores every Known entry, as HOSTILETEST_STRICT=1 does.
func Strict() Option { return func(c *config) { c.strict = true } }

// SkipHeavyKnown skips, without running, Known cases that are heavy or
// isolated, as HOSTILETEST_SKIPKNOWN=1 does. Use it on runners where even a
// contained 3 GiB parse is too much.
func SkipHeavyKnown() Option { return func(c *config) { c.skipKnownHeavy = true } }

// MaxSize skips cases with an input larger than n bytes, as
// HOSTILETEST_MAXSIZE does.
func MaxSize(n int64) Option { return func(c *config) { c.maxSize = n } }

// Diverges documents that a parser knowingly differs from the reference on
// the named case. The difference is logged instead of failing.
func Diverges(name, reason string) Option {
	return func(c *config) { c.diverges[name] = reason }
}

const (
	envChild    = "HOSTILETEST_CHILD"
	envStrict   = "HOSTILETEST_STRICT"
	envMaxSize  = "HOSTILETEST_MAXSIZE"
	envSkipKnwn = "HOSTILETEST_SKIPKNOWN"

	// Markers a child process prints for its parent.
	childOK     = "HOSTILETEST_OK"
	childFail   = "HOSTILETEST_FAIL"
	childMemCap = "HOSTILETEST_MEMCAP"
)

func newConfig(opts []Option) *config {
	c := &config{
		unknown:        map[string]bool{},
		unknownFor:     map[Profile]map[string]bool{},
		adopted:        map[string]bool{},
		diverges:       map[string]string{},
		strict:         os.Getenv(envStrict) == "1",
		skipKnownHeavy: os.Getenv(envSkipKnwn) == "1",
		childMode:      os.Getenv(envChild) == "1",
	}
	if n, err := strconv.ParseInt(os.Getenv(envMaxSize), 10, 64); err == nil && n > 0 {
		c.maxSize = n
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// known returns the reason the case is expected to fail under p, or "" when
// it is expected to pass (no entry, or the format has opted in).
func (cfg *config) known(c Case, p Profile, want Expect) string {
	reason := c.Known[p]
	if reason == "" || cfg.strict || cfg.unknown[c.Name] || cfg.unknownFor[p][c.Name] {
		return ""
	}
	if want.Outcome == Limit && len(want.Limits) > 0 {
		for _, l := range want.Limits {
			if cfg.adopted[l] {
				return ""
			}
		}
	}
	return reason
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

// failKind classifies a failure so a Known entry only excuses the kinds of
// failure it is about.
type failKind string

const (
	kVerdict  failKind = "verdict"      // wrong outcome or error type
	kResource failKind = "resource"     // heap over its cap
	kCrash    failKind = "crash"        // the process died
	kPanic    failKind = "panic"        // the parser panicked
	kHang     failKind = "hang"         // no result in time
	kCheck    failKind = "value"        // a successful parse has the wrong value
	kDiff     failKind = "differential" // differs from the reference
)

type failure struct {
	kind failKind
	msg  string
}

func (f *failure) Error() string { return f.msg }

func fail(k failKind, format string, args ...any) error {
	return &failure{kind: k, msg: fmt.Sprintf(format, args...)}
}

// kindOf returns the failKind of err, or "" for other errors.
func kindOf(err error) failKind {
	var f *failure
	if errors.As(err, &f) {
		return f.kind
	}
	return ""
}

// excusable reports whether a Known entry may absorb err: the format gave the
// wrong verdict, ran out of resources, or crashed. A panic or a wrong value is
// a different bug and still fails. So does a hang, except for an isolated
// case: those are the inputs that take an unguarded parser minutes or
// gigabytes today (an alias bomb), so running out of time is how the known gap
// shows.
func excusable(c Case, err error) bool {
	switch kindOf(err) {
	case kVerdict, kResource, kCrash:
		return true
	case kHang:
		return c.Isolated()
	}
	return false
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
			if k := kindOf(err); k != "" {
				fmt.Println(childFail, k)
			}
			r.Errorf("%v", err)
			return
		}
		r.Logf("%s", desc)
		fmt.Println(childOK) // the parent looks for this marker
		return
	}

	known := cfg.known(c, p, want)
	if known != "" && cfg.skipKnownHeavy && (c.Heavy() || c.Isolated()) {
		r.Skipf("known gap, not run (heavy): %s", known)
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

	switch {
	case known != "" && err != nil && excusable(c, err):
		r.Skipf("known gap: %s: %v", known, err)
	case known != "" && err != nil:
		r.Errorf("known gap (%s) hit a different failure: %v", known, err)
	case known != "":
		r.Errorf("known gap now passes, remove its Known entry: %s", known)
	case err != nil:
		r.Errorf("%v", err)
	case desc != "":
		r.Logf("%s", desc)
	}
}

// Timing of one case. The variables exist so tests can shorten them.
var (
	caseTimeout time.Duration // when > 0, replaces the per-case timeout
	graceTime   = 5 * time.Second
)

func timeout(c Case) time.Duration {
	switch {
	case caseTimeout > 0:
		return caseTimeout
	case c.Heavy():
		return 60 * time.Second
	}
	return 10 * time.Second
}

const (
	// liveSlack is allowed on top of MaxMemory when checking the live heap.
	liveSlack = 4 << 20
	minPeak   = 256 << 20
)

// peakCap is the most the heap may grow, at any moment, during one parse. The
// Guard estimate ignores the old and new slice that exist together while an
// append grows, and garbage the collector has not freed; a correct
// Guard-limited parse of small ints measured a peak near 2.5 times MaxMemory.
// So the cap is 3 times MaxMemory (at least 256 MiB, and 2 GiB when no memory
// limit is set). It catches allocation the Guard never sees, such as the 3 GiB
// of an unguarded parse.
func peakCap(l data.Limits) uint64 {
	if l.MaxMemory > 0 {
		return max(3*uint64(l.MaxMemory), minPeak)
	}
	return 2 << 30
}

// measure serializes cases that read the heap, so two Run calls in parallel
// tests of one package do not count each other's allocations. Other
// allocations in the process still count: do not run unrelated parallel
// tests alongside Run.
var measure sync.Mutex

type parseResult struct {
	v    data.Value
	err  error
	fail error // the parser panicked or did not return
}

// callParse runs parse in its own goroutine so a panic or a hang fails the
// case instead of the test binary. A hung goroutine is left behind.
func callParse(parse ParseFunc, in []byte, req Request, limit time.Duration) parseResult {
	type raw struct {
		v     data.Value
		err   error
		pan   any
		stack []byte
	}
	done := make(chan raw, 1)
	go func() {
		var r raw
		defer func() {
			if p := recover(); p != nil {
				r.pan, r.stack = p, debug.Stack()
			}
			done <- r
		}()
		r.v, r.err = parse(in, req)
	}()

	timer := time.NewTimer(limit)
	defer timer.Stop()
	select {
	case r := <-done:
		if r.pan != nil {
			return parseResult{fail: fail(kPanic, "parser panicked: %v\n%s", r.pan, tail(string(r.stack)))}
		}
		return parseResult{v: r.v, err: r.err}
	case <-timer.C:
		return parseResult{fail: fail(kHang, "parser did not return within %s; it must stop when Limits.Context is done", limit)}
	}
}

// runInProcess parses the case input in this process and checks the verdict,
// the value, the heap and the differential. desc describes the verdict.
func runInProcess(c Case, p Profile, want Expect, cfg *config, parse ParseFunc) (desc string, err error) {
	in := c.Input()
	req := RequestFor(p)
	ctx, cancel := context.WithTimeout(context.Background(), timeout(c))
	defer cancel()
	req.Limits.Context = ctx

	measure.Lock()
	defer measure.Unlock()

	// Measure growth from a collected baseline, so garbage left by earlier
	// work is not counted against this parse.
	runtime.GC()
	base := heapBytes()
	hard := peakCap(req.Limits)
	var onExceed func()
	if cfg.childMode {
		// Stop a runaway parse before the kernel does.
		onExceed = func() {
			fmt.Println(childMemCap)
			os.Exit(3)
		}
	}
	stop := startPeakSampler(base, hard, onExceed)
	res := callParse(parse, in, req, timeout(c)+graceTime)
	peak := stop()
	if res.fail != nil {
		return "", res.fail
	}
	if peak > hard {
		return "", fail(kResource, "heap grew by %d MiB during the parse, over the %d MiB cap for %s", peak>>20, hard>>20, p)
	}

	desc, err = verdict(c, want, res.v, res.err)
	if err != nil {
		return "", err
	}

	// What the parse kept must fit the memory budget. This catches a parser
	// whose result is bigger than its Guard accounting says.
	if res.err == nil && req.Limits.MaxMemory > 0 {
		runtime.GC()
		live := heapBytes()
		runtime.KeepAlive(res.v)
		if grown := int64(live) - int64(base); grown > req.Limits.MaxMemory+liveSlack {
			return "", fail(kResource, "result holds %d MiB of live heap, over MaxMemory of %d MiB", grown>>20, req.Limits.MaxMemory>>20)
		}
	}

	if cfg.ref != nil && c.Diff && p == Default && want.Outcome&(Valid|Syntax) != 0 && c.Size <= heavySize {
		if derr := differential(cfg, in, res.v, res.err); derr != nil {
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
				return "", fail(kCheck, "value: %v", cerr)
			}
		}
	case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
		return "", fail(kHang, "parse did not finish in time: %v", err)
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
		return "", fail(kVerdict, "want %s, got %s: %v", describe(want), got, clip(fmt.Sprint(err)))
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
			return fail(kDiff, "value differs from reference: got %s, reference %s", clip(v.String()), clip(rv.String()))
		}
	case rerr == nil:
		return fail(kDiff, "parser rejects input the reference accepts: %v", clip(perr.Error()))
	case perr == nil:
		return fail(kDiff, "parser accepts input the reference rejects: %v", clip(rerr.Error()))
	}
	return nil
}

// heapBytes is the heap occupied by objects, live or not yet swept. Right
// after runtime.GC it is close to the live heap; at other times it includes
// garbage, so it measures a peak, not what a parse keeps.
func heapBytes() uint64 {
	s := []metrics.Sample{{Name: "/memory/classes/heap/objects:bytes"}}
	metrics.Read(s)
	if s[0].Value.Kind() != metrics.KindUint64 {
		return 0
	}
	return s[0].Value.Uint64()
}

// startPeakSampler samples heapBytes every few milliseconds and returns a func
// that stops sampling and reports the peak growth over base. When growth
// passes hard and onExceed is set, onExceed is called (once) from the
// sampler.
func startPeakSampler(base, hard uint64, onExceed func()) func() uint64 {
	var peak atomic.Uint64
	observe := func() {
		if n := heapBytes(); n > base && n-base > peak.Load() {
			peak.Store(n - base)
		}
	}
	stop, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		tick := time.NewTicker(2 * time.Millisecond)
		defer tick.Stop()
		for {
			observe()
			if onExceed != nil && peak.Load() > hard {
				onExceed()
				onExceed = nil
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
		observe()
		return peak.Load()
	}
}

var childFailRE = regexp.MustCompile(childFail + ` (\w+)`)

// runIsolated runs one case in a re-executed copy of the test binary, so a
// fatal crash fails this case alone. The child selects the case by its
// subtest path and runs it in-process, printing childOK when it passed or
// childFail and the failure kind when it did not.
func runIsolated(path []string, c Case) error {
	parts := make([]string, len(path))
	for i, s := range path {
		parts[i] = "^" + regexp.QuoteMeta(s) + "$"
	}
	limit := timeout(c) + graceTime + 10*time.Second
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()

	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	//nolint:gosec // re-executes this test binary with fixed flags
	cmd := exec.CommandContext(ctx, exe,
		"-test.run="+strings.Join(parts, "/"),
		"-test.count=1",
		"-test.v",
		"-test.timeout="+limit.String(),
	)
	cmd.Env = append(os.Environ(), envChild+"=1")
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	runErr := cmd.Run()
	text := out.String()

	switch {
	case runErr == nil && strings.Contains(text, childOK):
		return nil
	case runErr == nil:
		return fmt.Errorf("child process did not run the case (no %s marker; is this a test binary?): %s", childOK, tail(text))
	case ctx.Err() != nil:
		return fail(kHang, "child process timed out after %s", limit)
	case strings.Contains(text, childMemCap):
		return fail(kResource, "heap passed the cap, child process stopped: %s", tail(text))
	}
	if m := childFailRE.FindStringSubmatch(text); m != nil {
		return fail(failKind(m[1]), "child process: %s", tail(text))
	}
	switch {
	case strings.Contains(text, "stack overflow") || strings.Contains(text, "goroutine stack exceeds"):
		return fail(kCrash, "process crash: fatal stack overflow")
	case strings.Contains(text, "out of memory") || strings.Contains(runErr.Error(), "killed"):
		return fail(kCrash, "process killed, probably out of memory (%v)", runErr)
	}
	return fail(kCrash, "child process failed (%v): %s", runErr, tail(text))
}

func tail(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 1500 {
		s = "..." + s[len(s)-1500:]
	}
	return s
}

// clip shortens s for messages, cutting on a rune boundary.
func clip(s string) string {
	const n = 120
	if len(s) <= n {
		return s
	}
	cut := n
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "..."
}
