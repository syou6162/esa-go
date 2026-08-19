package esa

// Revision represents a single revision of an esa.io post.
//
// The Revision API is a beta feature and is not documented in the official esa
// API reference (https://docs.esa.io/posts/102). The primary sources are the
// esa-ruby README (https://github.com/esaio/esa-ruby) and the API definition in
// https://github.com/esaio/esa-mcp-server/pull/338.
type Revision struct {
	Number    int      `json:"number"`
	Name      string   `json:"name"`
	FullName  string   `json:"full_name"`
	Category  string   `json:"category"`
	Tags      []string `json:"tags"`
	BodyMD    string   `json:"body_md"`
	BodyHTML  string   `json:"body_html"`
	Diff      string   `json:"diff"`
	Message   string   `json:"message"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	WIP       bool     `json:"wip"`
	CreatedBy User     `json:"created_by"`
	URL       string   `json:"url"`
}

// RevisionList is the response returned by the revision list endpoint.
// Revisions are ordered by revision number descending (newest first).
type RevisionList struct {
	Revisions  []Revision `json:"revisions"`
	PrevPage   *int       `json:"prev_page"`
	NextPage   *int       `json:"next_page"`
	TotalCount int        `json:"total_count"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
	MaxPerPage int        `json:"max_per_page"`
}

// RevisionDiff is the response returned by the revision compare endpoint.
type RevisionDiff struct {
	FromRevisionNumber int    `json:"from_revision_number"`
	ToRevisionNumber   int    `json:"to_revision_number"`
	DiffHTML           string `json:"diff_html"`
	DiffText           string `json:"diff_text"`
}
