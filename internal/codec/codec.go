// Package codec provides serialization for intervals.
package codec

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"interval-tree/internal/interval"
)

// ErrFormat is returned when input cannot be parsed.
var ErrFormat = errors.New("codec: format error")

// JSONInterval is the JSON representation of an interval.
type JSONInterval struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

// MarshalJSON serializes intervals to JSON.
func MarshalJSON(ivs []interval.Interval) ([]byte, error) {
	jivs := make([]JSONInterval, len(ivs))
	for i, iv := range ivs {
		jivs[i] = JSONInterval{Start: iv.Start, End: iv.End}
	}
	return json.Marshal(jivs)
}

// UnmarshalJSON deserializes intervals from JSON.
func UnmarshalJSON(data []byte) ([]interval.Interval, error) {
	var jivs []JSONInterval
	if err := json.Unmarshal(data, &jivs); err != nil {
		return nil, err
	}
	ivs := make([]interval.Interval, len(jivs))
	for i, j := range jivs {
		ivs[i] = interval.Interval{Start: j.Start, End: j.End}
	}
	return ivs, nil
}

// WriteBinary writes intervals in binary format.
func WriteBinary(w io.Writer, ivs []interval.Interval) error {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(len(ivs)))
	if _, err := w.Write(buf[:]); err != nil {
		return err
	}
	var pair [16]byte
	for _, iv := range ivs {
		binary.BigEndian.PutUint64(pair[0:8], uint64(iv.Start))
		binary.BigEndian.PutUint64(pair[8:16], uint64(iv.End))
		if _, err := w.Write(pair[:]); err != nil {
			return err
		}
	}
	return nil
}

// ReadBinary reads intervals from binary format.
func ReadBinary(r io.Reader) ([]interval.Interval, error) {
	var buf [4]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return nil, err
	}
	count := binary.BigEndian.Uint32(buf[:])
	if count > 1<<24 {
		return nil, ErrFormat
	}
	ivs := make([]interval.Interval, count)
	var pair [16]byte
	for i := range ivs {
		if _, err := io.ReadFull(r, pair[:]); err != nil {
			return nil, err
		}
		ivs[i] = interval.Interval{
			Start: int64(binary.BigEndian.Uint64(pair[0:8])),
			End:   int64(binary.BigEndian.Uint64(pair[8:16])),
		}
	}
	return ivs, nil
}

// FormatText formats an interval as "[start, end]".
func FormatText(iv interval.Interval) string {
	return fmt.Sprintf("[%d, %d]", iv.Start, iv.End)
}

// ParseText parses "[start, end]".
func ParseText(s string) (interval.Interval, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		return interval.Interval{}, ErrFormat
	}
	s = s[1 : len(s)-1]
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return interval.Interval{}, ErrFormat
	}
	start, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		return interval.Interval{}, ErrFormat
	}
	end, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return interval.Interval{}, ErrFormat
	}
	return interval.Interval{Start: start, End: end}, nil
}

// FormatList formats a list of intervals as text.
func FormatList(ivs []interval.Interval) string {
	var sb strings.Builder
	for _, iv := range ivs {
		sb.WriteString(FormatText(iv))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// ParseList parses multiple lines of "[start, end]".
func ParseList(text string) ([]interval.Interval, error) {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var ivs []interval.Interval
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		iv, err := ParseText(line)
		if err != nil {
			return nil, err
		}
		ivs = append(ivs, iv)
	}
	return ivs, nil
}
