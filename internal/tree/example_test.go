package tree

import (
	"fmt"

	"interval-tree/internal/interval"
)

// ExampleTree_Stab shows finding all intervals covering a point.
func ExampleTree_Stab() {
	var tr Tree
	_ = tr.Insert(interval.Interval{Start: 0, End: 10})
	_ = tr.Insert(interval.Interval{Start: 5, End: 15})
	for _, iv := range tr.Stab(7) {
		fmt.Println(iv)
	}
	// Output:
	// 0 10
	// 5 15
}

// ExampleTree_Insert shows the basic insert / query lifecycle.
func ExampleTree_Insert() {
	var tr Tree
	_ = tr.Insert(interval.Interval{Start: 1, End: 5})
	_ = tr.Insert(interval.Interval{Start: 10, End: 20})
	fmt.Println(tr.Len())
	fmt.Println(tr.Contains(interval.Interval{Start: 1, End: 5}))
	// Output:
	// 2
	// true
}
