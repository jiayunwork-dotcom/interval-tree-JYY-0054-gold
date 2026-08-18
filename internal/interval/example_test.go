package interval

import (
	"fmt"
)

// ExampleInterval_Overlaps shows the inclusive overlap predicate.
func ExampleInterval_Overlaps() {
	a := Interval{Start: 1, End: 5}
	b := Interval{Start: 3, End: 8}
	c := Interval{Start: 6, End: 9}
	fmt.Println(a.Overlaps(b))
	fmt.Println(a.Overlaps(c))
	// Output:
	// true
	// false
}

// ExampleInterval_Merge shows that only overlapping intervals merge.
func ExampleInterval_Merge() {
	a := Interval{Start: 1, End: 5}
	b := Interval{Start: 3, End: 8}
	merged, ok := a.Merge(b)
	fmt.Println(merged, ok)
	// Output: 1 8 true
}

// ExampleParse shows reading an interval from text.
func ExampleParse() {
	iv, err := Parse("12 34")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(iv, iv.Length())
	// Output: 12 34 23
}

// ExampleByStart_Sort shows ordering by Start then End.
func ExampleByStart_Sort() {
	in := []Interval{{5, 9}, {1, 2}, {1, 10}, {3, 8}}
	ByStart(in).Sort()
	fmt.Println(in)
	// Output: [1 2 1 10 3 8 5 9]
}
