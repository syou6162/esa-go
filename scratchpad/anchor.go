package scratchpad

import (
	"fmt"
	"strconv"
	"time"
)

// TimestampID identifies an entry with HHMMSSffffff precision.
//
// Values must be obtained from ParseTimestampID or
// NewTimestampIDFromTime. Behavior for other values, including the zero
// value, is undefined.
type TimestampID string

// ParseTimestampID validates a 12-digit timestamp ID without trimming input.
func ParseTimestampID(value string) (TimestampID, error) {
	if len(value) != 12 {
		return "", fmt.Errorf("must be 12 digits, got %q", value)
	}
	for _, c := range value {
		if c < '0' || c > '9' {
			return "", fmt.Errorf("must be 12 digits, got %q", value)
		}
	}
	hours, err := strconv.Atoi(value[0:2])
	if err != nil {
		return "", fmt.Errorf("invalid hours in timestamp %q: %w", value, err)
	}
	minutes, err := strconv.Atoi(value[2:4])
	if err != nil {
		return "", fmt.Errorf("invalid minutes in timestamp %q: %w", value, err)
	}
	seconds, err := strconv.Atoi(value[4:6])
	if err != nil {
		return "", fmt.Errorf("invalid seconds in timestamp %q: %w", value, err)
	}
	if hours > 23 || minutes > 59 || seconds > 59 {
		return "", fmt.Errorf("time out of range: %s", value)
	}
	return TimestampID(value), nil
}

// NewTimestampIDFromTime creates an ID from the time and six-digit microseconds.
func NewTimestampIDFromTime(t time.Time) TimestampID {
	return TimestampID(t.Format("150405") + fmt.Sprintf("%06d", t.Nanosecond()/1000))
}

// String returns the raw timestamp ID.
func (id TimestampID) String() string {
	return string(id)
}

// DisplayTime returns the timestamp's HH:MM display.
//
// The receiver must have been obtained from ParseTimestampID or
// NewTimestampIDFromTime. Behavior for other values, including the zero
// value, is undefined.
func (id TimestampID) DisplayTime() string {
	s := string(id)
	return s[0:2] + ":" + s[2:4]
}

// AnchorHTML returns the HTML anchor for the timestamp ID.
//
// The receiver must have been obtained from ParseTimestampID or
// NewTimestampIDFromTime. Behavior for other values, including the zero
// value, is undefined.
func (id TimestampID) AnchorHTML() string {
	s := string(id)
	return fmt.Sprintf(`<a id="%s" href="#%s">%s</a>`, s, s, id.DisplayTime())
}

// GenerateTimestampAnchor creates an HTML anchor from a time.
func GenerateTimestampAnchor(t time.Time) string {
	return NewTimestampIDFromTime(t).AnchorHTML()
}

// EntryURL returns postURL#id, or an empty string if either input is empty.
//
// postURL must be a post URL without a fragment. The result is undefined when
// postURL already contains a fragment.
func EntryURL(postURL string, id TimestampID) string {
	if postURL == "" || id == "" {
		return ""
	}
	return postURL + "#" + id.String()
}
