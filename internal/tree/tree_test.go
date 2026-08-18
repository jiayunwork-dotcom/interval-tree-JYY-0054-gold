package tree

import (
	"math"
	"testing"

	"interval-tree/internal/interval"
)

// iv is a short constructor for building interval.Interval values in tests.
func iv(s, e int64) interval.Interval {
	return interval.Interval{Start: s, End: e}
}

// checkAugment verifies the AVL + maxEnd invariants for the whole subtree
// rooted at n and returns the maximum End found in that subtree (or
// math.MinInt64 for an empty subtree). It is the brute force oracle used by
// TestTreeMaxEndAugment and the other structural tests.
func checkAugment(n *node) int64 {
	if n == nil {
		return math.MinInt64
	}
	leftMax := checkAugment(n.left)
	rightMax := checkAugment(n.right)

	// BST ordering: left < n < right under CompareKey.
	if n.left != nil && n.left.iv.CompareKey(n.iv) >= 0 {
		panic("BST left child not smaller")
	}
	if n.right != nil && n.right.iv.CompareKey(n.iv) <= 0 {
		panic("BST right child not larger")
	}

	expected := n.iv.End
	if leftMax > expected {
		expected = leftMax
	}
	if rightMax > expected {
		expected = rightMax
	}
	if n.maxEnd != expected {
		panic("maxEnd augmentation mismatch")
	}
	// height consistency.
	h := 1
	if hl := height(n.left); hl > h-1 {
		h = hl + 1
	}
	if hr := height(n.right); hr > h-1 {
		h = hr + 1
	}
	if int(n.height) != h {
		panic("height mismatch")
	}
	// AVL balance property.
	if d := height(n.left) - height(n.right); d > 1 || d < -1 {
		panic("AVL balance violated")
	}
	return expected
}

func mustInsert(t *testing.T, tr *Tree, s, e int64) {
	t.Helper()
	if err := tr.Insert(iv(s, e)); err != nil {
		t.Fatalf("Insert(%d,%d) unexpected error: %v", s, e, err)
	}
}

func TestTreeEmpty(t *testing.T) {
	var tr Tree
	if tr.Len() != 0 {
		t.Error("empty tree should have length 0")
	}
	if tr.MaxEnd() != math.MinInt64 {
		t.Error("empty tree MaxEnd should be MinInt64")
	}
	if got := tr.Stab(5); len(got) != 0 {
		t.Errorf("Stab on empty tree should be empty, got %v", got)
	}
	if got := tr.Overlapping(iv(0, 10)); len(got) != 0 {
		t.Errorf("Overlapping on empty tree should be empty, got %v", got)
	}
	if tr.Contains(iv(1, 2)) {
		t.Error("empty tree should not contain anything")
	}
}

func TestTreeInsertDelete(t *testing.T) {
	var tr Tree
	ivs := []interval.Interval{
		iv(1, 5), iv(10, 20), iv(3, 8), iv(15, 25), iv(0, 2),
	}
	for _, iv := range ivs {
		mustInsert(t, &tr, iv.Start, iv.End)
	}
	if tr.Len() != len(ivs) {
		t.Fatalf("Len = %d, want %d", tr.Len(), len(ivs))
	}
	checkAugment(tr.root)

	// Delete a present interval.
	if err := tr.Delete(iv(3, 8)); err != nil {
		t.Fatalf("Delete present: %v", err)
	}
	if tr.Contains(iv(3, 8)) {
		t.Error("interval should have been deleted")
	}
	if tr.Len() != len(ivs)-1 {
		t.Fatalf("Len after delete = %d, want %d", tr.Len(), len(ivs)-1)
	}
	checkAugment(tr.root)

	// Delete an absent interval -> error, tree unchanged.
	if err := tr.Delete(iv(100, 200)); err == nil {
		t.Error("Delete absent should return an error")
	}
	if tr.Len() != len(ivs)-1 {
		t.Fatalf("Len changed after failed delete: %d", tr.Len())
	}
	checkAugment(tr.root)

	// Remove every remaining interval exactly once.
	for _, v := range ivs {
		if v.Equal(iv(3, 8)) {
			continue // already removed above
		}
		if err := tr.Delete(v); err != nil {
			t.Fatalf("Delete %v: %v", v, err)
		}
		checkAugment(tr.root)
	}
	if tr.Len() != 0 || tr.root != nil {
		t.Fatalf("tree should be empty, Len = %d, root = %v", tr.Len(), tr.root)
	}

	// Deleting from an empty tree errors.
	if err := tr.Delete(iv(1, 5)); err == nil {
		t.Error("Delete on empty tree should error")
	}

	// Clear leaves the tree reusable.
	mustInsert(t, &tr, 1, 2)
	tr.Clear()
	if tr.Len() != 0 || tr.root != nil {
		t.Error("Clear did not empty the tree")
	}
}

func TestTreeDuplicateMultiplicity(t *testing.T) {
	var tr Tree
	mustInsert(t, &tr, 1, 5)
	mustInsert(t, &tr, 1, 5)
	mustInsert(t, &tr, 1, 5)
	if tr.Len() != 3 {
		t.Fatalf("Len with duplicates = %d, want 3", tr.Len())
	}
	if got := tr.Stab(3); len(got) != 3 {
		t.Fatalf("Stab covering point should return 3 copies, got %d", len(got))
	}
	// Removing one of three leaves two behind.
	if err := tr.Delete(iv(1, 5)); err != nil {
		t.Fatal(err)
	}
	if tr.Len() != 2 {
		t.Fatalf("Len after one delete = %d, want 2", tr.Len())
	}
	if got := tr.Stab(3); len(got) != 2 {
		t.Fatalf("Stab should return 2 copies, got %d", len(got))
	}
	checkAugment(tr.root)
	// Remove the remaining two.
	_ = tr.Delete(iv(1, 5))
	_ = tr.Delete(iv(1, 5))
	if tr.Len() != 0 {
		t.Fatalf("Len after removing all = %d, want 0", tr.Len())
	}
	if err := tr.Delete(iv(1, 5)); err == nil {
		t.Error("deleting a fully removed interval should error")
	}
}

func TestTreeStab(t *testing.T) {
	var tr Tree
	// Build a small set with deliberate overlaps.
	in := []interval.Interval{
		iv(0, 10), iv(5, 15), iv(12, 20), iv(25, 30), iv(-5, 2), iv(100, 110),
	}
	for _, iv := range in {
		mustInsert(t, &tr, iv.Start, iv.End)
	}

	cases := []struct {
		point int64
		want  []interval.Interval
	}{
		{0, []interval.Interval{iv(-5, 2), iv(0, 10)}},
		{1, []interval.Interval{iv(-5, 2), iv(0, 10)}},
		{5, []interval.Interval{iv(0, 10), iv(5, 15)}},
		{10, []interval.Interval{iv(0, 10), iv(5, 15)}},
		{11, []interval.Interval{iv(5, 15)}},
		{13, []interval.Interval{iv(5, 15), iv(12, 20)}},
		{20, []interval.Interval{iv(12, 20)}},
		{21, nil},
		{27, []interval.Interval{iv(25, 30)}},
		{105, []interval.Interval{iv(100, 110)}},
		{-5, []interval.Interval{iv(-5, 2)}},
		{-6, nil},
	}
	for _, c := range cases {
		got := tr.Stab(c.point)
		if len(got) != len(c.want) {
			t.Errorf("Stab(%d) len = %d, want %d (%v)", c.point, len(got), len(c.want), got)
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("Stab(%d)[%d] = %v, want %v", c.point, i, got[i], c.want[i])
			}
		}
	}
}

func TestTreeOverlapping(t *testing.T) {
	var tr Tree
	in := []interval.Interval{
		iv(0, 10), iv(5, 15), iv(20, 30), iv(25, 35), iv(40, 50),
	}
	for _, iv := range in {
		mustInsert(t, &tr, iv.Start, iv.End)
	}

	cases := []struct {
		q    interval.Interval
		want []interval.Interval
	}{
		{iv(3, 7), []interval.Interval{iv(0, 10), iv(5, 15)}},  // inside first two
		{iv(11, 19), []interval.Interval{iv(5, 15)}},           // only middle
		{iv(16, 19), nil},                                     // gap before 20
		{iv(20, 30), []interval.Interval{iv(20, 30), iv(25, 35)}}, // touches
		{iv(33, 45), []interval.Interval{iv(25, 35), iv(40, 50)}},
		{iv(100, 200), nil},
		{iv(-10, -1), nil},
		{iv(0, 50), in},
	}
	for _, c := range cases {
		got := tr.Overlapping(c.q)
		if len(got) != len(c.want) {
			t.Errorf("Overlapping(%v) len = %d, want %d (%v)", c.q, len(got), len(c.want), got)
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("Overlapping(%v)[%d] = %v, want %v", c.q, i, got[i], c.want[i])
			}
		}
	}
}

func TestTreeMaxEndAugment(t *testing.T) {
	var tr Tree
	// Insert in ascending order (worst case for an unbalanced BST) to force
	// many AVL rotations, then verify the augmentation after every step.
	ins := []interval.Interval{}
	for i := 0; i < 200; i++ {
		iv := iv(int64(i*2), int64(i*2+5+(i%7)))
		ins = append(ins, iv)
		mustInsert(t, &tr, iv.Start, iv.End)
		checkAugment(tr.root)
	}
	// Global maxEnd must equal the largest End among all insertions.
	var wantMax int64 = math.MinInt64
	for _, iv := range ins {
		if iv.End > wantMax {
			wantMax = iv.End
		}
	}
	if tr.MaxEnd() != wantMax {
		t.Errorf("MaxEnd = %d, want %d", tr.MaxEnd(), wantMax)
	}

	// Delete half the intervals in a scrambled order and re-check.
	for i := 199; i >= 0; i -= 2 {
		if err := tr.Delete(ins[i]); err != nil {
			t.Fatalf("delete %v: %v", ins[i], err)
		}
		checkAugment(tr.root)
	}
	// Recompute expected global max from survivors.
	wantMax = math.MinInt64
	for i := 0; i < 200; i += 2 {
		if ins[i].End > wantMax {
			wantMax = ins[i].End
		}
	}
	if tr.MaxEnd() != wantMax {
		t.Errorf("MaxEnd after deletes = %d, want %d", tr.MaxEnd(), wantMax)
	}
	checkAugment(tr.root)
}

func TestTreeAll(t *testing.T) {
	var tr Tree
	mustInsert(t, &tr, 5, 9)
	mustInsert(t, &tr, 1, 2)
	mustInsert(t, &tr, 1, 2) // duplicate
	mustInsert(t, &tr, 3, 8)
	all := tr.All()
	want := []interval.Interval{iv(1, 2), iv(1, 2), iv(3, 8), iv(5, 9)}
	if len(all) != len(want) {
		t.Fatalf("All len = %d, want %d (%v)", len(all), len(want), all)
	}
	for i := range want {
		if all[i] != want[i] {
			t.Errorf("All()[%d] = %v, want %v", i, all[i], want[i])
		}
	}
}

func TestTreeErrors(t *testing.T) {
	var tr Tree
	// Invalid interval (End < Start) is rejected by Insert.
	if err := tr.Insert(iv(10, 1)); err == nil {
		t.Error("Insert with End < Start should error")
	}
	if tr.Len() != 0 {
		t.Error("failed Insert should not change tree")
	}
	// Delete on empty and absent.
	if err := tr.Delete(iv(1, 5)); err == nil {
		t.Error("Delete on empty tree should error")
	}
	mustInsert(t, &tr, 1, 5)
	if err := tr.Delete(iv(6, 9)); err == nil {
		t.Error("Delete absent should error")
	}
	// Delete invalid.
	if err := tr.Delete(iv(9, 1)); err == nil {
		t.Error("Delete invalid interval should error")
	}
}
