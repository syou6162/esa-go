// Package scratchpad provides pure logic for timestamped Markdown entries.
//
// It defines TimestampID, Entry, and Entries together with parsing,
// serialization, ordering, collision avoidance, and text/title validation.
// ValidationError and ErrEntryNotFound provide machine-checkable error
// contracts; application-specific post, date, tag, and concurrency policy is
// outside this package.
package scratchpad
