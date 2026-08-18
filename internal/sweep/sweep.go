// Package sweep implements a sweep-line algorithm for interval problems.
package sweep

import (
	"sort"

	"interval-tree/internal/interval"
)

// EventKind distinguishes start from end events.
type EventKind int

const (
	Start EventKind = iota
	End
)

// Event is a sweep-line event at a particular position.
type Event struct {
	Pos      int64
	Kind     EventKind
	Interval interval.Interval
}

// GenerateEvents produces sorted sweep-line events from intervals.
func GenerateEvents(ivs []interval.Interval) []Event {
	events := make([]Event, 0, 2*len(ivs))
	for _, iv := range ivs {
		events = append(events,
			Event{Pos: iv.Start, Kind: Start, Interval: iv},
			Event{Pos: iv.End, Kind: End, Interval: iv},
		)
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].Pos != events[j].Pos {
			return events[i].Pos < events[j].Pos
		}
		return events[i].Kind < events[j].Kind
	})
	return events
}

// CoverageProfile computes how many intervals cover each point in [lo, hi].
func CoverageProfile(ivs []interval.Interval, lo, hi int64) []int {
	if hi < lo {
		return nil
	}
	profile := make([]int, hi-lo+1)
	for _, iv := range ivs {
		s := iv.Start
		e := iv.End
		if s < lo {
			s = lo
		}
		if e > hi {
			e = hi
		}
		for p := s; p <= e; p++ {
			profile[p-lo]++
		}
	}
	return profile
}

// MaxCoverage returns the maximum number of overlapping intervals.
func MaxCoverage(ivs []interval.Interval) int {
	events := GenerateEvents(ivs)
	max := 0
	cur := 0
	for _, ev := range events {
		if ev.Kind == Start {
			cur++
		} else {
			cur--
		}
		if cur > max {
			max = cur
		}
	}
	return max
}

// MinCoverage returns the minimum coverage across the span of all intervals.
func MinCoverage(ivs []interval.Interval) int {
	if len(ivs) == 0 {
		return 0
	}
	lo := ivs[0].Start
	hi := ivs[0].End
	for _, iv := range ivs[1:] {
		if iv.Start < lo {
			lo = iv.Start
		}
		if iv.End > hi {
			hi = iv.End
		}
	}
	profile := CoverageProfile(ivs, lo, hi)
	min := profile[0]
	for _, v := range profile[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// UncoveredPoints returns positions in [lo, hi] not covered by any interval.
func UncoveredPoints(ivs []interval.Interval, lo, hi int64) []int64 {
	profile := CoverageProfile(ivs, lo, hi)
	var uncovered []int64
	for i, v := range profile {
		if v == 0 {
			uncovered = append(uncovered, lo+int64(i))
		}
	}
	return uncovered
}

// CoverageRatio returns the fraction of points in [lo, hi] covered by ≥1 interval.
func CoverageRatio(ivs []interval.Interval, lo, hi int64) float64 {
	profile := CoverageProfile(ivs, lo, hi)
	covered := 0
	for _, v := range profile {
		if v > 0 {
			covered++
		}
	}
	total := int(hi - lo + 1)
	if total <= 0 {
		return 0
	}
	return float64(covered) / float64(total)
}
