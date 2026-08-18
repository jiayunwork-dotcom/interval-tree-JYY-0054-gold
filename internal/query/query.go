// Package query provides high-level query operations on interval collections.
package query

import (
	"sort"

	"interval-tree/internal/interval"
)

// Collection is a sorted set of intervals for efficient querying.
type Collection struct {
	intervals []interval.Interval
}

// NewCollection creates a collection from intervals.
func NewCollection(ivs []interval.Interval) *Collection {
	c := &Collection{intervals: make([]interval.Interval, len(ivs))}
	copy(c.intervals, ivs)
	c.sort()
	return c
}

// Len returns the number of intervals.
func (c *Collection) Len() int { return len(c.intervals) }

// All returns a copy of all intervals.
func (c *Collection) All() []interval.Interval {
	out := make([]interval.Interval, len(c.intervals))
	copy(out, c.intervals)
	return out
}

// Add inserts an interval and re-sorts.
func (c *Collection) Add(iv interval.Interval) {
	c.intervals = append(c.intervals, iv)
	c.sort()
}

// StabPoint returns all intervals containing point p.
func (c *Collection) StabPoint(p int64) []interval.Interval {
	var result []interval.Interval
	for _, iv := range c.intervals {
		if p >= iv.Start && p <= iv.End {
			result = append(result, iv)
		}
	}
	return result
}

// Overlapping returns all intervals that overlap with the query.
func (c *Collection) Overlapping(q interval.Interval) []interval.Interval {
	var result []interval.Interval
	for _, iv := range c.intervals {
		if iv.Overlaps(q) {
			result = append(result, iv)
		}
	}
	return result
}

// Containing returns all intervals that fully contain the query.
func (c *Collection) Containing(q interval.Interval) []interval.Interval {
	var result []interval.Interval
	for _, iv := range c.intervals {
		if iv.Start <= q.Start && iv.End >= q.End {
			result = append(result, iv)
		}
	}
	return result
}

// ContainedBy returns all intervals fully contained within the query.
func (c *Collection) ContainedBy(q interval.Interval) []interval.Interval {
	var result []interval.Interval
	for _, iv := range c.intervals {
		if iv.Start >= q.Start && iv.End <= q.End {
			result = append(result, iv)
		}
	}
	return result
}

// Nearest returns the interval whose midpoint is closest to the given point.
func (c *Collection) Nearest(p int64) (interval.Interval, bool) {
	if len(c.intervals) == 0 {
		return interval.Interval{}, false
	}
	best := c.intervals[0]
	bestDist := distToMid(best, p)
	for _, iv := range c.intervals[1:] {
		d := distToMid(iv, p)
		if d < bestDist {
			bestDist = d
			best = iv
		}
	}
	return best, true
}

// Gaps returns all uncovered regions within [lo, hi].
func (c *Collection) Gaps(lo, hi int64) []interval.Interval {
	var gaps []interval.Interval
	cursor := lo
	for _, iv := range c.intervals {
		if iv.Start > cursor && iv.Start <= hi {
			gaps = append(gaps, interval.Interval{Start: cursor, End: iv.Start - 1})
		}
		if iv.End >= cursor {
			cursor = iv.End + 1
		}
	}
	if cursor <= hi {
		gaps = append(gaps, interval.Interval{Start: cursor, End: hi})
	}
	return gaps
}

// Span returns the minimum bounding interval.
func (c *Collection) Span() (interval.Interval, bool) {
	if len(c.intervals) == 0 {
		return interval.Interval{}, false
	}
	lo := c.intervals[0].Start
	hi := c.intervals[0].End
	for _, iv := range c.intervals[1:] {
		if iv.Start < lo {
			lo = iv.Start
		}
		if iv.End > hi {
			hi = iv.End
		}
	}
	return interval.Interval{Start: lo, End: hi}, true
}

// Filter returns intervals matching the predicate.
func (c *Collection) Filter(fn func(interval.Interval) bool) []interval.Interval {
	var result []interval.Interval
	for _, iv := range c.intervals {
		if fn(iv) {
			result = append(result, iv)
		}
	}
	return result
}

// MaxGap returns the largest gap between consecutive intervals.
func (c *Collection) MaxGap() int64 {
	if len(c.intervals) < 2 {
		return 0
	}
	var maxG int64
	for i := 1; i < len(c.intervals); i++ {
		gap := c.intervals[i].Start - c.intervals[i-1].End - 1
		if gap > maxG {
			maxG = gap
		}
	}
	return maxG
}

// Density returns coverage density within the span.
func (c *Collection) Density() float64 {
	span, ok := c.Span()
	if !ok {
		return 0
	}
	spanLen := span.End - span.Start + 1
	if spanLen <= 0 {
		return 0
	}
	var covered int64
	for _, iv := range c.intervals {
		covered += iv.End - iv.Start + 1
	}
	result := float64(covered) / float64(spanLen)
	if result > 1 {
		result = 1
	}
	return result
}

func (c *Collection) sort() {
	sort.Slice(c.intervals, func(i, j int) bool {
		if c.intervals[i].Start != c.intervals[j].Start {
			return c.intervals[i].Start < c.intervals[j].Start
		}
		return c.intervals[i].End < c.intervals[j].End
	})
}

func distToMid(iv interval.Interval, p int64) int64 {
	mid := (iv.Start + iv.End) / 2
	d := p - mid
	if d < 0 {
		d = -d
	}
	return d
}
