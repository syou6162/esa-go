package scratchpad

import (
	"errors"
	"fmt"
	"strings"
)

// ValidationError reports invalid scratchpad text or title.
type ValidationError struct {
	Message string
	Issues  []string
}

// Error returns the validation message.
func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a typed validation error from issue messages.
func NewValidationError(issues []string) *ValidationError {
	copied := append([]string(nil), issues...)
	return &ValidationError{
		Message: fmt.Sprintf("バリデーションエラー: %s", strings.Join(copied, "; ")),
		Issues:  copied,
	}
}

// EntryNotFoundError reports a missing entry timestamp.
type EntryNotFoundError struct {
	TimestampID TimestampID
}

// Error describes the missing timestamp.
func (e *EntryNotFoundError) Error() string {
	return "entry not found: timestamp_id=" + e.TimestampID.String()
}

// Unwrap makes EntryNotFoundError match ErrEntryNotFound.
func (e *EntryNotFoundError) Unwrap() error {
	return ErrEntryNotFound
}

// ErrEntryNotFound identifies a missing entry.
var ErrEntryNotFound = errors.New("scratchpad: entry not found")
