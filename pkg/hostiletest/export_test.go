package hostiletest

// HeapBytes and StartPeakSampler let the external tests use the harness's
// peak measurement.
var (
	HeapBytes        = heapBytes
	StartPeakSampler = startPeakSampler
)
