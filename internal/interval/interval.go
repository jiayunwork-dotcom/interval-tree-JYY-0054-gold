// Package interval provides the core interval type and the small set of
// pure helpers used by the rest of the project.
//
// An Interval represents the inclusive integer range [Start, End]. Both
// endpoints belong to the interval, so the length of [a, b] is b - a + 1.
// This inclusive convention is shared by every package in this module
// (the tree stores intervals, the coverage analyzer counts covered
// integers), which keeps the semantics consistent end to end.
//
// The package is intentionally dependency free: it only depends on the
// standard library. All functions are total and never panic on well formed
// inputs; invalid input (End < Start) is reported through error values
// returned by the constructors and parsers rather than via panics.
package interval

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ErrInvalidInterval is returned when an interval has End < Start.
var ErrInvalidInterval = errors.New("interval: invalid interval (End < Start)")

// Interval is an inclusive integer range [Start, End].
//
// Invariant: for a valid interval, Start <= End. Methods that require a
// valid interval (Length, Covers, ...) still compute a sensible value when
// the invariant is broken, but the constructor Parse and the tree insert
// path reject such intervals with ErrInvalidInterval.
type Interval struct {
	Start int64
	End   int64
}

// Valid reports whether the interval satisfies Start <= End.
func (i Interval) Valid() bool {
	return i.Start <= i.End
}

// Length returns the number of integers in the inclusive range.
// For a valid interval [a, b] the result is b - a + 1.
// If End < Start the result is 0.
func (i Interval) Length() int64 {
	if i.End < i.Start {
		return 0
	}
	return i.End - i.Start + 1
}

// Covers reports whether the integer point p lies inside the interval,
// i.e. Start <= p <= End.
func (i Interval) Covers(p int64) bool {
	return i.Start <= p && p <= i.End
}

// Overlaps reports whether two inclusive intervals share at least one
// integer. Two intervals [a, b] and [c, d] overlap when
// max(a, c) <= min(b, d), which is equivalent to a <= d && c <= b.
func (i Interval) Overlaps(o Interval) bool {
	if i.Start > o.End {
		return false
	}
	if o.Start > i.End {
		return false
	}
	return true
}

// Merge returns the smallest interval that contains both i and o, together
// with a boolean reporting whether a merge was possible.
//
// A merge is possible when the two intervals overlap (share at least one
// integer). In that case the merged interval is
// [min(Start, o.Start), max(End, o.End)]. When the intervals do not
// overlap the boolean is false and the returned interval is the zero
// value.
//
// Note that two adjacent intervals such as [1, 5] and [6, 10] do NOT
// overlap under the inclusive definition used here, so they are not
// merged. Callers that want adjacency-aware merging (touching intervals
// become one) should use the coverage package's merge routine instead.
func (i Interval) Merge(o Interval) (Interval, bool) {
	if !i.Overlaps(o) {
		return Interval{}, false
	}
	lo := i.Start
	if o.Start < lo {
		lo = o.Start
	}
	hi := i.End
	if o.End > hi {
		hi = o.End
	}
	return Interval{Start: lo, End: hi}, true
}

// Before reports whether i lies strictly before o, i.e. i.End < o.Start.
func (i Interval) Before(o Interval) bool {
	return i.End < o.Start
}

// After reports whether i lies strictly after o, i.e. i.Start > o.End.
func (i Interval) After(o Interval) bool {
	return i.Start > o.End
}

// CompareKey orders intervals first by Start, then by End. It returns -1,
// 0 or +1. This ordering is the one used by the interval tree as the node
// key, which keeps the tree a proper binary search tree.
func (i Interval) CompareKey(o Interval) int {
	if i.Start != o.Start {
		if i.Start < o.Start {
			return -1
		}
		return 1
	}
	if i.End != o.End {
		if i.End < o.End {
			return -1
		}
		return 1
	}
	return 0
}

// Equal reports whether two intervals have identical endpoints.
func (i Interval) Equal(o Interval) bool {
	return i.Start == o.Start && i.End == o.End
}

// String renders the interval in the canonical "start end" form.
func (i Interval) String() string {
	return fmt.Sprintf("%d %d", i.Start, i.End)
}

// Parse builds an Interval from a string of the form "start end" where
// both endpoints are base-10 int64 values. Whitespace around the tokens is
// tolerated. A line with the wrong number of fields, non numeric fields or
// an inverted range (End < Start) yields ErrInvalidInterval.
func Parse(s string) (Interval, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Interval{}, ErrInvalidInterval
	}
	fields := strings.Fields(s)
	if len(fields) != 2 {
		return Interval{}, fmt.Errorf("%w: expected 2 fields, got %d", ErrInvalidInterval, len(fields))
	}
	start, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return Interval{}, fmt.Errorf("%w: bad start %q", ErrInvalidInterval, fields[0])
	}
	end, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return Interval{}, fmt.Errorf("%w: bad end %q", ErrInvalidInterval, fields[1])
	}
	iv := Interval{Start: start, End: end}
	if !iv.Valid() {
		return Interval{}, ErrInvalidInterval
	}
	return iv, nil
}

// ByStart implements sort.Interface for a slice of intervals ordered by
// their Start endpoint (ties broken by End). It is used to present query
// results in a stable, human readable order.
type ByStart []Interval

func (s ByStart) Len() int           { return len(s) }
func (s ByStart) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByStart) Less(i, j int) bool { return s[i].CompareKey(s[j]) < 0 }

// Sort reorders the receiver in place by Start then End.
func (s ByStart) Sort() { sort.Sort(s) }

// ByEnd implements sort.Interface for a slice of intervals ordered by their
// End endpoint (ties broken by Start). It is primarily useful for the
// sweep line algorithms in the coverage package.
type ByEnd []Interval

func (s ByEnd) Len() int      { return len(s) }
func (s ByEnd) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByEnd) Less(i, j int) bool {
	if s[i].End != s[j].End {
		return s[i].End < s[j].End
	}
	return s[i].Start < s[j].Start
}

// Sort reorders the receiver in place by End then Start.
func (s ByEnd) Sort() { sort.Sort(s) }

// SortStart returns a new slice containing the intervals sorted by Start
// then End. The input is not modified.
func SortStart(in []Interval) []Interval {
	out := append([]Interval(nil), in...)
	ByStart(out).Sort()
	return out
}

// Dedupe returns a copy of in with duplicate intervals removed. The order
// of first appearance is preserved. Two intervals are considered equal only
// when both endpoints match exactly.
func Dedupe(in []Interval) []Interval {
	seen := make(map[Interval]struct{}, len(in))
	out := make([]Interval, 0, len(in))
	for _, iv := range in {
		if _, ok := seen[iv]; ok {
			continue
		}
		seen[iv] = struct{}{}
		out = append(out, iv)
	}
	return out
}
