package codec

import (
	"bytes"
	"testing"

	"interval-tree/internal/interval"
)

func iv(s, e int64) interval.Interval {
	return interval.Interval{Start: s, End: e}
}

func TestJSONRoundTrip(t *testing.T) {
	ivs := []interval.Interval{iv(1, 5), iv(10, 20)}
	data, err := MarshalJSON(ivs)
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Start != 1 || got[1].End != 20 {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestBinaryRoundTrip(t *testing.T) {
	ivs := []interval.Interval{iv(0, 100), iv(200, 300)}
	var buf bytes.Buffer
	if err := WriteBinary(&buf, ivs); err != nil {
		t.Fatal(err)
	}
	got, err := ReadBinary(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].End != 100 || got[1].Start != 200 {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestTextFormat(t *testing.T) {
	s := FormatText(iv(3, 7))
	if s != "[3, 7]" {
		t.Fatalf("got %q", s)
	}
}

func TestTextParse(t *testing.T) {
	got, err := ParseText("[3, 7]")
	if err != nil {
		t.Fatal(err)
	}
	if got.Start != 3 || got.End != 7 {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestListRoundTrip(t *testing.T) {
	ivs := []interval.Interval{iv(1, 2), iv(5, 10)}
	text := FormatList(ivs)
	got, err := ParseList(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}
