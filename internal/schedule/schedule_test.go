package schedule

import (
	"testing"
)

func TestFindConflicts(t *testing.T) {
	events := []Event{
		{Name: "A", Start: 1, End: 5},
		{Name: "B", Start: 3, End: 7},
		{Name: "C", Start: 8, End: 10},
	}
	conflicts := FindConflicts(events)
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
}

func TestHasConflict(t *testing.T) {
	noConflict := []Event{
		{Name: "A", Start: 1, End: 3},
		{Name: "B", Start: 4, End: 6},
	}
	if HasConflict(noConflict) {
		t.Fatal("should not have conflict")
	}
	withConflict := []Event{
		{Name: "A", Start: 1, End: 5},
		{Name: "B", Start: 3, End: 7},
	}
	if !HasConflict(withConflict) {
		t.Fatal("should have conflict")
	}
}

func TestMaxNonOverlapping(t *testing.T) {
	events := []Event{
		{Name: "A", Start: 1, End: 3},
		{Name: "B", Start: 2, End: 5},
		{Name: "C", Start: 4, End: 6},
		{Name: "D", Start: 7, End: 9},
	}
	selected := MaxNonOverlapping(events)
	if len(selected) != 3 {
		t.Fatalf("expected 3, got %d", len(selected))
	}
}

func TestMaxOverlap(t *testing.T) {
	events := []Event{
		{Start: 1, End: 5},
		{Start: 2, End: 6},
		{Start: 4, End: 8},
		{Start: 7, End: 9},
	}
	max := MaxOverlap(events)
	if max != 3 {
		t.Fatalf("expected 3, got %d", max)
	}
}

func TestGapsBetween(t *testing.T) {
	events := []Event{
		{Start: 2, End: 4},
		{Start: 7, End: 9},
	}
	gaps := GapsBetween(events, 1, 12)
	if len(gaps) != 3 {
		t.Fatalf("expected 3 gaps, got %d", len(gaps))
	}
}

func TestUtilization(t *testing.T) {
	events := []Event{{Start: 1, End: 10}}
	u := Utilization(events)
	if u != 1.0 {
		t.Fatalf("expected 1.0, got %f", u)
	}
}

func TestUtilizationPartial(t *testing.T) {
	events := []Event{
		{Start: 1, End: 5},
		{Start: 8, End: 10},
	}
	u := Utilization(events)
	if u < 0.79 || u > 0.81 {
		t.Fatalf("expected ~0.8, got %f", u)
	}
}
