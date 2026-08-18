package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestReorder(t *testing.T) {
	cases := []struct {
		in   []string
		want []string
	}{
		{[]string{"stab", "5", "-file", "data.txt"}, []string{"-file", "data.txt", "stab", "5"}},
		{[]string{"count", "-file", "data.txt"}, []string{"-file", "data.txt", "count"}},
		{[]string{"gaps", "1", "10", "-file", "d.txt"}, []string{"-file", "d.txt", "gaps", "1", "10"}},
		{[]string{"union", "--flag=x"}, []string{"--flag=x", "union"}},
		{[]string{"count"}, []string{"count"}},
	}
	for _, c := range cases {
		got := reorder(c.in)
		if len(got) != len(c.want) {
			t.Errorf("reorder(%v) len = %d, want %d (%v)", c.in, len(got), len(c.want), got)
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("reorder(%v)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "intervals.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// capture runs fn with stdout redirected to a buffer and returns the output.
func capture(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	runErr := fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String(), runErr
}

func TestMainCount(t *testing.T) {
	path := writeTemp(t, "1 5\n10 20\n# comment\n30 40\n")
	out, err := capture(t, func() error {
		return run([]string{"count", "-file", path})
	})
	if err != nil {
		t.Fatalf("run count: %v", err)
	}
	if out != "3\n" {
		t.Errorf("count output = %q, want %q", out, "3\n")
	}
}

func TestMainUnion(t *testing.T) {
	path := writeTemp(t, "1 5\n6 10\n20 25\n")
	out, err := capture(t, func() error {
		return run([]string{"union", "-file", path})
	})
	if err != nil {
		t.Fatalf("run union: %v", err)
	}
	// [1,5]=5, [6,10]=5, [20,25]=6 -> 16
	if out != "16\n" {
		t.Errorf("union output = %q, want %q", out, "16\n")
	}
}

func TestMainGaps(t *testing.T) {
	path := writeTemp(t, "1 5\n7 10\n")
	out, err := capture(t, func() error {
		return run([]string{"gaps", "1", "10", "-file", path})
	})
	if err != nil {
		t.Fatalf("run gaps: %v", err)
	}
	want := "6 6\n"
	if out != want {
		t.Errorf("gaps output = %q, want %q", out, want)
	}
}

func TestMainStab(t *testing.T) {
	path := writeTemp(t, "1 5\n3 8\n10 20\n")
	out, err := capture(t, func() error {
		return run([]string{"stab", "4", "-file", path})
	})
	if err != nil {
		t.Fatalf("run stab: %v", err)
	}
	want := "1 5\n3 8\n"
	if out != want {
		t.Errorf("stab output = %q, want %q", out, want)
	}
}

func TestMainOverlaps(t *testing.T) {
	path := writeTemp(t, "1 5\n3 8\n10 20\n")
	out, err := capture(t, func() error {
		return run([]string{"overlaps", "4", "-file", path})
	})
	if err != nil {
		t.Fatalf("run overlaps: %v", err)
	}
	want := "1 5\n3 8\n"
	if out != want {
		t.Errorf("overlaps output = %q, want %q", out, want)
	}
}

func TestMainInlineIntervals(t *testing.T) {
	out, err := capture(t, func() error {
		return run([]string{"union", "1", "5", "6", "10"})
	})
	if err != nil {
		t.Fatalf("run inline union: %v", err)
	}
	if out != "10\n" {
		t.Errorf("inline union output = %q, want %q", out, "10\n")
	}
}

func TestMainUnknownCommand(t *testing.T) {
	if err := run([]string{"frobnicate"}); err == nil {
		t.Error("unknown command should return an error")
	}
}

func TestMainMissingFile(t *testing.T) {
	_, err := capture(t, func() error {
		return run([]string{"count", "-file", "/no/such/file.txt"})
	})
	if err == nil {
		t.Error("missing file should return an error")
	}
}

func Example_reorder() {
	fmt.Println(reorder([]string{"stab", "5", "-file", "data.txt"}))
	// Output: [-file data.txt stab 5]
}
