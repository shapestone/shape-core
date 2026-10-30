package hostiletest

import (
	"strconv"
	"strings"
	"sync"

	"github.com/shapestone/shape-core/pkg/data"
)

// Format names the format a case applies to. Each case belongs to exactly one
// format, so Known gaps are per format.
type Format string

// The four formats of the Shape ecosystem.
const (
	JSON Format = "json"
	YAML Format = "yaml"
	CSV  Format = "csv"
	XML  Format = "xml"
)

// Formats returns every format in the corpus.
func Formats() []Format { return []Format{JSON, YAML, CSV, XML} }

// Profile selects the data.Limits a case runs under.
type Profile string

const (
	// Default is data.Limits{}: depth 10,000 and alias expansion on, the
	// rest unlimited.
	Default Profile = "default"

	// Wire mirrors the WireScript client: 64 MiB of input, 1 GiB of
	// estimated memory, depth data.DefaultMaxDepth, and a context. Wire
	// also expects duplicate keys to be rejected.
	Wire Profile = "wire"

	// Tight sets every limit low (see Limits) so each limit name can be
	// provoked with a small input.
	Tight Profile = "tight"
)

// Profiles returns every profile.
func Profiles() []Profile { return []Profile{Default, Wire, Tight} }

// Limits returns the data.Limits for p. The harness adds a context. Unknown
// profiles get the defaults.
func Limits(p Profile) data.Limits {
	switch p {
	case Wire:
		return data.Limits{
			MaxBytes:  64 << 20,
			MaxMemory: 1 << 30,
			MaxDepth:  data.DefaultMaxDepth,
		}
	case Tight:
		return data.Limits{
			MaxDepth:          100,
			MaxBytes:          1 << 20,
			MaxNodes:          10_000,
			MaxStringBytes:    1024,
			MaxKeysPerMap:     100,
			MaxAliasExpansion: 10_000,
			MaxMemory:         1 << 20,
		}
	}
	return data.Limits{}
}

// Outcome is the verdict a parse is expected to reach. Valid, Syntax and
// Limit combine with |; Record stands alone.
type Outcome uint8

const (
	// Valid means the parse succeeds.
	Valid Outcome = 1 << iota
	// Syntax means a *data.SyntaxError (errors.Is(err, data.ErrSyntax)).
	Syntax
	// Limit means a *data.LimitError (errors.Is(err, data.ErrLimit)).
	Limit
	// Record asserts nothing about the verdict: the harness only logs it.
	// It marks inputs whose handling is a per-format policy decision.
	Record
)

// Expect is the expected verdict for one profile. When Outcome includes Limit
// and Limits is not empty, the LimitError's Limit must be one of the names
// (data.LimitDepth and friends). The zero Expect means "not run".
type Expect struct {
	Outcome Outcome
	Limits  []string
}

// Tags the harness acts on. Cases carry other free-form tags for filtering.
const (
	// TagHeavy skips the case under -short. Inputs over 1 MiB are heavy
	// whether tagged or not.
	TagHeavy = "heavy"

	// TagCrash runs the case in a re-executed test binary because it may
	// end the process (stack overflow, out of memory). Inputs of 16 MiB or
	// more are isolated whether tagged or not.
	TagCrash = "crash"
)

const (
	heavySize    = 1 << 20
	isolatedSize = 16 << 20
)

// Case is one hostile (or merely awkward) input and what each profile must
// do with it.
type Case struct {
	Name   string // unique within Format; no spaces or slashes
	Format Format
	Tags   []string

	// Size is the length of Input() in bytes, known without building it.
	Size int64

	// Input builds the bytes. Callers must not modify them.
	Input func() []byte

	// Want maps a profile to its expectation. A profile that is absent is
	// not run.
	Want map[Profile]Expect

	// Known maps a profile to the reason the case is expected to fail today.
	Known map[Profile]string

	// Check, when set, inspects the value of a successful parse.
	Check func(data.Value) error

	// Diff enables the differential comparison against the reference. Only
	// inputs that are small and safe for the reference set it.
	Diff bool
}

// HasTag reports whether the case carries tag.
func (c Case) HasTag(tag string) bool {
	for _, t := range c.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// Heavy reports whether -short skips the case.
func (c Case) Heavy() bool { return c.HasTag(TagHeavy) || c.Size > heavySize }

// Isolated reports whether the case runs in a separate process.
func (c Case) Isolated() bool { return c.HasTag(TagCrash) || c.Size >= isolatedSize }

// gap returns a copy that is expected to fail today under ps (every profile
// in Want when ps is empty).
func (c Case) gap(reason string, ps ...Profile) Case {
	known := make(map[Profile]string, len(c.Known)+len(c.Want))
	for p, r := range c.Known {
		known[p] = r
	}
	if len(ps) == 0 {
		for p := range c.Want {
			ps = append(ps, p)
		}
	}
	for _, p := range ps {
		known[p] = reason
	}
	c.Known = known
	return c
}

var (
	allCases []Case
	allOnce  sync.Once
)

// All returns the whole corpus. The slice is shared; do not modify it.
func All() []Case {
	allOnce.Do(func() { allCases = build() })
	return allCases
}

// For returns the cases of one format.
func For(f Format) []Case {
	var out []Case
	for _, c := range All() {
		if c.Format == f {
			out = append(out, c)
		}
	}
	return out
}

// Seeds returns the small inputs of a format (4 KiB or less), for seeding a
// fuzz corpus with f.Add.
func Seeds(f Format) [][]byte {
	var out [][]byte
	for _, c := range For(f) {
		if c.Size <= 4096 {
			out = append(out, c.Input())
		}
	}
	return out
}

// Reasons recorded with Known entries.
const (
	gapUnadopted = "format has not adopted pkg/data Limits and Guard yet (Phases 2-4)"
	gapStack     = "fatal stack overflow today (measured on Go 1.26, arm64)"
	gapAlias     = "alias bomb expands past 3 GiB today"
	gapJSONHeap  = "64 MiB of small ints uses about 1.8 GiB of live heap today"
	gapCSVHeap   = "64 MiB of CSV uses more than 3 GiB today"
	gapDupKeys   = "duplicate-key rejection not confirmed for this format; its task confirms and removes this"
)

// --- generators ---------------------------------------------------------

// gen builds an input and knows its size without building it.
type gen struct {
	size  int64
	build func() []byte
}

func lit(s string) gen {
	return gen{int64(len(s)), func() []byte { return []byte(s) }}
}

// fill appends unit n times by doubling, so it makes O(log n) copies.
func fill(b []byte, unit string, n int) []byte {
	if n <= 0 || unit == "" {
		return b
	}
	start := len(b)
	end := start + len(unit)*n
	b = append(b, unit...)
	for len(b) < end {
		b = append(b, b[start:start+min(len(b)-start, end-len(b))]...)
	}
	return b
}

// rep is prefix + unit*n + suffix.
func rep(prefix, unit string, n int, suffix string) gen {
	size := len(prefix) + len(unit)*n + len(suffix)
	return gen{int64(size), func() []byte {
		b := make([]byte, 0, size)
		b = append(b, prefix...)
		b = fill(b, unit, n)
		return append(b, suffix...)
	}}
}

// nest is open*n + leaf + closer*n.
func nest(open, leaf, closer string, n int) gen {
	size := n*(len(open)+len(closer)) + len(leaf)
	return gen{int64(size), func() []byte {
		b := make([]byte, 0, size)
		b = fill(b, open, n)
		b = append(b, leaf...)
		return fill(b, closer, n)
	}}
}

// sumDigits is the total number of decimal digits in 0..n-1.
func sumDigits(n int) int64 {
	var total int64
	width, limit := int64(1), 10
	for start := 0; start < n; start, limit, width = limit, limit*10, width+1 {
		total += int64(min(limit, n)-start) * width
	}
	return total
}

// seq is prefix + join(sep, pre+i+post for i in 0..n-1) + suffix.
func seq(prefix string, n int, sep, pre, post, suffix string) gen {
	size := int64(len(prefix)+len(suffix)) + int64(n)*int64(len(pre)+len(post)) + sumDigits(n)
	if n > 1 {
		size += int64(n-1) * int64(len(sep))
	}
	return gen{size, func() []byte {
		b := make([]byte, 0, size)
		b = append(b, prefix...)
		for i := 0; i < n; i++ {
			if i > 0 {
				b = append(b, sep...)
			}
			b = append(b, pre...)
			b = strconv.AppendInt(b, int64(i), 10)
			b = append(b, post...)
		}
		return append(b, suffix...)
	}}
}

// --- expectations -------------------------------------------------------

var (
	valid  = Expect{Outcome: Valid}
	syntax = Expect{Outcome: Syntax}
	record = Expect{Outcome: Record}
)

func limit(names ...string) Expect { return Expect{Outcome: Limit, Limits: names} }

// by builds a Want map; a zero Expect leaves that profile out.
func by(d, w, t Expect) map[Profile]Expect {
	m := make(map[Profile]Expect, 3)
	for p, e := range map[Profile]Expect{Default: d, Wire: w, Tight: t} {
		if e.Outcome != 0 {
			m[p] = e
		}
	}
	return m
}

func every(e Expect) map[Profile]Expect { return by(e, e, e) }

func mk(f Format, name string, g gen, want map[Profile]Expect, tags ...string) Case {
	return Case{Name: name, Format: f, Tags: tags, Size: g.size, Input: g.build, Want: want}
}

func (c Case) diff() Case { c.Diff = true; return c }

func (c Case) check(fn func(data.Value) error) Case { c.Check = fn; return c }

// --- the corpus ---------------------------------------------------------

func build() []Case {
	cs := make([]Case, 0, 256)
	cs = append(cs, conformance()...)
	cs = append(cs, nesting()...)
	cs = append(cs, numbers()...)
	cs = append(cs, aliases()...)
	cs = append(cs, sizes()...)
	cs = append(cs, encoding()...)
	cs = append(cs, duplicateKeys()...)
	cs = append(cs, xmlCases()...)

	// With no format on pkg/data yet, no case that demands a LimitError
	// can pass today. Mark them rather than repeating it at each site.
	for i, c := range cs {
		for p, w := range c.Want {
			if w.Outcome == Limit && c.Known[p] == "" {
				cs[i] = c.gap(gapUnadopted, p)
				c = cs[i]
			}
		}
	}
	return cs
}

type snippet struct{ name, in string }

func snippets(f Format, ok bool, ss []snippet) []Case {
	exp := syntax
	if ok {
		exp = valid
	}
	out := make([]Case, 0, len(ss))
	prefix := "syntax-"
	if ok {
		prefix = "valid-"
	}
	for _, s := range ss {
		c := mk(f, prefix+s.name, lit(s.in), every(exp), "conformance")
		c.Diff = f == JSON
		out = append(out, c)
	}
	return out
}

// conformance holds small inputs every parser must get right. JSON cases are
// also compared with the reference.
func conformance() []Case {
	cs := make([]Case, 0, 128)
	cs = append(cs, snippets(JSON, true, []snippet{
		{"null", `null`}, {"true", `true`}, {"false", `false`},
		{"zero", `0`}, {"negative-zero", `-0`}, {"negative-fraction", `-1.5`},
		{"exponent", `1e5`}, {"exponent-upper-plus", `1E+5`}, {"fraction-exponent", `1.5e-3`},
		{"string", `"str"`}, {"empty-string", `""`},
		{"escapes", `"a\"b\\c\/d\b\f\n\r\t"`}, {"unicode-escape", `"é"`},
		{"surrogate-pair", `"😀"`},
		{"empty-list", `[]`}, {"empty-map", `{}`}, {"list", `[1,2,3]`},
		{"map", `{"a":1,"b":[true,false,null]}`}, {"empty-key", `{"":0}`},
		{"whitespace", " \t\r\n[ 1 , 2 ] \n"}, {"nested-empties", `[[],[[]],{}]`},
	})...)
	cs = append(cs, snippets(JSON, false, []snippet{
		{"empty", ``}, {"blank", ` `}, {"open-list", `[`}, {"close-list", `]`},
		{"trailing-comma-list", `[1,]`}, {"leading-comma-list", `[,1]`},
		{"trailing-comma-map", `{"a":1,}`}, {"map-no-colon", `{"a" 1}`}, {"map-no-value", `{"a"}`},
		{"unquoted-key", `{a:1}`}, {"leading-zero", `01`}, {"bare-dot", `1.`},
		{"leading-dot", `.5`}, {"plus-sign", `+1`}, {"bare-minus", `-`}, {"bare-exponent", `1e`},
		{"nan", `NaN`}, {"infinity", `Infinity`}, {"negative-infinity", `-Infinity`},
		{"single-quotes", `'a'`}, {"missing-comma", `[1 2]`},
		{"bad-escape", `"\x"`}, {"short-unicode-escape", `"\u12"`},
		{"raw-newline-in-string", "\"a\nb\""}, {"two-values", `1 2`},
		{"extra-close-list", `[1]]`}, {"extra-close-map", `{"a":1}}`},
		{"truncated-null", `nul`}, {"truncated-true", `tru`}, {"nul-byte", "\x00"},
	})...)

	cs = append(cs, snippets(YAML, true, []snippet{
		{"mapping", "a: 1\n"}, {"sequence", "- 1\n- 2\n"}, {"flow-sequence", `[1, 2]`},
		{"flow-mapping", `{a: 1}`}, {"alias-small", "a: &a [1, 2]\nb: *a\n"},
	})...)
	cs = append(cs, snippets(YAML, false, []snippet{
		{"open-flow-sequence", `[1, 2`}, {"open-flow-in-mapping", "a: [\n"},
		{"open-flow-mapping", `{a: 1`}, {"unterminated-string", "a: \"abc\n"},
	})...)

	cs = append(cs, snippets(XML, true, []snippet{
		{"empty-element", `<a/>`}, {"text", `<a>text</a>`},
		{"declaration", `<?xml version="1.0"?><a/>`}, {"attribute", `<a x="1"/>`},
	})...)
	cs = append(cs, snippets(XML, false, []snippet{
		{"unclosed", `<a>`}, {"mismatched", `<a></b>`}, {"bare-open", `<`},
		{"unquoted-attribute", `<a x=1/>`},
	})...)

	cs = append(cs, snippets(CSV, true, []snippet{
		{"two-rows", "a,b\n1,2\n"}, {"quoted-comma", "\"a,b\",c\n"},
		{"quoted-newline", "\"a\nb\",c\n"}, {"escaped-quote", "\"a\"\"b\",c\n"},
	})...)
	cs = append(cs, snippets(CSV, false, []snippet{
		{"unterminated-quote", "\"abc"},
	})...)
	return cs
}

// nestAt is a document nested depth levels deep (innermost level empty), and
// nestOpen is depth unclosed openings.
func nestAt(f Format, depth int) gen {
	if f == XML {
		return nest("<a>", "<a/>", "</a>", depth-1)
	}
	return nest("[", "", "]", depth)
}

func nestOpen(f Format, depth int) gen {
	if f == XML {
		return rep("", "<a>", depth, "")
	}
	return rep("", "[", depth, "")
}

func nesting() []Case {
	cs := make([]Case, 0, 16)
	for _, f := range []Format{JSON, YAML, XML} {
		// The tight profile caps depth at 100, the others at 10,000.
		cs = append(cs,
			mk(f, "nest-100", nestAt(f, 100), every(valid), "nesting"),
			mk(f, "nest-101", nestAt(f, 101), by(valid, valid, limit(data.LimitDepth)), "nesting"),
			mk(f, "nest-10000", nestAt(f, 10000), by(valid, valid, limit(data.LimitDepth)), "nesting"),
			mk(f, "nest-10001", nestAt(f, 10001), every(limit(data.LimitDepth)), "nesting"),
		)
	}
	cs[0] = cs[0].diff() // json nest-100
	cs[2] = cs[2].diff() // json nest-10000

	// Depths that overflow the Go stack today.
	deep := map[Format]int{JSON: 5_000_000, YAML: 4_000_000, XML: 1_000_000}
	for _, f := range []Format{JSON, YAML, XML} {
		// Under Tight the input also exceeds the byte budget, which a
		// parser may check first.
		cs = append(cs, mk(f, "nest-unclosed-huge", nestOpen(f, deep[f]),
			by(limit(data.LimitDepth), limit(data.LimitDepth), limit(data.LimitDepth, data.LimitBytes)),
			"nesting", TagHeavy, TagCrash).gap(gapStack))
	}
	return cs
}

// onlyNumber checks that a parse produced a one-element list holding the
// number text unchanged, of the given kind.
func onlyNumber(text string, kind data.Kind) func(data.Value) error {
	return func(v data.Value) error {
		if v.Kind() != data.KindList || v.List().Len() != 1 {
			return errf("want a one-element list, got %s", clip(v.String()))
		}
		e := v.List()[0]
		if e.Kind() != kind || e.Number().Text != text {
			return errf("want %s %s, got %s", kind, clip(text), clip(e.String()))
		}
		return nil
	}
}

func numbers() []Case {
	digits100 := strings.Repeat("1234567890", 10)
	digits1M := strings.Repeat("9", 1_000_000)
	cs := make([]Case, 0, 12)
	for _, f := range []Format{JSON, YAML} {
		cs = append(cs,
			mk(f, "int-100-digits", lit("["+digits100+"]"), every(valid), "number").
				check(onlyNumber(digits100, data.KindInteger)),
			mk(f, "int-negative-100-digits", lit("[-"+digits100+"]"), every(valid), "number").
				check(onlyNumber("-"+digits100, data.KindInteger)),
			mk(f, "exponent-huge", lit("[1e999999999]"), every(valid), "number"),
			mk(f, "exponent-huge-negative", lit("[1e-999999999]"), every(valid), "number"),
			mk(f, "int-1m-digits", lit("["+digits1M+"]"), by(valid, valid, limit(data.LimitStringBytes)), "number"),
		)
	}
	// The JSON ones have an exact expected shape and a safe reference.
	for i := range cs {
		if cs[i].Format != JSON {
			continue
		}
		cs[i] = cs[i].diff()
		switch cs[i].Name {
		case "exponent-huge":
			cs[i] = cs[i].check(onlyNumber("1e999999999", data.KindDecimal))
		case "exponent-huge-negative":
			cs[i] = cs[i].check(onlyNumber("1e-999999999", data.KindDecimal))
		}
	}
	return cs
}

// ladder is the classic YAML alias bomb with levels levels of fan-out 9.
func ladder(levels int) string {
	var sb strings.Builder
	names := "abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < levels; i++ {
		sb.WriteByte(names[i])
		sb.WriteString(": &")
		sb.WriteByte(names[i])
		sb.WriteString(" [")
		for j := 0; j < 9; j++ {
			if j > 0 {
				sb.WriteByte(',')
			}
			if i == 0 {
				sb.WriteString(`"lol"`)
			} else {
				sb.WriteByte('*')
				sb.WriteByte(names[i-1])
			}
		}
		sb.WriteString("]\n")
	}
	return sb.String()
}

func aliases() []Case {
	return []Case{
		mk(YAML, "alias-ladder-3", lit(ladder(3)), every(valid), "alias"),
		mk(YAML, "alias-bomb", lit(ladder(9)), every(limit(data.LimitAliasExpansion)),
			"alias", TagHeavy, TagCrash).gap(gapAlias),
	}
}

func hugeString(f Format, n int) gen {
	switch f {
	case XML:
		return rep("<a>", "a", n, "</a>")
	case CSV:
		return rep("", "a", n, "\n")
	}
	return rep(`"`, "a", n, `"`) // JSON string, YAML double-quoted scalar
}

func stringLen(n int) func(data.Value) error {
	return func(v data.Value) error {
		if v.Kind() != data.KindString || len(v.Str()) != n {
			return errf("want a string of %d bytes, got %s", n, clip(v.String()))
		}
		return nil
	}
}

// pad is a tiny valid document followed by filler, n+len(doc) bytes long.
func pad(f Format, n int) gen {
	switch f {
	case YAML:
		return rep("0", "\n", n, "")
	case XML:
		return rep("<a/>", " ", n-3, "")
	case CSV:
		return rep("a", "\n", n, "")
	}
	return rep("0", " ", n, "")
}

func sizes() []Case {
	cs := make([]Case, 0, 32)
	for _, f := range Formats() {
		c4 := mk(f, "string-4k", hugeString(f, 4096), by(valid, valid, limit(data.LimitStringBytes)), "string")
		cc := mk(f, "string-huge", hugeString(f, 8<<20), by(valid, valid, Expect{}), "string", TagHeavy)
		if f == JSON || f == YAML {
			c4 = c4.check(stringLen(4096))
			cc = cc.check(stringLen(8 << 20))
		}
		cs = append(cs, c4, cc,
			// One byte over each profile's byte budget.
			mk(f, "bytes-over-tight", pad(f, 1<<20), by(valid, valid, limit(data.LimitBytes)), "bytes", TagHeavy),
			mk(f, "bytes-over-wire", pad(f, 64<<20), by(Expect{}, limit(data.LimitBytes), Expect{}), "bytes", TagHeavy, TagCrash),
		)
	}

	longStr := `"` + strings.Repeat("x", 40) + `",`
	for _, f := range []Format{JSON, YAML} {
		cs = append(cs,
			mk(f, "nodes-10001", rep("[", "0,", 10_000, "0]"), by(valid, valid, limit(data.LimitNodes)), "nodes"),
			// 9,000 strings of 40 bytes: about 1.08 MB estimated, under the
			// byte, node and string limits of Tight but over its memory.
			mk(f, "memory-9000-strings", rep("[", longStr, 9000, `""]`), by(valid, valid, limit(data.LimitMemory)), "memory"),
		)
	}
	cs = append(cs,
		mk(JSON, "keys-1000", seq("{", 1000, ",", `"k`, `":0`, "}"), by(valid, valid, limit(data.LimitKeysPerMap)), "keys"),
		mk(YAML, "keys-1000", seq("", 1000, "\n", "k", ": 0", "\n"), by(valid, valid, limit(data.LimitKeysPerMap)), "keys"),
		mk(JSON, "keys-300k", seq("{", 300_000, ",", `"k`, `":0`, "}"), by(valid, valid, Expect{}), "keys", TagHeavy),
		mk(YAML, "keys-300k", seq("", 300_000, "\n", "k", ": 0", "\n"), by(valid, valid, Expect{}), "keys", TagHeavy),

		// 64 MiB of small ints measured about 1.8 GiB of live heap.
		mk(JSON, "small-ints-64MiB", rep("[", "0,", 33_554_430, "0]"), by(Expect{}, limit(data.LimitMemory), Expect{}),
			"memory", TagHeavy, TagCrash).gap(gapJSONHeap),
		// 64 MiB of CSV measured more than 3 GiB.
		mk(CSV, "rows-64MiB", rep("", "1,2,3\n", 11_184_810, ""), by(Expect{}, limit(data.LimitMemory), Expect{}),
			"memory", TagHeavy, TagCrash).gap(gapCSVHeap),

		// One row of 15,000 fields: some node, key or memory limit must
		// stop it under Tight, whichever the format counts first.
		mk(CSV, "wide-row-15000", rep("", "a,", 15_000, "a\n"),
			by(valid, valid, limit(data.LimitNodes, data.LimitKeysPerMap, data.LimitMemory)), "wide"),
		mk(CSV, "wide-row-1m", rep("", "a,", 1_000_000, "a\n"), by(valid, valid, Expect{}), "wide", TagHeavy),
		mk(CSV, "wide-header-1000", rep("", "h,", 1000, "h\n1\n"),
			by(valid, valid, limit(data.LimitKeysPerMap, data.LimitNodes)), "wide"),
	)
	return cs
}

func encoding() []Case {
	const bom = "\xEF\xBB\xBF"
	return []Case{
		// RFC 8259 lets a parser reject a BOM and the standard library does.
		mk(JSON, "bom", lit(bom+"[1]"), every(syntax), "bom").diff(),
		// YAML and XML specify a BOM at the start of the stream.
		mk(YAML, "bom", lit(bom+"a: 1\n"), every(valid), "bom"),
		mk(XML, "bom", lit(bom+"<a/>"), every(valid), "bom"),
		mk(CSV, "bom", lit(bom+"a,b\n1,2\n"), every(record), "bom"),

		// JSON and CSV carry bytes through; the standard library
		// substitutes U+FFFD, so what to do is a per-format policy.
		mk(JSON, "invalid-utf8-string", lit("[\"\xff\xfe\"]"), every(record), "utf8"),
		mk(JSON, "invalid-utf8-key", lit("{\"\xff\":1}"), every(record), "utf8"),
		mk(JSON, "lone-surrogate-escape", lit(`["\ud800"]`), every(record), "utf8"),
		mk(CSV, "invalid-utf8-field", lit("a,\xff\xfe\n"), every(record), "utf8"),
		mk(YAML, "invalid-utf8-string", lit("a: \"\xff\xfe\"\n"), every(syntax), "utf8"),
		mk(XML, "invalid-utf8-text", lit("<a>\xff\xfe</a>"), every(syntax), "utf8"),
	}
}

// duplicateKeys: wire expects a rejection; elsewhere the behaviour is only
// recorded. Each format's task confirms its behaviour and drops the Known entry.
func duplicateKeys() []Case {
	in := map[Format]string{
		JSON: `{"a":1,"a":2}`,
		YAML: "a: 1\na: 2\n",
		XML:  `<a x="1" x="2"/>`,
		CSV:  "a,a\n1,2\n",
	}
	cs := make([]Case, 0, 4)
	for _, f := range Formats() {
		cs = append(cs, mk(f, "duplicate-keys", lit(in[f]), by(record, syntax, Expect{}), "dup-keys").
			gap(gapDupKeys, Wire))
	}
	return cs
}

func xmlCases() []Case {
	const laughs = `<?xml version="1.0"?>
<!DOCTYPE lolz [
<!ENTITY lol "lol">
<!ENTITY lol2 "&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;">
<!ENTITY lol3 "&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;">
<!ENTITY lol4 "&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;">
<!ENTITY lol5 "&lol4;&lol4;&lol4;&lol4;&lol4;&lol4;&lol4;&lol4;&lol4;&lol4;">
<!ENTITY lol6 "&lol5;&lol5;&lol5;&lol5;&lol5;&lol5;&lol5;&lol5;&lol5;&lol5;">
<!ENTITY lol7 "&lol6;&lol6;&lol6;&lol6;&lol6;&lol6;&lol6;&lol6;&lol6;&lol6;">
<!ENTITY lol8 "&lol7;&lol7;&lol7;&lol7;&lol7;&lol7;&lol7;&lol7;&lol7;&lol7;">
<!ENTITY lol9 "&lol8;&lol8;&lol8;&lol8;&lol8;&lol8;&lol8;&lol8;&lol8;&lol8;">
]>
<lolz>&lol9;</lolz>`
	const external = `<?xml version="1.0"?>
<!DOCTYPE a [<!ENTITY x SYSTEM "file:///etc/passwd">]>
<a>&x;</a>`

	// Entities are never resolved: a parser either rejects the document or
	// stops the expansion with the alias-expansion limit.
	stop := Expect{Outcome: Syntax | Limit, Limits: []string{data.LimitAliasExpansion}}
	return []Case{
		mk(XML, "entity-expansion", lit(laughs), every(stop), "entity", TagHeavy, TagCrash),
		mk(XML, "entity-external", lit(external), every(syntax), "entity"),
		// Whether a bare DOCTYPE is allowed is a per-format policy.
		mk(XML, "doctype-plain", lit(`<!DOCTYPE a><a/>`), every(record), "entity"),
	}
}
