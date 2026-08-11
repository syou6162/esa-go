package esa

import "strings"

// CategoryMatchesPrefix reports whether category is prefix or is below prefix.
//
// A category is below a prefix only when the boundary is a slash, so
// "notes-old" does not match "notes". The prefix itself is considered a
// match. Trailing slashes in prefix are ignored. An empty prefix matches
// nothing, including an empty category.
func CategoryMatchesPrefix(category, prefix string) bool {
	prefix = strings.TrimRight(prefix, "/")
	if prefix == "" {
		return false
	}
	return category == prefix || strings.HasPrefix(category, prefix+"/")
}
