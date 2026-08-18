// Package coverage analyzes a collection of intervals as a set: how much of a
// range they cover, where the gaps are, and how many intervals pile up on
// each integer. It is built on top of the interval package but is
// completely independent of the tree package - coverage is a set/sweep-line
// problem, not a point/interval query problem, so a different algorithm is
// appropriate.
//
// All intervals are treated as inclusive integer ranges [Start, End], the
// same convention used everywhere else in this module. Methods that take a
// (min, max) window only consider integers inside that window, and a window
// with max < min is treated as empty (zero result, no error).
package coverage

import (
	"errors"
	"fmt"
	"interval-tree/internal/interval"
	"sort"
)

// ErrInvalidInterval is returned by New when the supplied slice contains an
// interval with End < Start.
var ErrInvalidInterval = errors.New("coverage: invalid interval (End < Start)")

// Run is one maximal contiguous integer range that carries the same overlap
// count. It is produced by OverlapCount.
type Run struct {
	Start int64
	End   int64
	Count int
}

// String renders a Run as "[start,end]:count".
func (r Run) String() string {
	return fmt.Sprintf("[%d,%d]:%d", r.Start, r.End, r.Count)
}

// Analyzer holds a set of intervals and answers set-wide questions about
// them. It is constructed with New, which validates the input.
type Analyzer struct {
	intervals []interval.Interval
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// New returns an Analyzer over a copy of intervals. It rejects any invalid
// interval (End < Start) so that downstream math never has to deal with
// negative-length ranges.
func New(intervals []interval.Interval) (*Analyzer, error) {
	for _, iv := range intervals {
		if !iv.Valid() {
			return nil, fmt.Errorf("%w: %v", ErrInvalidInterval, iv)
		}
	}
	return &Analyzer{intervals: append([]interval.Interval(nil), intervals...)}, nil
}

// Intervals returns a copy of the underlying interval set.
func (a *Analyzer) Intervals() []interval.Interval {
	return append([]interval.Interval(nil), a.intervals...)
}

// mergeTouching sorts the intervals by Start and folds overlapping or
// touching intervals together. Touching means the next interval starts at
// most one past the current end (End+1), since for inclusive integer ranges
// [1, 5] and [6, 10] cover every integer from 1 to 10 with no gap and should
// be reported as a single merged range for union/gap purposes.
func mergeTouching(in []interval.Interval) []interval.Interval {
	if len(in) == 0 {
		return nil
	}
	s := interval.SortStart(in)
	out := make([]interval.Interval, 0, len(s))
	out = append(out, s[0])
	for i := 1; i < len(s); i++ {
		iv := s[i]
		last := &out[len(out)-1]
		if iv.Start <= last.End+1 {
			if iv.End > last.End {
				last.End = iv.End
			}
		} else {
			out = append(out, iv)
		}
	}
	return out
}

// Merged returns the intervals after merging overlapping/touching ones,
// sorted by Start. The receiver is not modified.
func (a *Analyzer) Merged() []interval.Interval {
	return mergeTouching(a.intervals)
}

// UnionLength returns the total number of integers covered by at least one
// interval (the size of the union). Touching intervals count only once for
// each integer, so [1, 5] and [6, 10] contribute 10, not 11.
func (a *Analyzer) UnionLength() int64 {
	var total int64
	for _, iv := range mergeTouching(a.intervals) {
		total += iv.Length()
	}
	return total
}

// Covered returns the number of integers inside [min, max] that are covered
// by at least one interval. If max < min the result is 0.
func (a *Analyzer) Covered(min, max int64) int64 {
	if max < min {
		return 0
	}
	var total int64
	for _, iv := range mergeTouching(a.intervals) {
		s := max64(iv.Start, min)
		e := min64(iv.End, max)
		if s <= e {
			total += e - s + 1
		}
	}
	return total
}

// Gaps returns the uncovered integer ranges inside [min, max], sorted by
// Start. If the whole window is covered the result is an empty (non-nil)
// slice; if no interval touches the window at all, the whole window is
// returned as a single gap.
func (a *Analyzer) Gaps(min, max int64) []interval.Interval {
	if max < min {
		return nil
	}
	merged := mergeTouching(a.intervals)
	gaps := make([]interval.Interval, 0)
	cur := min
	for _, iv := range merged {
		if iv.End < min {
			continue
		}
		if iv.Start > max {
			break
		}
		s := max64(iv.Start, min)
		e := min64(iv.End, max)
		if s > cur {
			gaps = append(gaps, interval.Interval{Start: cur, End: s - 1})
		}
		if e >= cur {
			cur = e + 1
		}
	}
	if cur <= max {
		gaps = append(gaps, interval.Interval{Start: cur, End: max})
	}
	return gaps
}

// OverlapCount returns a compact summary of how many intervals cover each
// integer in [min, max]. Equal consecutive counts are collapsed into a single
// Run, so the result is a list of maximal runs rather than one entry per
// integer. This keeps the output small even for huge ranges. If max < min the
// result is nil.
//
// The count for a run [s, e] is the number of input intervals whose range
// contains every integer from s to e.
func (a *Analyzer) OverlapCount(min, max int64) []Run {
	if max < min {
		return nil
	}
	delta := make(map[int64]int)
	// Seed the sentinel positions so the first and last runs are emitted.
	delta[min] = 0
	delta[max+1] = 0
	for _, iv := range a.intervals {
		s := max64(iv.Start, min)
		e := min64(iv.End, max)
		if s > e {
			continue
		}
		delta[s]++
		delta[e+1]--
	}

	pos := make([]int64, 0, len(delta))
	for p := range delta {
		pos = append(pos, p)
	}
	sort.Slice(pos, func(i, j int) bool { return pos[i] < pos[j] })

	runs := make([]Run, 0)
	cnt := 0
	for k, p := range pos {
		cnt += delta[p]
		if k+1 < len(pos) {
			nextP := pos[k+1]
			if p <= nextP-1 {
				runs = append(runs, Run{Start: p, End: nextP - 1, Count: cnt})
			}
		}
	}
	return runs
}

// CoveragePercent returns the fraction of integers in [min, max] covered by
// at least one interval, expressed as a percentage in [0, 100]. If max < min
// the result is 0.
func (a *Analyzer) CoveragePercent(min, max int64) float64 {
	if max < min {
		return 0
	}
	total := max - min + 1
	covered := a.Covered(min, max)
	var pct float64
	if total > 0 {
		pct = float64(covered) / float64(total) * 100.0
	}
	if pct > 100 {
		pct = 100
	}
	return pct
}

// Summary bundles the most commonly requested window statistics into one
// call so a CLI or report can render them together.
type Summary struct {
	Min            int64
	Max            int64
	Total          int64 // integers in the window (max-min+1)
	Covered        int64 // integers covered by >=1 interval
	UnionLength    int64 // same as Covered (kept for clarity)
	GapCount       int   // number of uncovered runs
	CoveragePercent float64
}

// Summarize computes Summary for the window [min, max].
func (a *Analyzer) Summarize(min, max int64) Summary {
	if max < min {
		return Summary{}
	}
	covered := a.Covered(min, max)
	gaps := a.Gaps(min, max)
	total := max - min + 1
	s := Summary{
		Min:            min,
		Max:            max,
		Total:          total,
		Covered:        covered,
		UnionLength:    covered,
		GapCount:       len(gaps),
		CoveragePercent: a.CoveragePercent(min, max),
	}
	return s
}
