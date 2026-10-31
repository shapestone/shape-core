package hostiletest_test

import (
	"errors"
	"runtime"
	"strconv"
	"testing"

	"github.com/shapestone/shape-core/pkg/data"
	"github.com/shapestone/shape-core/pkg/hostiletest"
)

// measurePeak parses in under MaxMemory max and returns the peak heap growth
// over a collected baseline, as the harness measures it, and the guard.
func measurePeak(t *testing.T, in []byte, max int64, noGrow bool) (uint64, *data.Guard, error) {
	t.Helper()
	req := hostiletest.RequestFor(hostiletest.Default)
	req.Limits.MaxMemory = max

	runtime.GC()
	stop := hostiletest.StartPeakSampler(hostiletest.HeapBytes(), 0, nil)
	v, g, err := parseJSONWith(in, req, noGrow)
	peak := stop()
	runtime.KeepAlive(v)
	return peak, g, err
}

func intList(n int) []byte {
	in := make([]byte, 0, 2*n+2)
	in = append(in, '[')
	for i := 0; i < n; i++ {
		in = append(in, '0', ',')
	}
	return append(in, '0', ']')
}

func bigMap(n int) []byte {
	in := make([]byte, 0, 12*n)
	in = append(in, '{')
	for i := 0; i < n; i++ {
		in = append(in, '"', 'k')
		in = strconv.AppendInt(in, int64(i), 10)
		in = append(in, '"', ':', '0', ',')
	}
	return append(in, `"z":0}`...)
}

// TestGrowKeepsPeakNearBudget parses a list of small integers far over
// MaxMemory with and without Guard.Grow. Both stop at the limit; Grow stops
// before the append that would hold two arrays at once.
func TestGrowKeepsPeakNearBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("allocates close to 1 GiB")
	}
	const max = 256 << 20
	in := intList(8_000_000) // more than max / 81 bytes per integer

	for _, noGrow := range []bool{true, false} {
		peak, g, err := measurePeak(t, in, max, noGrow)
		var le *data.LimitError
		if !errors.As(err, &le) || le.Limit != data.LimitMemory {
			t.Fatalf("noGrow=%v: err = %v, want a memory LimitError", noGrow, err)
		}
		ratio := float64(peak) / max
		t.Logf("noGrow=%v: estimate %d MiB, peak estimate %d MiB, peak heap %d MiB = %.2fx MaxMemory",
			noGrow, g.MemoryUsed()>>20, g.PeakMemory()>>20, peak>>20, ratio)
		if !noGrow && ratio > 1.25 {
			t.Errorf("peak heap with Grow is %.2fx MaxMemory, want at most 1.25x", ratio)
		}
		if noGrow && ratio < 1.5 {
			t.Errorf("peak heap without Grow is only %.2fx MaxMemory; the test no longer shows what Grow saves", ratio)
		}
		if !noGrow && g.PeakMemory() > max {
			t.Errorf("PeakMemory %d is over MaxMemory %d", g.PeakMemory(), int64(max))
		}
	}
}

// TestGrowMapPeak is the same measurement for one large object.
func TestGrowMapPeak(t *testing.T) {
	if testing.Short() {
		t.Skip("allocates close to 1 GiB")
	}
	const max = 256 << 20
	in := bigMap(2_500_000)

	for _, noGrow := range []bool{true, false} {
		peak, g, err := measurePeak(t, in, max, noGrow)
		var le *data.LimitError
		if !errors.As(err, &le) || le.Limit != data.LimitMemory {
			t.Fatalf("noGrow=%v: err = %v, want a memory LimitError", noGrow, err)
		}
		ratio := float64(peak) / max
		t.Logf("noGrow=%v: estimate %d MiB, peak heap %d MiB = %.2fx MaxMemory", noGrow, g.MemoryUsed()>>20, peak>>20, ratio)
		// The Go index map of a large object is not charged, so the bound
		// is looser than for a list.
		if !noGrow && ratio > 1.4 {
			t.Errorf("peak heap with Grow is %.2fx MaxMemory, want at most 1.4x", ratio)
		}
		if noGrow && ratio < 1.5 {
			t.Errorf("peak heap without Grow is only %.2fx MaxMemory; the test no longer shows what Grow saves", ratio)
		}
	}
}
