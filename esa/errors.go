package esa

import "errors"

// ErrNotFound indicates that a post lookup found no matching post.
var ErrNotFound = errors.New("esa: post not found")
