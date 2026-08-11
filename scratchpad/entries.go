package scratchpad

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var anchorRE = regexp.MustCompile(`(?s)^<a id="(\d+)" href="#\d+">[^<]+</a> (.*)`)

// Entries is a collection of scratchpad entries.
type Entries []Entry

// ParseEntries parses Markdown entries, normalizing CRLF and CR to LF.
func ParseEntries(bodyMD string) (Entries, error) {
	normalized := strings.ReplaceAll(strings.ReplaceAll(bodyMD, "\r\n", "\n"), "\r", "\n")
	var entries Entries
	for i, block := range strings.Split(normalized, "\n---") {
		trimmed := strings.TrimLeft(block, "\n")
		if strings.HasSuffix(trimmed, "\n") {
			trimmed = trimmed[:len(trimmed)-1]
		}
		if trimmed == "" {
			continue
		}
		matches := anchorRE.FindStringSubmatch(trimmed)
		if matches == nil {
			return nil, fmt.Errorf("malformed scratchpad entry at block %d: does not match anchor format", i+1)
		}
		timestampID, err := ParseTimestampID(matches[1])
		if err != nil {
			return nil, fmt.Errorf("malformed scratchpad entry at block %d: invalid timestamp ID: %w", i+1, err)
		}
		entries = append(entries, Entry{TimestampID: timestampID, Text: matches[2]})
	}
	return entries, nil
}

// Body serializes entries newest-first with a trailing separator per entry.
func (e Entries) Body() string {
	if len(e) == 0 {
		return ""
	}
	sorted := e.Sorted()
	parts := make([]string, len(sorted))
	for i, entry := range sorted {
		parts[i] = entry.Body() + "\n\n---"
	}
	return strings.Join(parts, "\n\n")
}

// Sorted returns a descending TimestampID copy without changing the receiver.
func (e Entries) Sorted() Entries {
	if len(e) == 0 {
		return Entries{}
	}
	result := make(Entries, len(e))
	copy(result, e)
	sort.Slice(result, func(i, j int) bool {
		return result[i].TimestampID.value > result[j].TimestampID.value
	})
	return result
}

// Find returns the first entry with ts.
func (e Entries) Find(ts TimestampID) (Entry, bool) {
	for _, entry := range e {
		if entry.IsSameTimestampID(ts) {
			return entry, true
		}
	}
	return Entry{}, false
}

// FindBy returns the first entry for which match returns true.
func (e Entries) FindBy(match func(Entry) bool) (Entry, bool) {
	if match == nil {
		return Entry{}, false
	}
	for _, entry := range e {
		if match(entry) {
			return entry, true
		}
	}
	return Entry{}, false
}

// Add inserts or replaces an entry in a descending receiver.
//
// The receiver must already be sorted descending; call Sorted first when its
// order is not guaranteed.
func (e Entries) Add(entry Entry) Entries {
	for i, existing := range e {
		if existing.IsSameTimestamp(entry) {
			result := make(Entries, len(e))
			copy(result, e)
			result[i] = entry
			return result
		}
		if existing.TimestampID.value < entry.TimestampID.value {
			result := make(Entries, len(e)+1)
			copy(result, e[:i])
			result[i] = entry
			copy(result[i+1:], e[i:])
			return result
		}
	}
	result := make(Entries, len(e)+1)
	copy(result, e)
	result[len(e)] = entry
	return result
}

// Update replaces a matching entry and returns a new collection.
func (e Entries) Update(entry Entry) Entries {
	for i, existing := range e {
		if existing.IsSameTimestamp(entry) {
			result := make(Entries, len(e))
			copy(result, e)
			result[i] = entry
			return result
		}
	}
	result := make(Entries, len(e))
	copy(result, e)
	return result
}

// Delete removes matching timestamps and returns a new collection.
func (e Entries) Delete(ts TimestampID) Entries {
	result := make(Entries, 0, len(e))
	for _, existing := range e {
		if !existing.IsSameTimestampID(ts) {
			result = append(result, existing)
		}
	}
	return result
}

// MakeUniqueTimestamp returns an unused timestamp, trying the smallest
// increment up to 999999 candidates. Invalid candidates are skipped.
func (e Entries) MakeUniqueTimestamp(ts TimestampID) (TimestampID, error) {
	used := func(candidate TimestampID) bool {
		for _, entry := range e {
			if entry.IsSameTimestampID(candidate) {
				return true
			}
		}
		return false
	}
	if !used(ts) {
		return ts, nil
	}
	base, err := strconv.ParseUint(ts.String(), 10, 64)
	if err != nil {
		return TimestampID{}, fmt.Errorf("parse timestamp ID %q: %w", ts.String(), err)
	}
	for offset := uint64(1); offset <= 999999; offset++ {
		candidateStr := fmt.Sprintf("%012d", base+offset)
		candidate, err := ParseTimestampID(candidateStr)
		if err != nil {
			continue
		}
		if !used(candidate) {
			return candidate, nil
		}
	}
	return TimestampID{}, fmt.Errorf("no unique timestamp found within search limit for %q", ts.String())
}
