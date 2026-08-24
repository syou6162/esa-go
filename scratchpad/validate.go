package scratchpad

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/syou6162/esa-go/textstyle"
)

var (
	leadingTimeRE = regexp.MustCompile(`^(?:[01]?\d|2[0-3]):[0-5]\d`)
	listMarkerRE  = regexp.MustCompile(`^[-*]\s`)
)

func validatePostText(text string) []string {
	issues := textstyle.Validate(text)
	if leadingTimeRE.MatchString(text) {
		issues = append(issues, "先頭（行頭）に時刻が含まれています。時刻はシステムが自動挿入するため不要です")
	}
	if listMarkerRE.MatchString(text) {
		issues = append(issues, "先頭（行頭）にマークダウンリスト記法(- / *)が使用されています。タイムスタンプ挿入でスタイルが崩れるため使用できません")
	}
	return issues
}

func validateScratchpadTitle(title string) []string {
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
