package esa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.esa.io/v1"

const (
	DefaultSearchPage    = 1
	DefaultSearchPerPage = 20
	DefaultSearchSort    = "updated"
	DefaultSearchOrder   = "desc"
	MaxSearchPerPage     = 100
)

// SearchPostsInput holds search and pagination parameters.
type SearchPostsInput struct {
	Query   string
	Page    int
	PerPage int
	Sort    string
	Order   string
}

// CreatePostInput holds parameters for creating a post.
type CreatePostInput struct {
	Name     string
	Category string
	BodyMD   string
	Message  string
	Tags     []string
	WIP      bool
}

// UpdatePostInput holds parameters for updating a post body and tags.
type UpdatePostInput struct {
	PostNumber int
	BodyMD     string
	Message    string
	Tags       []string
}

// UpdatePostBodyOnlyInput holds parameters for updating only a post body.
type UpdatePostBodyOnlyInput struct {
	PostNumber int
	BodyMD     string
	Message    string
}

// UpdatePostNameInput holds parameters for updating a post name.
type UpdatePostNameInput struct {
	PostNumber int
	Name       string
	Message    string
}

// UpdateTagsInput holds parameters for updating post tags.
type UpdateTagsInput struct {
	PostNumber int
	Tags       []string
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient replaces the HTTP client used by Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.client = httpClient
		}
	}
}

// WithBaseURL replaces the API base URL used by Client.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL != "" {
			c.baseURL = strings.TrimRight(baseURL, "/")
		}
	}
}

// PostReader groups read-only post operations.
type PostReader interface {
	GetPost(ctx context.Context, postNumber int) (*Post, error)
	SearchByCategory(ctx context.Context, category string) (*Post, error)
	SearchPosts(ctx context.Context, input SearchPostsInput) (*SearchResult, error)
}

// PostWriter groups post mutation operations.
type PostWriter interface {
	CreatePost(ctx context.Context, input CreatePostInput) (*Post, error)
	UpdatePost(ctx context.Context, input UpdatePostInput) (*Post, error)
	UpdatePostName(ctx context.Context, input UpdatePostNameInput) (*Post, error)
	UpdateTags(ctx context.Context, input UpdateTagsInput) error
}

// PostBodyUpdater updates a post body without changing its tags.
type PostBodyUpdater interface {
	UpdatePostBodyOnly(ctx context.Context, input UpdatePostBodyOnlyInput) (*Post, error)
}

// UploadImageInput contains an image stream and its metadata.
type UploadImageInput struct {
	Reader      io.Reader
	FileName    string
	Size        int64
	ContentType string
}

// ImageUploader uploads an image and returns its esa URL.
type ImageUploader interface {
	UploadImage(ctx context.Context, input UploadImageInput) (string, error)
}

// TeamNamer provides the configured team name.
type TeamNamer interface {
	TeamName() string
}

// Client is an esa.io REST API client.
type Client struct {
	teamName    string
	accessToken string
	baseURL     string
	client      *http.Client
}

// NewClient creates an esa.io API client.
func NewClient(teamName, accessToken string, opts ...Option) *Client {
	c := &Client{
		teamName:    teamName,
		accessToken: accessToken,
		baseURL:     defaultBaseURL,
		client:      &http.Client{Timeout: 60 * time.Second},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

var _ PostReader = (*Client)(nil)
var _ PostWriter = (*Client)(nil)
var _ PostBodyUpdater = (*Client)(nil)
var _ ImageUploader = (*Client)(nil)
var _ TeamNamer = (*Client)(nil)

// TeamName returns the configured team name.
func (c *Client) TeamName() string {
	return c.teamName
}

func (c *Client) authHeader() string {
	return "Bearer " + c.accessToken
}

func normalizeSearchPostsInput(in SearchPostsInput) (SearchPostsInput, error) {
	if in.Page < 0 {
		return SearchPostsInput{}, fmt.Errorf("page must not be negative")
	}
	if in.Page == 0 {
		in.Page = DefaultSearchPage
	}
	if in.PerPage < 0 {
		return SearchPostsInput{}, fmt.Errorf("per_page must not be negative")
	}
	if in.PerPage == 0 {
		in.PerPage = DefaultSearchPerPage
	}
	if in.PerPage > MaxSearchPerPage {
		return SearchPostsInput{}, fmt.Errorf("per_page must be at most %d", MaxSearchPerPage)
	}
	if in.Sort == "" {
		in.Sort = DefaultSearchSort
	}
	if in.Order == "" {
		in.Order = DefaultSearchOrder
	}
	return in, nil
}

func validateCreatePostInput(in CreatePostInput) error {
	if in.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func validatePostNumber(postNumber int) error {
	if postNumber <= 0 {
		return fmt.Errorf("post number must be positive")
	}
	return nil
}

func validateUpdatePostNameInput(in UpdatePostNameInput) error {
	if err := validatePostNumber(in.PostNumber); err != nil {
		return err
	}
	if in.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func teamPath(teamName string) string {
	return url.PathEscape(teamName)
}

func (c *Client) postsURL() string {
	return fmt.Sprintf("%s/teams/%s/posts", c.baseURL, teamPath(c.teamName))
}

func (c *Client) attachmentPoliciesURL() string {
	return fmt.Sprintf("%s/teams/%s/attachments/policies", c.baseURL, teamPath(c.teamName))
}

func (c *Client) postURL(postNumber int) string {
	return fmt.Sprintf("%s/%d", c.postsURL(), postNumber)
}

// SearchByCategory returns the first post returned by the category search.
// It wraps ErrNotFound when the search result is empty.
func (c *Client) SearchByCategory(ctx context.Context, category string) (*Post, error) {
	op := fmt.Sprintf("esa.io search by category %q", category)
	endpoint := c.postsURL()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: build request: %s", op, redactSecrets(err.Error()))
	}
	query := req.URL.Query()
	query.Set("q", fmt.Sprintf("category:\"%s\"", category))
	query.Set("per_page", "1")
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Authorization", c.authHeader())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, wrapTransportError(op, req.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, c.httpStatusError(op, req.URL, resp)
	}

	var result struct {
		TotalCount int    `json:"total_count"`
		Posts      []Post `json:"posts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("%s via %s: decode response: %w", op, safeURL(req.URL), err)
	}
	if result.TotalCount == 0 || len(result.Posts) == 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrNotFound)
	}
	return &result.Posts[0], nil
}

// SearchPosts searches posts using the supplied query and pagination options.
func (c *Client) SearchPosts(ctx context.Context, input SearchPostsInput) (*SearchResult, error) {
	input, err := normalizeSearchPostsInput(input)
	if err != nil {
		return nil, fmt.Errorf("esa.io search posts: %w", err)
	}
	op := fmt.Sprintf("esa.io search posts q=%q", input.Query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.postsURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: build request: %s", op, redactSecrets(err.Error()))
	}
	query := req.URL.Query()
	if input.Query != "" {
		query.Set("q", input.Query)
	}
	query.Set("page", fmt.Sprintf("%d", input.Page))
	query.Set("per_page", fmt.Sprintf("%d", input.PerPage))
	query.Set("sort", input.Sort)
	query.Set("order", input.Order)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Authorization", c.authHeader())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, wrapTransportError(op, req.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, c.httpStatusError(op, req.URL, resp)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("%s via %s: decode response: %w", op, safeURL(req.URL), err)
	}
	return &result, nil
}

// GetPost retrieves a post by number and wraps ErrNotFound for HTTP 404.
func (c *Client) GetPost(ctx context.Context, postNumber int) (*Post, error) {
	if err := validatePostNumber(postNumber); err != nil {
		return nil, fmt.Errorf("esa.io get post %d: invalid input: %w", postNumber, err)
	}
	op := fmt.Sprintf("esa.io get post %d", postNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.postURL(postNumber), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: build request: %s", op, redactSecrets(err.Error()))
	}
	req.Header.Set("Authorization", c.authHeader())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, wrapTransportError(op, req.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusNotFound {
			return nil, c.notFoundStatusError(op, req.URL, resp)
		}
		return nil, c.httpStatusError(op, req.URL, resp)
	}

	var post Post
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return nil, fmt.Errorf("%s via %s: decode response: %w", op, safeURL(req.URL), err)
	}
	return &post, nil
}

// CreatePost creates a post, including its body, tags, message, and WIP flag.
func (c *Client) CreatePost(ctx context.Context, in CreatePostInput) (*Post, error) {
	if err := validateCreatePostInput(in); err != nil {
		return nil, fmt.Errorf("esa.io create post: invalid input: %w", err)
	}
	payload := map[string]any{
		"post": map[string]any{
			"name":     in.Name,
			"category": in.Category,
			"body_md":  in.BodyMD,
			"wip":      in.WIP,
			"tags":     tagsOrEmpty(in.Tags),
			"message":  in.Message,
		},
	}
	op := fmt.Sprintf("esa.io create post name=%q category=%q", in.Name, in.Category)
	return c.postJSON(ctx, c.postsURL(), payload, op)
}

// UpdatePost updates a post body and tags together.
func (c *Client) UpdatePost(ctx context.Context, in UpdatePostInput) (*Post, error) {
	if err := validatePostNumber(in.PostNumber); err != nil {
		return nil, fmt.Errorf("esa.io update post %d: invalid input: %w", in.PostNumber, err)
	}
	payload := map[string]any{
		"post": map[string]any{
			"body_md": in.BodyMD,
			"tags":    tagsOrEmpty(in.Tags),
			"message": in.Message,
		},
	}
	op := fmt.Sprintf("esa.io update post %d", in.PostNumber)
	return c.patchJSON(ctx, c.postURL(in.PostNumber), payload, op)
}

// UpdatePostBodyOnly updates only the body and message; it does not send tags.
func (c *Client) UpdatePostBodyOnly(ctx context.Context, in UpdatePostBodyOnlyInput) (*Post, error) {
	if err := validatePostNumber(in.PostNumber); err != nil {
		return nil, fmt.Errorf("esa.io update post %d body only: invalid input: %w", in.PostNumber, err)
	}
	payload := map[string]any{
		"post": map[string]any{
			"body_md": in.BodyMD,
			"message": in.Message,
		},
	}
	op := fmt.Sprintf("esa.io update post %d body only", in.PostNumber)
	return c.patchJSON(ctx, c.postURL(in.PostNumber), payload, op)
}

// UpdatePostName updates a post name and sends wip=false without body or tags.
func (c *Client) UpdatePostName(ctx context.Context, in UpdatePostNameInput) (*Post, error) {
	if err := validateUpdatePostNameInput(in); err != nil {
		return nil, fmt.Errorf("esa.io update post %d name=%q: invalid input: %w", in.PostNumber, in.Name, err)
	}
	payload := map[string]any{
		"post": map[string]any{
			"name":    in.Name,
			"message": in.Message,
			"wip":     false,
		},
	}
	op := fmt.Sprintf("esa.io update post %d name=%q", in.PostNumber, in.Name)
	return c.patchJSON(ctx, c.postURL(in.PostNumber), payload, op)
}

// UpdateTags replaces the tags on an existing post.
func (c *Client) UpdateTags(ctx context.Context, in UpdateTagsInput) error {
	if err := validatePostNumber(in.PostNumber); err != nil {
		return fmt.Errorf("esa.io update tags post %d: invalid input: %w", in.PostNumber, err)
	}
	payload := map[string]any{
		"post": map[string]any{
			"tags": tagsOrEmpty(in.Tags),
		},
	}
	op := fmt.Sprintf("esa.io update tags post %d", in.PostNumber)
	_, err := c.patchJSON(ctx, c.postURL(in.PostNumber), payload, op)
	return err
}

func tagsOrEmpty(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}

func validateUploadImageInput(input UploadImageInput) error {
	if input.Reader == nil {
		return fmt.Errorf("image reader is required")
	}
	if input.FileName == "" {
		return fmt.Errorf("image file name is required")
	}
	if input.Size <= 0 {
		return fmt.Errorf("image size must be positive")
	}
	if input.ContentType == "" {
		return fmt.Errorf("image content type is required")
	}
	return nil
}

// UploadImage uploads an image through esa's upload-policy flow.
func (c *Client) UploadImage(ctx context.Context, input UploadImageInput) (string, error) {
	if err := validateUploadImageInput(input); err != nil {
		return "", err
	}
	policyPayload := map[string]any{
		"type": input.ContentType,
		"name": input.FileName,
		"size": input.Size,
	}
	policyBody, err := json.Marshal(policyPayload)
	if err != nil {
		return "", fmt.Errorf("marshal image upload policy request: %w", err)
	}

	policyOp := fmt.Sprintf("esa.io get upload policy for image %q", input.FileName)
	policyReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.attachmentPoliciesURL(), bytes.NewReader(policyBody))
	if err != nil {
		return "", fmt.Errorf("%s: build request: %s", policyOp, redactSecrets(err.Error()))
	}
	policyReq.Header.Set("Authorization", c.authHeader())
	policyReq.Header.Set("Content-Type", "application/json")

	policyResp, err := c.client.Do(policyReq)
	if err != nil {
		return "", wrapTransportError(policyOp, policyReq.URL, err)
	}
	defer policyResp.Body.Close()
	if policyResp.StatusCode < 200 || policyResp.StatusCode >= 300 {
		return "", c.httpStatusError(policyOp, policyReq.URL, policyResp)
	}

	var policyResult struct {
		Attachment struct {
			Endpoint string `json:"endpoint"`
			URL      string `json:"url"`
		} `json:"attachment"`
		Form map[string]string `json:"form"`
	}
	if err := json.NewDecoder(policyResp.Body).Decode(&policyResult); err != nil {
		return "", fmt.Errorf("%s via %s: decode response: %w", policyOp, safeURL(policyReq.URL), err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range policyResult.Form {
		if err := writer.WriteField(key, value); err != nil {
			return "", fmt.Errorf("write form field %s for image %q: %w", key, input.FileName, err)
		}
	}
	part, err := writer.CreateFormFile("file", input.FileName)
	if err != nil {
		return "", fmt.Errorf("create form file for image %q: %w", input.FileName, err)
	}
	if _, err := io.Copy(part, input.Reader); err != nil {
		return "", fmt.Errorf("write file content for image %q: %w", input.FileName, err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close multipart writer for image %q: %w", input.FileName, err)
	}

	uploadOp := fmt.Sprintf("esa.io upload image %q", input.FileName)
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPost, policyResult.Attachment.Endpoint, &body)
	if err != nil {
		return "", fmt.Errorf("%s: build request: %s", uploadOp, redactSecrets(err.Error()))
	}
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	uploadResp, err := c.client.Do(uploadReq)
	if err != nil {
		return "", wrapTransportError(uploadOp, uploadReq.URL, err)
	}
	defer uploadResp.Body.Close()
	if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusCreated && uploadResp.StatusCode != http.StatusNoContent {
		return "", c.httpStatusError(uploadOp, uploadReq.URL, uploadResp)
	}
	return policyResult.Attachment.URL, nil
}

func (c *Client) postJSON(ctx context.Context, endpoint string, payload any, op string) (*Post, error) {
	return c.doJSON(ctx, http.MethodPost, endpoint, payload, op)
}

func (c *Client) patchJSON(ctx context.Context, endpoint string, payload any, op string) (*Post, error) {
	return c.doJSON(ctx, http.MethodPatch, endpoint, payload, op)
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, payload any, op string) (*Post, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: marshal request: %w", op, err)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%s: build request: %s", op, redactSecrets(err.Error()))
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, wrapTransportError(op, req.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, c.httpStatusError(op, req.URL, resp)
	}

	var post Post
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return nil, fmt.Errorf("%s via %s: decode response: %w", op, safeURL(req.URL), err)
	}
	return &post, nil
}

func (c *Client) httpStatusError(op string, endpoint *url.URL, resp *http.Response) error {
	body, readErr := readErrorBody(resp.Body)
	safeEndpoint := safeURL(endpoint)
	if readErr != nil {
		return fmt.Errorf("%s via %s: status %d: %w", op, safeEndpoint, resp.StatusCode, readErr)
	}
	return fmt.Errorf("%s via %s: status %d, body: %s", op, safeEndpoint, resp.StatusCode, redactSecrets(string(body)))
}

func (c *Client) notFoundStatusError(op string, endpoint *url.URL, resp *http.Response) error {
	body, readErr := readErrorBody(resp.Body)
	safeEndpoint := safeURL(endpoint)
	if readErr != nil {
		return fmt.Errorf("%s via %s: status %d: %w: %w", op, safeEndpoint, resp.StatusCode, ErrNotFound, readErr)
	}
	return fmt.Errorf("%s via %s: status %d, body: %s: %w", op, safeEndpoint, resp.StatusCode, redactSecrets(string(body)), ErrNotFound)
}

func readErrorBody(reader io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(reader, 4096))
}

type transportError struct {
	message string
	cause   error
}

func (e *transportError) Error() string {
	return e.message
}

func (e *transportError) Unwrap() error {
	return e.cause
}

func wrapTransportError(op string, endpoint *url.URL, cause error) error {
	return &transportError{
		message: fmt.Sprintf("%s via %s: %s", op, safeURL(endpoint), redactSecrets(cause.Error())),
		cause:   cause,
	}
}

func safeURL(endpoint *url.URL) string {
	if endpoint == nil || endpoint.Scheme == "" || endpoint.Host == "" {
		return "<unparseable-url>"
	}
	copyURL := *endpoint
	copyURL.RawQuery = ""
	copyURL.Fragment = ""
	copyURL.User = nil
	return copyURL.String()
}

var (
	bearerTokenPattern = regexp.MustCompile(`(?i)Bearer\s+[^\s'"]+`)
	authHeaderPattern  = regexp.MustCompile(`(?i)Authorization:\s*[^\n]+`)
	secretQueryPattern = regexp.MustCompile(`(?i)([?&](?:token|api_key|apikey|access_token|refresh_token|client_secret|sig|signature|password|secret)=)[^&\s'"]+`)
	urlPattern         = regexp.MustCompile(`https?://[^\s"'()\[\]{},]+`)
)

// URL redaction removes complete query strings; the fragment pattern below
// also protects scheme-less query fragments that are not recognized as URLs.
func redactSecrets(message string) string {
	message = bearerTokenPattern.ReplaceAllString(message, "Bearer [REDACTED]")
	message = authHeaderPattern.ReplaceAllString(message, "Authorization: [REDACTED]")
	message = urlPattern.ReplaceAllStringFunc(message, func(raw string) string {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "<unparseable-url>"
		}
		return safeURL(parsed)
	})
	return secretQueryPattern.ReplaceAllString(message, "${1}[REDACTED]")
}

// DetectImageContentType returns the MIME type associated with an image file
// name. Its result can be used to populate UploadImageInput.ContentType.
func DetectImageContentType(fileName string) string {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
