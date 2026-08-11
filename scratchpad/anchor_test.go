package scratchpad

import (
	"strings"
	"testing"
	"time"
)

func mustParseTimestampID(t *testing.T, value string) TimestampID {
	t.Helper()
	id, err := ParseTimestampID(value)
	if err != nil {
		t.Fatalf("ParseTimestampID(%q): %v", value, err)
	}
	return id
}

func TestTimestampIDParsingAndDisplay(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"valid", "000000000000", true},
		{"boundary", "235959999999", true},
		{"short", "12345", false},
		{"long", "1530000000000", false},
		{"hours", "240000000000", false},
		{"minutes", "126000000000", false},
		{"seconds", "125960000000", false},
		{"non-digits", "12ab56789012", false},
		{"leading whitespace", " 153000000000", false},
		{"trailing whitespace", "153000000000 ", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTimestampID(tt.value)
			if tt.ok {
				if err != nil || got.String() != tt.value {
					t.Fatalf("ParseTimestampID(%q) = %q, %v", tt.value, got, err)
				}
				if got.DisplayTime() != tt.value[:2]+":"+tt.value[2:4] {
					t.Fatalf("DisplayTime() = %q", got.DisplayTime())
				}
				return
			}
			if err == nil {
				t.Fatalf("ParseTimestampID(%q) succeeded", tt.value)
			}
		})
	}
}

func TestTimestampIDFromTimeAndAnchor(t *testing.T) {
	id := NewTimestampIDFromTime(time.Date(2025, 1, 15, 13, 45, 30, 123456000, time.UTC))
	if id.String() != "134530123456" {
		t.Fatalf("id = %q", id.String())
	}
	want := `<a id="134530123456" href="#134530123456">13:45</a>`
	if got := id.AnchorHTML(); got != want {
		t.Fatalf("AnchorHTML() = %q, want %q", got, want)
	}
	if got := GenerateTimestampAnchor(time.Date(2025, 1, 15, 13, 45, 30, 123456000, time.UTC)); got != want {
		t.Fatalf("GenerateTimestampAnchor() = %q, want %q", got, want)
	}
}

func TestTimestampIDZeroValue(t *testing.T) {
	var id TimestampID
	if !id.IsZero() {
		t.Fatal("zero TimestampID IsZero() = false")
	}
	if id.String() != "" {
		t.Fatalf("zero TimestampID String() = %q", id.String())
	}
	if id.DisplayTime() != "" {
		t.Fatalf("zero TimestampID DisplayTime() = %q", id.DisplayTime())
	}
	if id.AnchorHTML() != "" {
		t.Fatalf("zero TimestampID AnchorHTML() = %q", id.AnchorHTML())
	}
}

func TestTimestampIDComparableAndMapKey(t *testing.T) {
	first := mustParseTimestampID(t, "153000000000")
	second := mustParseTimestampID(t, "153000000000")
	if first != second {
		t.Fatal("equal parsed TimestampID values are not comparable")
	}
	ids := map[TimestampID]string{first: "value"}
	if ids[second] != "value" {
		t.Fatal("TimestampID cannot be used as an equivalent map key")
	}
	if first.IsZero() {
		t.Fatal("parsed TimestampID IsZero() = true")
	}
}

func TestEntryURL(t *testing.T) {
	id := mustParseTimestampID(t, "153000000000")
	if got := EntryURL("https://example.invalid/posts/1", id); got != "https://example.invalid/posts/1#153000000000" {
		t.Fatalf("EntryURL() = %q", got)
	}
	if got := EntryURL("", id); got != "" {
		t.Fatalf("EntryURL(empty URL) = %q", got)
	}
	if got := EntryURL("https://example.invalid/posts/1", TimestampID{}); got != "" {
		t.Fatalf("EntryURL(empty ID) = %q", got)
	}
}

func TestAnchorHTMLHasMatchingIDAndHref(t *testing.T) {
	anchor := mustParseTimestampID(t, "153000000000").AnchorHTML()
	if !strings.Contains(anchor, `id="153000000000"`) || !strings.Contains(anchor, `href="#153000000000"`) {
		t.Fatalf("anchor = %q", anchor)
	}
}
