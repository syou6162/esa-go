package esa

import (
	"fmt"
	"strings"
)

// Tag is a validated esa post tag.
//
// Tag values can be created only by ParseTag. The zero value is empty and
// String returns an empty string for it.
type Tag struct {
	value string
}

// String returns the raw tag name.
func (t Tag) String() string {
	return t.value
}

// ParseTag validates and parses an esa post tag.
func ParseTag(raw string) (Tag, error) {
	if raw == "" {
		return Tag{}, NewValidationError([]string{"タグが空です"})
	}
	if strings.TrimSpace(raw) != raw {
		return Tag{}, NewValidationError([]string{"タグの前後に空白は使用できません"})
	}
	return Tag{value: raw}, nil
}

// PostNumber is a validated positive esa post number.
//
// PostNumber values can be created only by ParsePostNumber. The zero value is
// invalid and Int returns zero for it.
type PostNumber struct {
	value int
}

// Int returns the raw post number.
func (n PostNumber) Int() int {
	return n.value
}

// ParsePostNumber validates and parses a positive esa post number.
func ParsePostNumber(raw int) (PostNumber, error) {
	if raw <= 0 {
		return PostNumber{}, NewValidationError(
			[]string{fmt.Sprintf("記事番号は正の整数である必要があります (post_number=%d)", raw)},
		)
	}
	return PostNumber{value: raw}, nil
}

// RevisionNumber is a validated positive esa revision number.
//
// RevisionNumber values can be created only by ParseRevisionNumber. The zero
// value is invalid and Int returns zero for it.
type RevisionNumber struct {
	value int
}

// Int returns the raw revision number.
func (n RevisionNumber) Int() int {
	return n.value
}

// ParseRevisionNumber validates and parses a positive esa revision number.
func ParseRevisionNumber(raw int) (RevisionNumber, error) {
	if raw <= 0 {
		return RevisionNumber{}, NewValidationError(
			[]string{fmt.Sprintf("リビジョン番号は正の整数である必要があります (revision_number=%d)", raw)},
		)
	}
	return RevisionNumber{value: raw}, nil
}
