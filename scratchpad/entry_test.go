package scratchpad

import "testing"

func TestEntryMethods(t *testing.T) {
	entry := Entry{TimestampID: mustParseTimestampID(t, "153000000000"), Text: "\n  first line  \nsecond line"}
	if entry.ID() != "153000000000" {
		t.Fatalf("ID() = %q", entry.ID())
	}
	if entry.DisplayTime() != "15:30" {
		t.Fatalf("DisplayTime() = %q", entry.DisplayTime())
	}
	if entry.FirstLine() != "first line" {
		t.Fatalf("FirstLine() = %q", entry.FirstLine())
	}
	if entry.Body() != `<a id="153000000000" href="#153000000000">15:30</a> `+"\n  first line  \nsecond line" {
		t.Fatalf("Body() = %q", entry.Body())
	}
	updated := entry.WithText("updated")
	if updated.TimestampID != entry.TimestampID || updated.Text != "updated" || entry.Text == "updated" {
		t.Fatalf("WithText() changed values: %#v %#v", entry, updated)
	}
	if !entry.IsSameTimestamp(Entry{TimestampID: entry.TimestampID, Text: "other"}) {
		t.Fatal("IsSameTimestamp() = false")
	}
	if !entry.IsSameTimestampID(entry.TimestampID) {
		t.Fatal("IsSameTimestampID() = false")
	}
	if entry.IsSameTimestamp(Entry{TimestampID: mustParseTimestampID(t, "160000000000")}) {
		t.Fatal("IsSameTimestamp() = true for different timestamp")
	}
	if entry.IsSameTimestampID(mustParseTimestampID(t, "160000000000")) {
		t.Fatal("IsSameTimestampID() = true for different timestamp")
	}
}

func TestEntryBodyWithZeroTimestampID(t *testing.T) {
	if got := (Entry{Text: "text"}).Body(); got != " text" {
		t.Fatalf("Body() with zero TimestampID = %q, want %q", got, " text")
	}
}

func TestEntryFirstLineEmpty(t *testing.T) {
	if got := (Entry{Text: " \n\t "}).FirstLine(); got != "" {
		t.Fatalf("FirstLine() = %q", got)
	}
}
