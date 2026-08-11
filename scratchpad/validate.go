package scratchpad

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	boldRE        = regexp.MustCompile(`\*\*[^*]+\*\*`)
	headingRE     = regexp.MustCompile(`(?m)^#{1,6}\s`)
	leadingTimeRE = regexp.MustCompile(`(?m)^(?:[01]?\d|2[0-3]):[0-5]\d`)
	listMarkerRE  = regexp.MustCompile(`(?m)^[-*]\s`)
)

// ValidatePostText returns all validation issues found in post text.
func ValidatePostText(text string) []string {
	var issues []string
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "---") && !strings.Contains(line, "|") {
			issues = append(issues, "区切り線(---)が含まれています")
			break
		}
	}
	if boldRE.MatchString(text) {
		issues = append(issues, "ボールド体(**テキスト**)が使用されています")
	}
	if headingRE.MatchString(text) {
		issues = append(issues, "Markdown見出し(#)が使用されています")
	}
	if strings.ContainsRune(text, '\uFF1A') {
		issues = append(issues, "全角コロン(：)が使用されています。半角コロン(:)を使ってください")
	}
	if strings.ContainsRune(text, '\uFF08') || strings.ContainsRune(text, '\uFF09') {
		issues = append(issues, "全角括弧が使用されています。半角括弧を使ってください")
	}
	if leadingTimeRE.MatchString(text) {
		issues = append(issues, "先頭（行頭）に時刻が含まれています。時刻はシステムが自動挿入するため不要です")
	}
	if listMarkerRE.MatchString(text) {
		issues = append(issues, "先頭（行頭）にマークダウンリスト記法(- / *)が使用されています。タイムスタンプ挿入でスタイルが崩れるため使用できません")
	}
	return issues
}

// CheckTextValidation returns a ValidationError when text is invalid.
func CheckTextValidation(text string) error {
	issues := ValidatePostText(text)
	if len(issues) == 0 {
		return nil
	}
	return NewValidationError(issues)
}

// ValidateScratchpadTitle returns all validation issues found in a title.
func ValidateScratchpadTitle(title string) []string {
	var issues []string
	if strings.Contains(title, "/") {
		issues = append(issues, "スラッシュ(/)はカテゴリ区切りとして解釈されるため使用できません")
	}
	if strings.ContainsRune(title, '\u3000') {
		issues = append(issues, "全角スペースは使用できません")
	}
	if title != "" {
		first, _ := utf8.DecodeRuneInString(title)
		if !isAllowedLeadingChar(first) {
			issues = append(issues, "先頭に記号は使用できません")
		}
	}
	if strings.Contains(title, ".") || strings.ContainsRune(title, '\u3002') {
		issues = append(issues, "ピリオド(.・。)は使用できません。区切りには読点(、)を使ってください")
	}
	if strings.ContainsAny(title, "!?") || strings.ContainsRune(title, '\uFF01') || strings.ContainsRune(title, '\uFF1F') {
		issues = append(issues, "感嘆符・疑問符(!?！？)は使用できません")
	}
	if strings.ContainsAny(title, "\n\r") {
		issues = append(issues, "改行文字は使用できません")
	}
	return issues
}

// CheckTitleValidation returns a ValidationError when the title is invalid.
func CheckTitleValidation(title string) error {
	issues := ValidateScratchpadTitle(title)
	if len(issues) == 0 {
		return nil
	}
	return NewValidationError(issues)
}

func isAllowedLeadingChar(c rune) bool {
	if unicode.IsLetter(c) || unicode.IsDigit(c) {
		return true
	}
	switch {
	case c >= '\u3040' && c <= '\u309F':
		return true
	case c >= '\u30A0' && c <= '\u30FF':
		return true
	case c >= '\u4E00' && c <= '\u9FFF':
		return true
	case c >= '\uFF65' && c <= '\uFF9F':
		return true
	default:
		return false
	}
}
