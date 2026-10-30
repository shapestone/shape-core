// Package hostiletest is a shared corpus of hostile inputs and a test harness
// for the Shape format modules (shape-json, shape-yaml, shape-csv, shape-xml).
//
// It is importable (not internal) so every format module can run the same
// cases in its own tests, but it is test support: runtime code must never
// import it. It imports only the standard library and pkg/data, and this
// package never imports encoding/json or encoding/xml; a format module
// supplies any stdlib reference parser from its own test files. (Other
// packages of this module may use those packages; the stdlib-only rule for the
// format modules is enforced in their own repositories.)
//
// # Corpus
//
// A [Case] belongs to one [Format] and builds its input on demand with a
// generator, so a 64 MiB document or a 5-million-deep nest costs nothing
// until it runs. [Seeds] returns the small inputs of a format, for
// f.Add in a fuzz test. Each case states the expected [Expect] under each
// [Profile]:
//
//   - [Default] is data.Limits{}.
//   - [Wire] mirrors the WireScript client: 64 MiB input, 1 GiB estimated
//     memory, depth data.DefaultMaxDepth, and a context deadline. It also
//     expects duplicate keys to be rejected, always as a *data.SyntaxError.
//   - [Tight] sets every limit low so each limit name can be provoked with a
//     small input: depth 100, 1 MiB of input, 10,000 nodes, 1,024 bytes per
//     string, 100 keys per map, 10,000 alias-expansion nodes and 1 MiB of
//     estimated memory.
//
// Profile is an open set and may gain values. [Limits] and [PolicyFor] return
// defaults for a profile they do not know, so a ParseFunc must not switch
// exhaustively on [Request.Profile]. It should use [Request.Limits] and
// [Request.Policy] instead.
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
// ParseFunc receives the input and a [Request]. It must report malformed
// input as *data.SyntaxError and limit violations as *data.LimitError, and it
// must stop when Request.Limits.Context is done: a parse that ignores the
// context is failed after a grace period, though its goroutine cannot be
// stopped. A format that has not moved onto pkg/data yet needs a small
// adapter, or its cases fail on error type. A panic in a ParseFunc fails that
// case, with its stack, instead of ending the test binary.
//
// With [WithReference], cases that set Diff are also run through a reference
// implementation, and the parser must agree with it. [Diverges] documents a
// known difference.
//
// # Known gaps
//
// A case that is expected to fail today carries a Known reason per profile.
// While it fails with the wrong verdict, runs out of resources or crashes,
// Run reports it as skipped. A panic, a hang or a wrong value is a different
// bug and fails. Once the case passes, Run fails it, so a stale entry is
// noticed. A format drops entries as it adopts pkg/data: [Adopted] drops
// those for limits it now enforces, [WithoutKnownFor] drops named cases for
// one profile, [WithoutKnown] for all, and [Strict] drops every one.
//
// # Crashes, memory and CI
//
// A Go stack overflow or out-of-memory kill ends the process and cannot be
// recovered. Cases that may do that (tag [TagCrash], or any input of 16 MiB
// or more) therefore run in a re-executed copy of the test binary, found
// with os.Executable, so they need a real test binary and do not work under
// go run. A crash is reported as an ordinary failure of that case, and the
// child stops itself when its heap passes the cap for the profile.
//
// Each parse is measured against a garbage-collected baseline. Its peak heap
// growth may not pass 3 times Limits.MaxMemory (at least 256 MiB; 2 GiB when
// there is no memory limit), because the Guard estimate does not count the
// old and new slice that exist together while a list grows. When the parse
// succeeds, the live heap after a collection may not pass MaxMemory. The heap
// is the whole process's, so do not run unrelated parallel tests while Run
// is working; Run itself measures one case at a time.
//
// Cases above 1 MiB (or tagged [TagHeavy]) are skipped under -short.
//
// # Environment
//
//   - HOSTILETEST_STRICT=1 ignores every Known entry.
//   - HOSTILETEST_SKIPKNOWN=1 does not run Known cases that are heavy or
//     isolated, for runners where even a contained 3 GiB parse is too much.
//   - HOSTILETEST_MAXSIZE=<bytes> skips cases with larger inputs.
package hostiletest
