//go:build racebug

// This file is intentionally gated behind the `racebug` build tag.
// It contains tests that document EXISTING data races / goroutine leaks
// in torrstor. Under `go test -race` they will (and should) FAIL — that
// failure is the proof of the bug. Once the underlying bug is fixed,
// remove the build tag and the test moves into the regular test set.
//
// Currently documented:
//   (none — Bug #16 has been fixed; see TestCleanPieces_NoRace_Concurrent
//   in integration_test.go)
//
// Fixed and promoted to integration_test.go:
//   * Bugs #13, #14, #15 — see TestReader_CloseUnblocksRead_NoRace and
//     TestReader_ClosedReader_NoSilentEOF.
//   * Bug #16 — Cache.cleanPieces / getRemPieces race; the previous
//     placeholder test (TestCleanPieces_DoesNotEvictPieceInActiveRange)
//     conflated bug #5 (eviction invariant) with bug #16 (race) and
//     relied on race-induced under-counting of c.filled to keep piece 8
//     alive. See cache.go's atomic.Bool isRemove / isClosed and the
//     piece atomic conversions; the race-clean version of the
//     invariant test now lives in integration_test.go.
//
// Run with:
//   go test -race -tags=racebug ./torr/storage/torrstor/

package torrstor_test

