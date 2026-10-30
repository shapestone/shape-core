// Package hostiletest is a shared corpus of hostile inputs and a test harness
// for the Shape format modules (shape-json, shape-yaml, shape-csv, shape-xml).
//
// It is importable (not internal) so every format module can run the same
// cases in its own tests, but it is test support: runtime code must never
// import it. It imports only the standard library and pkg/data, and never
// encoding/json or encoding/xml; a format module supplies any stdlib
// reference parser from its own test files.
//
// # Corpus
//
// A [Case] belongs to one [Format] and builds its input on demand with a
// generator, so a 64 MiB document or a 5-million-deep nest costs nothing
// until it runs. Each case states the expected [Expect] under each [Profile]:
//
//   - [Default] is data.Limits{}.
//   - [Wire] mirrors the WireScript client: 64 MiB input, 1 GiB estimated
//     memory, depth data.DefaultMaxDepth, and a context deadline.
//   - [Tight] sets every limit low so each limit name can be provoked with a
//     small input.
//
// # Running it
//
// A format module calls [Run] with a [ParseFunc]:
//
//	func TestHostile(t *testing.T) {
//		hostiletest.Run(t, hostiletest.JSON, parse,
//			hostiletest.WithReference(stdlibJSON))
//	}
//
// ParseFunc must report malformed input as *data.SyntaxError and limit
// violations as *data.LimitError. A format that has not moved onto pkg/data
// yet needs a small adapter, or its cases fail on error type.
//
// # Crashes, known gaps and CI
//
// A Go stack overflow or out-of-memory kill ends the process and cannot be
// recovered. Cases that may do that (tag [TagCrash], or any input of 16 MiB or
// more) therefore run in a re-executed copy of the test binary, and a crash
// is reported as an ordinary failure of that case.
//
// A case that is expected to fail today carries a Known reason per profile.
// Run reports it as skipped while it fails, and fails the test once it
// passes, so a stale entry is noticed. Strict mode, or [WithoutKnown],
// removes the allowance when a format adopts pkg/data.
//
// Cases above 1 MiB (or tagged [TagHeavy]) are skipped under -short.
//
// # Environment
//
//   - HOSTILETEST_STRICT=1 ignores every Known entry.
//   - HOSTILETEST_MAXSIZE=<bytes> skips cases with larger inputs, for small
//     CI runners.
package hostiletest
