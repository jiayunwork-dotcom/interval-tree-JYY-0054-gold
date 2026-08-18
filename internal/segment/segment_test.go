package segment

import (
	"testing"
)

func TestPointUpdate(t *testing.T) {
	tr := New(10)
	tr.Update(3, 3, 5)
	if tr.Query(3) != 5 {
		t.Fatalf("expected 5, got %d", tr.Query(3))
	}
	if tr.Query(4) != 0 {
		t.Fatalf("expected 0, got %d", tr.Query(4))
	}
}

func TestRangeUpdate(t *testing.T) {
	tr := New(10)
	tr.Update(2, 7, 3)
	for i := 2; i <= 7; i++ {
		if tr.Query(i) != 3 {
			t.Fatalf("pos %d: expected 3, got %d", i, tr.Query(i))
		}
	}
	if tr.Query(1) != 0 || tr.Query(8) != 0 {
		t.Fatal("outside range should be 0")
	}
}

func TestMultipleUpdates(t *testing.T) {
	tr := New(10)
	tr.Update(1, 5, 2)
	tr.Update(3, 8, 3)
	// [1,2] = 2, [3,5] = 5, [6,8] = 3, rest = 0
	if tr.Query(2) != 2 {
		t.Fatalf("pos 2: expected 2, got %d", tr.Query(2))
	}
	if tr.Query(4) != 5 {
		t.Fatalf("pos 4: expected 5, got %d", tr.Query(4))
	}
	if tr.Query(7) != 3 {
		t.Fatalf("pos 7: expected 3, got %d", tr.Query(7))
	}
}

func TestRangeSum(t *testing.T) {
	tr := New(5)
	tr.Update(0, 4, 1) // all positions have value 1
	sum := tr.RangeSum(0, 4)
	if sum != 5 {
		t.Fatalf("expected 5, got %d", sum)
	}
}

func TestRangeMax(t *testing.T) {
	tr := New(10)
	tr.Update(0, 9, 1)
	tr.Update(5, 5, 10)
	max := tr.RangeMax(0, 9)
	if max != 11 {
		t.Fatalf("expected 11, got %d", max)
	}
}

func TestSnapshot(t *testing.T) {
	tr := New(5)
	tr.Update(1, 3, 7)
	snap := tr.Snapshot()
	if snap[0] != 0 || snap[1] != 7 || snap[2] != 7 || snap[3] != 7 || snap[4] != 0 {
		t.Fatalf("unexpected snapshot: %v", snap)
	}
}

func TestReset(t *testing.T) {
	tr := New(5)
	tr.Update(0, 4, 99)
	tr.Reset()
	for i := 0; i < 5; i++ {
		if tr.Query(i) != 0 {
			t.Fatalf("expected 0 after reset at %d", i)
		}
	}
}
