// Package schedule implements interval scheduling algorithms using the interval
// tree. Given a set of intervals (representing tasks/events with start and end
// times), the scheduler can find maximum non-overlapping subsets, detect
// conflicts, and compute utilization metrics.
package schedule

import (
	"sort"

	"interval-tree/internal/interval"
)

// Event is a named interval representing a scheduled task or time slot.
type Event struct {
	Name  string
	Start int64
	End   int64
}

// ToInterval converts an Event to an Interval.
func (e Event) ToInterval() interval.Interval {
	return interval.Interval{Start: e.Start, End: e.End}
}

// Conflict represents two events that overlap in time.
type Conflict struct {
	A Event
	B Event
}

// FindConflicts returns all pairs of events that overlap.
func FindConflicts(events []Event) []Conflict {
	var conflicts []Conflict
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			ivI := events[i].ToInterval()
			ivJ := events[j].ToInterval()
			if ivI.Overlaps(ivJ) {
				conflicts = append(conflicts, Conflict{A: events[i], B: events[j]})
			}
		}
	}
	return conflicts
}

// HasConflict reports whether any two events overlap.
func HasConflict(events []Event) bool {
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].ToInterval().Overlaps(events[j].ToInterval()) {
				return true
			}
		}
	}
	return false
}

// MaxNonOverlapping finds the largest subset of non-overlapping events using
// a greedy earliest-end-first algorithm (optimal for interval scheduling).
func MaxNonOverlapping(events []Event) []Event {
	if len(events) == 0 {
		return nil
	}
	sorted := make([]Event, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].End < sorted[j].End
	})

	var selected []Event
	var lastEnd int64 = -(1 << 62)
	for _, e := range sorted {
		if e.Start > lastEnd {
			selected = append(selected, e)
			lastEnd = e.End
		}
	}
	return selected
}

// Utilization computes the fraction of the total span covered by events.
func Utilization(events []Event) float64 {
	if len(events) == 0 {
		return 0
	}
	minStart := events[0].Start
	maxEnd := events[0].End
	for _, e := range events[1:] {
		if e.Start < minStart {
			minStart = e.Start
		}
		if e.End > maxEnd {
			maxEnd = e.End
		}
	}
	span := maxEnd - minStart + 1
	if span <= 0 {
		return 0
	}

	covered := make(map[int64]struct{})
	for _, e := range events {
		for p := e.Start; p <= e.End && int64(len(covered)) <= span; p++ {
			covered[p] = struct{}{}
		}
	}
	return float64(len(covered)) / float64(span)
}

// GapsBetween returns the free intervals between events within [spanStart, spanEnd].
func GapsBetween(events []Event, spanStart, spanEnd int64) []interval.Interval {
	if len(events) == 0 {
		return []interval.Interval{{Start: spanStart, End: spanEnd}}
	}

	sorted := make([]Event, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Start < sorted[j].Start
	})

	var gaps []interval.Interval
	cursor := spanStart
	for _, e := range sorted {
		if e.Start > cursor {
			gaps = append(gaps, interval.Interval{Start: cursor, End: e.Start - 1})
		}
		if e.End >= cursor {
			cursor = e.End + 1
		}
	}
	if cursor <= spanEnd {
		gaps = append(gaps, interval.Interval{Start: cursor, End: spanEnd})
	}
	return gaps
}

// MaxOverlap returns the maximum number of events that overlap at any point.
func MaxOverlap(events []Event) int {
	if len(events) == 0 {
		return 0
	}
	type point struct {
		pos   int64
		delta int
	}
	points := make([]point, 0, 2*len(events))
	for _, e := range events {
		points = append(points, point{e.Start, 1})
		points = append(points, point{e.End + 1, -1})
	}
	sort.Slice(points, func(i, j int) bool {
		if points[i].pos != points[j].pos {
			return points[i].pos < points[j].pos
		}
		return points[i].delta < points[j].delta
	})

	max := 0
	cur := 0
	for _, p := range points {
		cur += p.delta
		if cur > max {
			max = cur
		}
	}
	return max
}
