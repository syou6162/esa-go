package esa

import (
	"errors"
	"strings"
)

// ValidationError reports invalid esa input.
type ValidationError struct {
	Message string
	Issues  []string
}

// Error returns the validation failure message.
func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a typed validation error from issue messages.
func NewValidationError(issues []string) *ValidationError {
	copied := append([]string(nil), issues...)
	return &ValidationError{
		Message: "バリデーションエラー: " + strings.Join(copied, "; "),
		Issues:  copied,
	}
}

// ErrNotFound indicates that a post lookup found no matching post.
var ErrNotFound = errors.New("esa: post not found")
