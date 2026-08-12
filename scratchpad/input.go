package scratchpad

import "strings"

// PostText is validated text for a scratchpad entry.
//
// PostText values can be created only by ParsePostText. The zero value is
// empty and String returns an empty string for it.
type PostText struct {
	value string
}

// String returns the raw post text.
func (t PostText) String() string {
	return t.value
}

// ParsePostText validates and parses scratchpad entry text.
func ParsePostText(raw string) (PostText, error) {
	if raw == "" {
		return PostText{}, NewValidationError([]string{"本文が空です"})
	}
	if issues := validatePostText(raw); len(issues) > 0 {
		return PostText{}, NewValidationError(issues)
	}
	return PostText{value: raw}, nil
}

// PostTitle is a validated scratchpad post title.
//
// PostTitle values can be created only by ParsePostTitle. The zero value is
// empty and String returns an empty string for it.
type PostTitle struct {
	value string
}

// String returns the raw post title.
func (t PostTitle) String() string {
	return t.value
}

// ParsePostTitle validates and parses a scratchpad post title.
func ParsePostTitle(raw string) (PostTitle, error) {
	if raw == "" {
		return PostTitle{}, NewValidationError([]string{"タイトルが空です"})
	}
	if strings.TrimSpace(raw) != raw {
		return PostTitle{}, NewValidationError([]string{"タイトルの前後に空白は使用できません"})
	}
	if issues := validateScratchpadTitle(raw); len(issues) > 0 {
		return PostTitle{}, NewValidationError(issues)
	}
	return PostTitle{value: raw}, nil
}
