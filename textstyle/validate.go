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

// Validate reports the style issues that apply to any esa.io article text.
//
// It covers the full-width punctuation rules and the leading middle dot, which
// are unwanted regardless of the article format. Rules that only some formats
// want, such as Markdown heading and bold syntax, have their own functions so
// callers can combine the ones they need.
//
// Every function in this package returns nil when text has no issue, and
// returns Japanese messages meant for the author of the text.
func Validate(text string) []string {
	var issues []string
	issues = append(issues, ValidateFullWidthColon(text)...)
	issues = append(issues, ValidateFullWidthParentheses(text)...)
	issues = append(issues, ValidateLeadingMiddleDot(text)...)
	return issues
}

// ValidateHeading reports Markdown heading syntax.
func ValidateHeading(text string) []string {
	if headingRE.MatchString(text) {
		return []string{"Markdown見出し(#)が使用されています"}
	}
	return nil
}

// ValidateBold reports Markdown bold syntax.
func ValidateBold(text string) []string {
	if boldRE.MatchString(text) {
		return []string{"ボールド体(**テキスト**)が使用されています"}
	}
	return nil
}

// ValidateFullWidthColon reports full-width colons.
func ValidateFullWidthColon(text string) []string {
	if strings.ContainsRune(text, '\uFF1A') {
		return []string{"全角コロン(：)が使用されています。半角コロン(:)を使ってください"}
	}
	return nil
}

// ValidateFullWidthParentheses reports full-width parentheses.
func ValidateFullWidthParentheses(text string) []string {
	if strings.ContainsRune(text, '\uFF08') || strings.ContainsRune(text, '\uFF09') {
		return []string{"全角括弧が使用されています。半角括弧を使ってください"}
	}
	return nil
}

// ValidateLeadingMiddleDot reports a middle dot at the beginning of a line. A
// middle dot inside a line is a normal Japanese delimiter.
func ValidateLeadingMiddleDot(text string) []string {
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimLeftFunc(line, unicode.IsSpace), "・") {
			return []string{"行頭の中黒(・)は使用できません。マークダウンリスト記法(- )を使ってください"}
		}
	}
	return nil
}
