package esa

// PostRef identifies an esa.io post by its number and URL.
type PostRef struct {
	Number int    `json:"post_number"`
	URL    string `json:"url"`
}

// Ref extracts the identifying fields from a post response.
func (p *Post) Ref() PostRef {
	if p == nil {
		return PostRef{}
	}
	return PostRef{
		Number: p.Number,
		URL:    p.URL,
	}
}
