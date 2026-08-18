package coverage

import (
	"math"
	"testing"

	"interval-tree/internal/interval"
)

// iv is a short constructor for building interval.Interval values in tests.
func iv(s, e int64) interval.Interval {
	return interval.Interval{Start: s, End: e}
}

func mustNew(t *testing.T, ivs ...interval.Interval) *Analyzer {
	t.Helper()
	a, err := New(ivs)
	if err != nil {
		t.Fatalf("New(%v) unexpected error: %v", ivs, err)
	}
	return a
}

func TestCoverageNewValidation(t *testing.T) {
	if _, err := New([]interval.Interval{iv(1, 5), iv(10, 1)}); err == nil {
		t.Error("New with invalid interval should error")
	}
	// Valid input should not mutate the caller's slice.
	in := []interval.Interval{iv(1, 5)}
	a := mustNew(t, in...)
	a.intervals[0] = iv(99, 99)
	if in[0] != (interval.Interval{Start: 1, End: 5}) {
		t.Error("New copied (not referenced) the input slice")
	}
}

func TestCoverageUnion(t *testing.T) {
	cases := []struct {
		name string
		ivs  []interval.Interval
		want int64
	}{
		{"empty", nil, 0},
		{"single", []interval.Interval{iv(1, 5)}, 5},
		{"disjoint", []interval.Interval{iv(1, 5), iv(10, 20)}, 5 + 11},
		{"overlap", []interval.Interval{iv(1, 5), iv(3, 8)}, 8},
		{"touching", []interval.Interval{iv(1, 5), iv(6, 10)}, 10},
		{"contained", []interval.Interval{iv(1, 10), iv(3, 4)}, 10},
		{"negative", []interval.Interval{iv(-5, -1), iv(0, 3)}, 5 + 4},
		{"big", []interval.Interval{iv(0, 100)}, 101},
	}
	for _, c := range cases {
		a := mustNew(t, c.ivs...)
		if got := a.UnionLength(); got != c.want {
			t.Errorf("%s: UnionLength = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestCoverageCovered(t *testing.T) {
	a := mustNew(t, iv(0, 10), iv(20, 30))
	if got := a.Covered(0, 30); got != 11+11 {
		t.Errorf("Covered(0,30) = %d, want %d", got, 22)
	}
	// Window smaller than coverage.
	if got := a.Covered(5, 25); got != (10-5+1) + (25-20+1) {
		t.Errorf("Covered(5,25) = %d, want %d", got, 6+6)
	}
	// Window completely in a gap.
	if got := a.Covered(11, 19); got != 0 {
		t.Errorf("Covered(11,19) = %d, want 0", got)
	}
	// Inverted window.
	if got := a.Covered(30, 0); got != 0 {
		t.Errorf("Covered(30,0) inverted = %d, want 0", got)
	}
}

func TestCoverageGaps(t *testing.T) {
	cases := []struct {
		name string
		ivs  []interval.Interval
		min  int64
		max  int64
		want []interval.Interval
	}{
		{
			"fully covered",
			[]interval.Interval{iv(1, 5), iv(6, 10)},
			1, 10,
			nil,
		},
		{
			"one gap in middle",
			[]interval.Interval{iv(1, 5), iv(7, 10)},
			1, 10,
			[]interval.Interval{iv(6, 6)},
		},
		{
			"gap at edges",
			[]interval.Interval{iv(5, 8)},
			1, 10,
			[]interval.Interval{iv(1, 4), iv(9, 10)},
		},
		{
			"no intervals",
			nil,
			1, 10,
			[]interval.Interval{iv(1, 10)},
		},
		{
			"touching no gap",
			[]interval.Interval{iv(1, 4), iv(5, 8), iv(9, 12)},
			1, 12,
			nil,
		},
		{
			"overlapping merged",
			[]interval.Interval{iv(1, 5), iv(3, 9), iv(8, 12)},
			0, 20,
			[]interval.Interval{iv(0, 0), iv(13, 20)},
		},
		{
			"inverted window",
			[]interval.Interval{iv(1, 5)},
			10, 1,
			nil,
		},
	}
	for _, c := range cases {
		a := mustNew(t, c.ivs...)
		got := a.Gaps(c.min, c.max)
		if len(got) != len(c.want) {
			t.Errorf("%s: Gaps len = %d, want %d (%v)", c.name, len(got), len(c.want), got)
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("%s: Gaps[%d] = %v, want %v", c.name, i, got[i], c.want[i])
			}
		}
	}
}

func TestCoverageOverlapCount(t *testing.T) {
	a := mustNew(t,
		iv(1, 5),
		iv(3, 7),
		iv(6, 10),
	)
	// Expected per-integer counts over [0,10]:
	//  0:0  1-2:1  3-5:2  6-7:2  8-10:1
	runs := a.OverlapCount(0, 10)
	want := []Run{
		{0, 0, 0},
		{1, 2, 1},
		{3, 5, 2},
		{6, 7, 2},
		{8, 10, 1},
	}
	if len(runs) != len(want) {
		t.Fatalf("OverlapCount runs = %d, want %d (%v)", len(runs), len(want), runs)
	}
	for i := range want {
		if runs[i] != want[i] {
			t.Errorf("OverlapCount[%d] = %v, want %v", i, runs[i], want[i])
		}
	}
}

func TestCoverageOverlapCountNoIntervals(t *testing.T) {
	a := mustNew(t)
	runs := a.OverlapCount(0, 5)
	want := []Run{{0, 5, 0}}
	if len(runs) != len(want) || runs[0] != want[0] {
		t.Errorf("OverlapCount empty = %v, want %v", runs, want)
	}
}

func TestCoverageOverlapCountClamp(t *testing.T) {
	a := mustNew(t, iv(-10, 100))
	// Window sits entirely inside the single interval.
	runs := a.OverlapCount(0, 9)
	if len(runs) != 1 || runs[0] != (Run{0, 9, 1}) {
		t.Errorf("OverlapCount clamp = %v, want [0,9]:1", runs)
	}
	// Inverted window.
	if got := a.OverlapCount(10, 0); got != nil {
		t.Errorf("OverlapCount inverted = %v, want nil", got)
	}
}

func TestCoveragePercent(t *testing.T) {
	cases := []struct {
		name string
		ivs  []interval.Interval
		min  int64
		max  int64
		want float64
	}{
		{"full", []interval.Interval{iv(0, 9)}, 0, 9, 100},
		{"half", []interval.Interval{iv(0, 4)}, 0, 9, 50},
		{"none", nil, 0, 9, 0},
		{"partial", []interval.Interval{iv(0, 2), iv(7, 9)}, 0, 9, 60}, // 6 of 10
		{"inverted", []interval.Interval{iv(0, 9)}, 9, 0, 0},
	}
	for _, c := range cases {
		a := mustNew(t, c.ivs...)
		got := a.CoveragePercent(c.min, c.max)
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s: CoveragePercent = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestCoverageMerged(t *testing.T) {
	a := mustNew(t,
		iv(1, 5),
		iv(3, 8),
		iv(9, 12),
		iv(20, 25),
	)
	got := a.Merged()
	// {1,5} + {3,8} + {9,12} touch/overlap and merge into {1,12}; {20,25} stays.
	want := []interval.Interval{iv(1, 12), iv(20, 25)}
	if len(got) != len(want) {
		t.Fatalf("Merged len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Merged[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestCoverageSummarize(t *testing.T) {
	a := mustNew(t, iv(0, 5), iv(8, 10))
	s := a.Summarize(0, 10)
	if s.Total != 11 {
		t.Errorf("Total = %d, want 11", s.Total)
	}
	if s.Covered != (6 + 3) {
		t.Errorf("Covered = %d, want 9", s.Covered)
	}
	if s.GapCount != 1 {
		t.Errorf("GapCount = %d, want 1", s.GapCount)
	}
	wantPct := (9.0 / 11.0) * 100.0
	if math.Abs(s.CoveragePercent-wantPct) > 1e-9 {
		t.Errorf("CoveragePercent = %v, want %v", s.CoveragePercent, wantPct)
	}
}
