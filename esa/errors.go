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

// ErrNotFound indicates that a post lookup found no matching post, or that a
// revision lookup found no matching post or revision: esa.io answers all of
// these cases with HTTP 404.
var ErrNotFound = errors.New("esa: post or revision not found")

// ErrRollbackToLatestRevision indicates a rollback whose target revision is
// already the latest revision of the post. esa.io rejects it with HTTP 400.
var ErrRollbackToLatestRevision = errors.New("esa: rollback to the latest revision is not allowed")
