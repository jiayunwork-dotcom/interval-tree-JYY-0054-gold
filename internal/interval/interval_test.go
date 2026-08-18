package interval

import (
	"math"
	"testing"
)

// rng returns the inclusive list of integer points covered by iv, used by
// several tests to cross check the higher level predicates against a brute
// force enumeration.
func points(iv Interval) []int64 {
	if !iv.Valid() {
		return nil
	}
	out := make([]int64, 0, iv.Length())
	for p := iv.Start; p <= iv.End; p++ {
		out = append(out, p)
	}
	return out
}

func TestIntervalLength(t *testing.T) {
	cases := []struct {
		iv   Interval
		want int64
	}{
		{Interval{0, 0}, 1},
		{Interval{1, 5}, 5},
		{Interval{-3, 3}, 7},
		{Interval{10, 20}, 11},
		{Interval{5, 1}, 0}, // invalid: End < Start
		{Interval{math.MinInt64, math.MinInt64}, 1},
		{Interval{math.MaxInt64 - 2, math.MaxInt64}, 3},
	}
	for _, c := range cases {
		if got := c.iv.Length(); got != c.want {
			t.Errorf("Length(%v) = %d, want %d", c.iv, got, c.want)
		}
	}
}

func TestIntervalValid(t *testing.T) {
	if (Interval{5, 1}).Valid() {
		t.Error("interval [5,1] should be invalid")
	}
	if !(Interval{1, 5}).Valid() {
		t.Error("interval [1,5] should be valid")
	}
	if !(Interval{3, 3}).Valid() {
		t.Error("degenerate interval [3,3] should be valid")
	}
}

func TestIntervalCovers(t *testing.T) {
	iv := Interval{10, 20}
	if !iv.Covers(10) || !iv.Covers(15) || !iv.Covers(20) {
		t.Error("interval should cover its endpoints and interior")
	}
	if iv.Covers(9) || iv.Covers(21) {
		t.Error("interval should not cover points outside [10,20]")
	}
}

func TestIntervalOverlaps(t *testing.T) {
	cases := []struct {
		a, b Interval
		want bool
	}{
		{Interval{1, 5}, Interval{3, 8}, true},    // partial overlap
		{Interval{1, 5}, Interval{5, 8}, true},    // touch at a point
		{Interval{1, 5}, Interval{6, 8}, false},   // adjacent, no shared point
		{Interval{1, 5}, Interval{0, 0}, false},  // strictly before
		{Interval{1, 5}, Interval{10, 12}, false}, // strictly after
		{Interval{1, 10}, Interval{3, 4}, true},   // containment
		{Interval{3, 4}, Interval{1, 10}, true},   // reverse containment
		{Interval{0, 100}, Interval{50, 50}, true}, // degenerate inside
		{Interval{-10, -1}, Interval{-5, 5}, true}, // negative overlap
		{Interval{-10, -5}, Interval{0, 5}, false}, // negative no overlap
	}
	for _, c := range cases {
		if got := c.a.Overlaps(c.b); got != c.want {
			t.Errorf("%v.Overlaps(%v) = %v, want %v", c.a, c.b, got, c.want)
		}
		// Overlap is symmetric.
		if got := c.b.Overlaps(c.a); got != c.want {
			t.Errorf("symmetric %v.Overlaps(%v) = %v, want %v", c.b, c.a, got, c.want)
		}
	}
}

func TestIntervalMerge(t *testing.T) {
	cases := []struct {
		a, b Interval
		want Interval
		ok   bool
	}{
		{Interval{1, 5}, Interval{3, 8}, Interval{1, 8}, true},
		{Interval{1, 5}, Interval{5, 8}, Interval{1, 8}, true},     // touch
		{Interval{1, 5}, Interval{6, 8}, Interval{}, false},        // gap
		{Interval{1, 10}, Interval{3, 4}, Interval{1, 10}, true},   // contained
		{Interval{3, 4}, Interval{1, 10}, Interval{1, 10}, true},   // contains
		{Interval{-5, 1}, Interval{0, 9}, Interval{-5, 9}, true},   // negative
	}
	for _, c := range cases {
		got, ok := c.a.Merge(c.b)
		if ok != c.ok {
			t.Errorf("%v.Merge(%v) ok = %v, want %v", c.a, c.b, ok, c.ok)
			continue
		}
		if ok && got != c.want {
			t.Errorf("%v.Merge(%v) = %v, want %v", c.a, c.b, got, c.want)
		}
		if ok {
			// The merged interval must contain every point of both inputs.
			for _, p := range append(points(c.a), points(c.b)...) {
				if !got.Covers(p) {
					t.Errorf("merged %v does not cover point %d from inputs %v %v", got, p, c.a, c.b)
				}
			}
		}
	}
}

func TestIntervalBeforeAfter(t *testing.T) {
	a := Interval{1, 5}
	b := Interval{10, 20}
	if !a.Before(b) {
		t.Error("a should be before b")
	}
	if !b.After(a) {
		t.Error("b should be after a")
	}
	if a.Before(a) {
		t.Error("interval is not before itself")
	}
	touch := Interval{6, 9}
	// Adjacent intervals (5 then 6) are strictly before: a.End < touch.Start.
	if !a.Before(touch) {
		t.Error("adjacent intervals should be strictly before")
	}
}

func TestIntervalCompareKey(t *testing.T) {
	cases := []struct {
		a, b Interval
		want int
	}{
		{Interval{1, 5}, Interval{1, 5}, 0},
		{Interval{1, 5}, Interval{2, 5}, -1},
		{Interval{2, 5}, Interval{1, 5}, 1},
		{Interval{1, 4}, Interval{1, 5}, -1},
		{Interval{1, 6}, Interval{1, 5}, 1},
	}
	for _, c := range cases {
		if got := c.a.CompareKey(c.b); got != c.want {
			t.Errorf("%v.CompareKey(%v) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestIntervalString(t *testing.T) {
	if got := (Interval{12, 34}).String(); got != "12 34" {
		t.Errorf("String() = %q, want %q", got, "12 34")
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		in    string
		want  Interval
		isErr bool
	}{
		{"1 5", Interval{1, 5}, false},
		{"  -3   3 ", Interval{-3, 3}, false}, // surrounding whitespace
		{"0 0", Interval{0, 0}, false},
		{"", Interval{}, true},
		{"1", Interval{}, true},     // too few fields
		{"1 2 3", Interval{}, true}, // too many fields
		{"a 5", Interval{}, true},   // non numeric
		{"5 1", Interval{}, true},   // inverted range
		{"1.5 2", Interval{}, true}, // float not allowed
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if c.isErr {
			if err == nil {
				t.Errorf("Parse(%q) expected error, got %v", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("Parse(%q) unexpected error: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("Parse(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestSortStart(t *testing.T) {
	in := []Interval{
		{5, 9}, {1, 2}, {1, 1}, {3, 8}, {1, 10},
	}
	want := []Interval{
		{1, 1}, {1, 2}, {1, 10}, {3, 8}, {5, 9},
	}
	got := SortStart(in)
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("SortStart()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
	// Input must be untouched.
	if in[0] != (Interval{5, 9}) {
		t.Error("SortStart modified the input slice")
	}
}

func TestDedupe(t *testing.T) {
	in := []Interval{{1, 2}, {3, 4}, {1, 2}, {5, 6}, {3, 4}, {1, 2}}
	got := Dedupe(in)
	want := []Interval{{1, 2}, {3, 4}, {5, 6}}
	if len(got) != len(want) {
		t.Fatalf("Dedupe length = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Dedupe()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}
