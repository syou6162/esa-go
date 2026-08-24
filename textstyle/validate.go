package textstyle

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	boldRE    = regexp.MustCompile(`\*\*[^*]+\*\*`)
	headingRE = regexp.MustCompile(`(?m)^#{1,6}\s`)
)

// Validate reports the style issues found in text.
//
// It returns nil when text has no issue. The issues are Japanese messages
// meant for the author of the text, and the order is stable so callers can
// join them into a single error message.
func Validate(text string) []string {
	var issues []string
	if containsSeparator(text) {
		issues = append(issues, "区切り線(---)が含まれています")
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
	if hasLeadingMiddleDot(text) {
		issues = append(issues, "行頭の中黒(・)は使用できません。マークダウンリスト記法(- )を使ってください")
	}
	return issues
}

// containsSeparator reports whether text has a line that reads as a Markdown
// horizontal rule. A line that also has a table cell delimiter is treated as a
// table separator row instead.
func containsSeparator(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "---") && !strings.Contains(line, "|") {
			return true
		}
	}
	return false
}

// hasLeadingMiddleDot reports whether any line of text starts with a middle
// dot. A middle dot inside a line is a normal Japanese delimiter.
func hasLeadingMiddleDot(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimLeftFunc(line, unicode.IsSpace), "・") {
			return true
		}
	}
	return false
}
