package esa

import "strings"

// CategoryMatchesPrefix reports whether category is prefix or is below prefix.
//
// A category is below a prefix only when the boundary is a slash, so
// "notes-old" does not match "notes". The prefix itself is considered a
// match. Callers must provide a category path without a trailing slash as
// prefix. An empty prefix matches nothing, including an empty category.
func CategoryMatchesPrefix(category, prefix string) bool {
	if prefix == "" {
		return false
	}
	return category == prefix || strings.HasPrefix(category, prefix+"/")
}
