package esa

import "errors"

// ValidationError reports invalid esa input.
type ValidationError struct {
	Message string
}

// Error returns the validation failure message.
func (e *ValidationError) Error() string {
	return "バリデーションエラー: " + e.Message
}

// NewValidationError creates a typed validation error from a reason.
func NewValidationError(reason string) *ValidationError {
	return &ValidationError{Message: reason}
}

// ErrNotFound indicates that a post lookup found no matching post.
var ErrNotFound = errors.New("esa: post not found")
