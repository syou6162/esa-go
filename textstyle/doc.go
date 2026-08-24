// Package textstyle provides reusable Markdown and Japanese punctuation style
// rules for esa.io article text.
//
// The rules are independent of any particular article format: they reject
// Markdown separators, headings, and bold syntax, plus full-width colons and
// parentheses. Rules that belong to a specific article format, such as the
// leading time and list marker of a scratchpad entry, stay in the package that
// owns that format.
//
// Validate reports every violation it finds as a Japanese issue message so
// callers can surface all of them at once, and combine them with their own
// format-specific issues before building a validation error.
package textstyle
