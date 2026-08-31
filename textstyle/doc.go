// Package textstyle provides reusable Markdown and Japanese punctuation style
// rules for esa.io article text.
//
// Validate holds the rules that apply to any article format: full-width colons
// and parentheses, and a middle dot at the beginning of a line. Rules that only
// some formats want, such as Markdown heading and bold syntax, have their own
// exported function so callers can combine the ones their format needs. Rules
// that belong to a specific article format, such as the entry separator and the
// leading time of a scratchpad entry, stay in the package that owns that format.
//
// Every function reports each violation it finds as a Japanese issue message so
// callers can surface all of them at once, and combine them with their own
// format-specific issues before building a validation error.
package textstyle
