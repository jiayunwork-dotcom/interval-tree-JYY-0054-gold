// Package merge implements interval merging algorithms.
package merge

import (
	"sort"

	"interval-tree/internal/interval"
)

// Merge takes a slice of intervals and returns the minimal non-overlapping set.
func Merge(intervals []interval.Interval) []interval.Interval {
	if len(intervals) == 0 {
		return nil
	}
	sorted := make([]interval.Interval, len(intervals))
	copy(sorted, intervals)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Start != sorted[j].Start {
			return sorted[i].Start < sorted[j].Start
		}
		return sorted[i].End < sorted[j].End
	})

	merged := []interval.Interval{sorted[0]}
	for _, iv := range sorted[1:] {
		last := &merged[len(merged)-1]
		if iv.Start <= last.End+1 {
			if iv.End > last.End {
				last.End = iv.End
			}
		} else {
			merged = append(merged, iv)
		}
	}
	return merged
}

// Subtract removes sub intervals from base intervals.
func Subtract(base, sub []interval.Interval) []interval.Interval {
	baseMerged := Merge(base)
	subMerged := Merge(sub)

	var result []interval.Interval
	si := 0
	for _, b := range baseMerged {
		cur := b.Start
		for si < len(subMerged) && subMerged[si].Start <= b.End {
			s := subMerged[si]
			if s.End < cur {
				si++
				continue
			}
			if s.Start > cur {
				result = append(result, interval.Interval{Start: cur, End: s.Start - 1})
			}
			if s.End >= b.End {
				cur = b.End + 1
				break
			}
			cur = s.End + 1
			si++
		}
		if cur <= b.End {
			result = append(result, interval.Interval{Start: cur, End: b.End})
		}
	}
	return result
}

// Intersect returns points covered by both sets.
func Intersect(a, b []interval.Interval) []interval.Interval {
	aMerged := Merge(a)
	bMerged := Merge(b)

	var result []interval.Interval
	ai, bi := 0, 0
	for ai < len(aMerged) && bi < len(bMerged) {
		lo := maxI64(aMerged[ai].Start, bMerged[bi].Start)
		hi := minI64(aMerged[ai].End, bMerged[bi].End)
		if lo <= hi {
			result = append(result, interval.Interval{Start: lo, End: hi})
		}
		if aMerged[ai].End < bMerged[bi].End {
			ai++
		} else {
			bi++
		}
	}
	return result
}

// Union returns the union of two interval sets.
func Union(a, b []interval.Interval) []interval.Interval {
	all := make([]interval.Interval, 0, len(a)+len(b))
	all = append(all, a...)
	all = append(all, b...)
	return Merge(all)
}

// SymmetricDifference returns points covered by exactly one set.
func SymmetricDifference(a, b []interval.Interval) []interval.Interval {
	union := Union(a, b)
	inter := Intersect(a, b)
	return Subtract(union, inter)
}

// CoveredLength returns the total number of integer points covered.
func CoveredLength(intervals []interval.Interval) int64 {
	merged := Merge(intervals)
	var total int64
	for _, iv := range merged {
		total += iv.End - iv.Start + 1
	}
	return total
}

func maxI64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func minI64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
