package scratchpad

import (
	"errors"
	"strings"
	"testing"
)

func TestEntriesParseSerializeRoundTrip(t *testing.T) {
	original := Entries{
		{TimestampID: "153000000000", Text: "new\nline"},
		{TimestampID: "150000000000", Text: "old  "},
	}
	body := original.Body()
	parsed, err := ParseEntries(body)
	if err != nil {
		t.Fatalf("ParseEntries() error = %v", err)
	}
	if len(parsed) != 2 || parsed[0].TimestampID != "153000000000" || parsed[1].Text != "old  " {
		t.Fatalf("parsed = %#v", parsed)
	}
	if parsed.Body() != body {
		t.Fatalf("round trip changed body:\n%s\n---\n%s", body, parsed.Body())
	}
}

func TestParseEntriesNormalizesLineEndingsAndRejectsMalformedBlocks(t *testing.T) {
	body := "<a id=\"153000000000\" href=\"#153000000000\">15:30</a> one\r\r---\r\r<a id=\"150000000000\" href=\"#150000000000\">15:00</a> two\r\n\r\n---"
	entries, err := ParseEntries(body)
	if err != nil || len(entries) != 2 {
		t.Fatalf("ParseEntries() = %#v, %v", entries, err)
	}
	if entries[0].Text != "one" || entries[1].Text != "two" {
		t.Fatalf("entries = %#v", entries)
	}
	_, err = ParseEntries("not an entry")
	if err == nil || !strings.Contains(err.Error(), "block 1") {
		t.Fatalf("malformed error = %v", err)
	}
	_, err = ParseEntries(`<a id="250000000000" href="#250000000000">25:00</a> bad`)
	if err == nil {
		t.Fatal("invalid timestamp unexpectedly parsed")
	}
	entries, err = ParseEntries(`<a id="153000000000" href="#150000000001">15:30</a> mismatched`)
	if err != nil || len(entries) != 1 || entries[0].TimestampID != "153000000000" {
		t.Fatalf("mismatched anchor should parse: %#v, %v", entries, err)
	}
}

func TestParseEntriesEmpty(t *testing.T) {
	entries, err := ParseEntries("")
	if err != nil {
		t.Fatalf("ParseEntries() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("len(entries) = %d, want 0", len(entries))
	}
}

func TestParseEntriesPreservesTrailingNewline(t *testing.T) {
	entries := Entries{{TimestampID: "153000000000", Text: "テスト\n"}}
	parsed, err := ParseEntries(entries.Body())
	if err != nil {
		t.Fatalf("ParseEntries() error = %v", err)
	}
	if len(parsed) != 1 || parsed[0].Text != "テスト\n" {
		t.Fatalf("parsed = %#v, want trailing newline preserved", parsed)
	}
}

func TestEntriesSortedAndBodyAreNonMutating(t *testing.T) {
	entries := Entries{
		{TimestampID: "100000000000"},
		{TimestampID: "300000000000"},
		{TimestampID: "200000000000"},
	}
	sorted := entries.Sorted()
	if sorted[0].ID() != "300000000000" || entries[0].ID() != "100000000000" {
		t.Fatalf("sorted = %#v, entries = %#v", sorted, entries)
	}
	if (Entries{}).Body() != "" {
		t.Fatal("empty Body() is not empty")
	}
}

func TestEntriesFindAndFindBy(t *testing.T) {
	entries := Entries{{TimestampID: "153000000000", Text: "one"}, {TimestampID: "160000000000", Text: "two"}}
	found, ok := entries.Find("160000000000")
	if !ok || found.Text != "two" {
		t.Fatalf("Find() = %#v, %v", found, ok)
	}
	found, ok = entries.FindBy(func(entry Entry) bool { return entry.Text == "one" })
	if !ok || found.ID() != "153000000000" {
		t.Fatalf("FindBy() = %#v, %v", found, ok)
	}
	if _, ok := entries.FindBy(nil); ok {
		t.Fatal("FindBy(nil) unexpectedly found an entry")
	}
}

func TestEntriesFindAndFindByReturnFirstDuplicate(t *testing.T) {
	entries := Entries{
		{TimestampID: "153000000000", Text: "first"},
		{TimestampID: "153000000000", Text: "second"},
	}
	found, ok := entries.Find("153000000000")
	if !ok || found.Text != "first" {
		t.Fatalf("Find() = %#v, %v; want first duplicate", found, ok)
	}
	found, ok = entries.FindBy(func(entry Entry) bool {
		return entry.TimestampID == "153000000000"
	})
	if !ok || found.Text != "first" {
		t.Fatalf("FindBy() = %#v, %v; want first duplicate", found, ok)
	}
}

func TestEntriesAddUpdateDeleteCopy(t *testing.T) {
	entries := Entries{{TimestampID: "160000000000", Text: "new"}, {TimestampID: "150000000000", Text: "old"}}
	added := entries.Add(Entry{TimestampID: "155000000000", Text: "middle"})
	if added[1].ID() != "155000000000" || len(entries) != 2 {
		t.Fatalf("Add() = %#v, original = %#v", added, entries)
	}
	replaced := entries.Add(Entry{TimestampID: "160000000000", Text: "updated"})
	if replaced[0].Text != "updated" || &replaced[0] == &entries[0] {
		t.Fatalf("replacement = %#v", replaced)
	}
	updated := entries.Update(Entry{TimestampID: "150000000000", Text: "changed"})
	if updated[1].Text != "changed" || entries[1].Text != "old" {
		t.Fatalf("Update() = %#v, original = %#v", updated, entries)
	}
	noUpdate := entries.Update(Entry{TimestampID: "170000000000"})
	noUpdate[0].Text = "mutated"
	if entries[0].Text != "new" {
		t.Fatal("Update() shared backing array")
	}
	deleted := entries.Delete("160000000000")
	if len(deleted) != 1 || deleted[0].ID() != "150000000000" || len(entries) != 2 {
		t.Fatalf("Delete() = %#v, original = %#v", deleted, entries)
	}
	noDelete := entries.Delete("170000000000")
	noDelete[0].Text = "mutated"
	if entries[0].Text != "new" {
		t.Fatal("Delete() shared backing array")
	}
}

func TestEntriesAddAtBothEnds(t *testing.T) {
	entries := Entries{
		{TimestampID: "153000000000", Text: "middle"},
		{TimestampID: "150000000000", Text: "old"},
	}
	newest := entries.Add(Entry{TimestampID: "160000000000", Text: "new"})
	if newest[0].ID() != "160000000000" {
		t.Fatalf("newest insertion = %#v, want new entry first", newest)
	}
	oldest := entries.Add(Entry{TimestampID: "140000000000", Text: "oldest"})
	if oldest[len(oldest)-1].ID() != "140000000000" {
		t.Fatalf("oldest insertion = %#v, want new entry last", oldest)
	}
}

func TestMakeUniqueTimestamp(t *testing.T) {
	entries := Entries{{TimestampID: "153000000000"}, {TimestampID: "153000000001"}}
	got, err := entries.MakeUniqueTimestamp("153000000000")
	if err != nil || got != "153000000002" {
		t.Fatalf("MakeUniqueTimestamp() = %q, %v", got, err)
	}
	entries = Entries{{TimestampID: "235959999999"}}
	_, err = entries.MakeUniqueTimestamp("235959999999")
	if err == nil {
		t.Fatal("exhausted timestamp candidates unexpectedly succeeded")
	}
	_, err = (Entries{{TimestampID: "not-a-number"}}).MakeUniqueTimestamp("not-a-number")
	if err == nil {
		t.Fatal("invalid colliding timestamp unexpectedly succeeded")
	}
}

func TestMakeUniqueTimestampWithoutConflict(t *testing.T) {
	entries := Entries{{TimestampID: "153000000000", Text: "existing"}}
	got, err := entries.MakeUniqueTimestamp("160000000000")
	if err != nil {
		t.Fatalf("MakeUniqueTimestamp() error = %v", err)
	}
	if got != "160000000000" {
		t.Fatalf("MakeUniqueTimestamp() = %q, want original timestamp", got)
	}
}

func TestEntriesDeleteDuplicateAndLast(t *testing.T) {
	duplicateID := TimestampID("153000000000")
	entries := Entries{
		{TimestampID: duplicateID, Text: "first"},
		{TimestampID: "160000000000", Text: "keep"},
		{TimestampID: duplicateID, Text: "second"},
	}
	deleted := entries.Delete(duplicateID)
	if len(deleted) != 1 || deleted[0].Text != "keep" {
		t.Fatalf("Delete() = %#v, want only non-duplicate entry", deleted)
	}

	only := Entries{{TimestampID: duplicateID, Text: "only"}}
	if got := only.Delete(duplicateID).Body(); got != "" {
		t.Fatalf("Body() after deleting last entry = %q, want empty", got)
	}
}

func TestEntryNotFoundError(t *testing.T) {
	err := &EntryNotFoundError{TimestampID: "153000000000"}
	if !errors.Is(err, ErrEntryNotFound) {
		t.Fatal("EntryNotFoundError does not unwrap to ErrEntryNotFound")
	}
}
