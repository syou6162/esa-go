package scratchpad

import "strings"

// Entry is one timestamped item in a scratchpad post.
type Entry struct {
	TimestampID TimestampID
	Text        string
}

// ID returns the entry's timestamp ID as a string.
func (e Entry) ID() string {
	return e.TimestampID.String()
}

// DisplayTime returns the entry's timestamp as HH:MM.
func (e Entry) DisplayTime() string {
	return e.TimestampID.DisplayTime()
}

// AnchorHTML returns the entry's timestamp anchor.
func (e Entry) AnchorHTML() string {
	return e.TimestampID.AnchorHTML()
}

// FirstLine returns the first non-empty, trimmed line of text.
func (e Entry) FirstLine() string {
	for _, line := range strings.Split(e.Text, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// Body returns the anchor, a space, and the raw entry text.
//
// A zero-value TimestampID produces an empty anchor followed by a space and
// the raw text. Serializing such an Entry is a caller programming error; the
// empty anchor is not a fallback representation.
func (e Entry) Body() string {
	return e.AnchorHTML() + " " + e.Text
}

// WithText returns a copy with the same timestamp and replacement text.
func (e Entry) WithText(text string) Entry {
	return Entry{TimestampID: e.TimestampID, Text: text}
}

// IsSameTimestamp reports whether two entries have the same timestamp.
func (e Entry) IsSameTimestamp(other Entry) bool {
	return e.TimestampID == other.TimestampID
}

// IsSameTimestampID reports whether ts equals the entry's timestamp.
func (e Entry) IsSameTimestampID(ts TimestampID) bool {
	return e.TimestampID == ts
}
