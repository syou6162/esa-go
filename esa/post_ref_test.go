package esa

import "testing"

func TestPostRef(t *testing.T) {
	post := &Post{Number: 42, URL: "https://example.invalid/posts/42"}
	if got := post.Ref(); got != (PostRef{Number: 42, URL: post.URL}) {
		t.Fatalf("Ref() = %#v", got)
	}

	var nilPost *Post
	if got := nilPost.Ref(); got != (PostRef{}) {
		t.Fatalf("nil Ref() = %#v", got)
	}
}
