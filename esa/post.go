package esa

// Post represents an esa.io post.
type Post struct {
	Number          int      `json:"number"`
	Name            string   `json:"name"`
	FullName        string   `json:"full_name"`
	Category        string   `json:"category"`
	BodyMD          string   `json:"body_md"`
	BodyHTML        string   `json:"body_html"`
	Tags            []string `json:"tags"`
	WIP             bool     `json:"wip"`
	Message         string   `json:"message"`
	Kind            string   `json:"kind"`
	URL             string   `json:"url"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	RevisionNumber  int      `json:"revision_number"`
	CreatedBy       User     `json:"created_by"`
	UpdatedBy       User     `json:"updated_by"`
	Star            bool     `json:"star"`
	Watch           bool     `json:"watch"`
	CommentsCount   int      `json:"comments_count"`
	TasksCount      int      `json:"tasks_count"`
	DoneTasksCount  int      `json:"done_tasks_count"`
	StargazersCount int      `json:"stargazers_count"`
	WatchersCount   int      `json:"watchers_count"`
	BacklinksCount  int      `json:"backlinks_count"`
	Overlapped      bool     `json:"overlapped,omitempty"`
}

// SearchResult is the response returned by the posts search endpoint.
type SearchResult struct {
	Posts      []Post `json:"posts"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
}

// User represents an esa.io user embedded in a post response.
type User struct {
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
	Icon       string `json:"icon"`
}
