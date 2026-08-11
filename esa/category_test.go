package esa

import "testing"

func TestCategoryMatchesPrefix(t *testing.T) {
	tests := []struct {
		name     string
		category string
		prefix   string
		want     bool
	}{
		{
			name:     "exact match",
			category: "notes",
			prefix:   "notes",
			want:     true,
		},
		{
			name:     "descendant",
			category: "notes/2026/08/11",
			prefix:   "notes",
			want:     true,
		},
		{
			name:     "similar category with suffix",
			category: "notes-old/2026/08/11",
			prefix:   "notes",
			want:     false,
		},
		{
			name:     "shorter category",
			category: "note",
			prefix:   "notes",
			want:     false,
		},
		{
			name:     "empty category",
			category: "",
			prefix:   "notes",
			want:     false,
		},
		{
			name:     "empty prefix",
			category: "notes",
			prefix:   "",
			want:     false,
		},
		{
			name:     "both empty",
			category: "",
			prefix:   "",
			want:     false,
		},
		{
			name:     "trailing slash in prefix",
			category: "notes/2026/08/11",
			prefix:   "notes/",
			want:     true,
		},
		{
			name:     "trailing slash exact category",
			category: "notes",
			prefix:   "notes/",
			want:     true,
		},
		{
			name:     "japanese category",
			category: "メモ/2026/08/11",
			prefix:   "メモ",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CategoryMatchesPrefix(tt.category, tt.prefix); got != tt.want {
				t.Fatalf("CategoryMatchesPrefix(%q, %q) = %v, want %v", tt.category, tt.prefix, got, tt.want)
			}
		})
	}
}
