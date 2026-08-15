package esa

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSearchByCategory(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		var gotMethod, gotQuery, gotAuth, gotPath, gotPerPage string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotQuery = r.URL.Query().Get("q")
			gotPerPage = r.URL.Query().Get("per_page")
			gotAuth = r.Header.Get("Authorization")
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"total_count":1,"posts":[{"number":42,"body_md":"hello","tags":["tag-a"]}]}`))
		}))
		defer srv.Close()

		client := NewClient("example-team", "dummy-token", WithBaseURL(srv.URL+"/v1"), WithHTTPClient(srv.Client()))
		post, err := client.SearchByCategory(context.Background(), "example/category/2025/05/03")
		if err != nil {
			t.Fatalf("SearchByCategory: %v", err)
		}
		if gotMethod != http.MethodGet {
			t.Errorf("method = %q, want GET", gotMethod)
		}
		if gotQuery != `category:"example/category/2025/05/03"` {
			t.Errorf("q = %q", gotQuery)
		}
		if gotPerPage != "1" {
			t.Errorf("per_page = %q, want 1", gotPerPage)
		}
		if gotAuth != "Bearer dummy-token" {
			t.Errorf("Authorization = %q", gotAuth)
		}
		if gotPath != "/v1/teams/example-team/posts" {
			t.Errorf("path = %q", gotPath)
		}
		if post == nil || post.Number != 42 {
			t.Fatalf("post = %#v", post)
		}
	})

	t.Run("empty result is not found", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"total_count":0,"posts":[]}`))
		}))
		defer srv.Close()

		client := testClient(srv)
		post, err := client.SearchByCategory(context.Background(), "example/category")
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
		if post != nil {
			t.Fatalf("post = %#v, want nil", post)
		}
	})

	t.Run("non-2xx", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"error":"unauthorized"}`)
		}))
		defer srv.Close()

		_, err := testClient(srv).SearchByCategory(context.Background(), "example/category")
		if err == nil || !strings.Contains(err.Error(), "status 401") {
			t.Fatalf("err = %v, want status 401", err)
		}
	})

	t.Run("decode error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, "{")
		}))
		defer srv.Close()

		_, err := testClient(srv).SearchByCategory(context.Background(), "example/category")
		if err == nil || !strings.Contains(err.Error(), "decode response") {
			t.Fatalf("err = %v, want decode response", err)
		}
	})
}

func TestSearchPosts(t *testing.T) {
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"posts":[{"number":7}],"total_count":1,"page":1,"per_page":20}`))
	}))
	defer srv.Close()

	result, err := testClient(srv).SearchPosts(context.Background(), SearchPostsInput{Query: "user:example"})
	if err != nil {
		t.Fatalf("SearchPosts: %v", err)
	}
	if result.TotalCount != 1 || result.Posts[0].Number != 7 {
		t.Fatalf("result = %#v", result)
	}
	for key, want := range map[string]string{
		"q": "user:example", "page": "1", "per_page": "20", "sort": "updated", "order": "desc",
	} {
		if got := gotQuery.Get(key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestSearchPostsValidation(t *testing.T) {
	client := NewClient("example-team", "dummy-token")
	tests := []SearchPostsInput{
		{Page: -1},
		{PerPage: -1},
		{PerPage: MaxSearchPerPage + 1},
	}
	for _, input := range tests {
		if _, err := client.SearchPosts(context.Background(), input); err == nil {
			t.Errorf("SearchPosts(%+v): want error", input)
		}
	}
}

func TestGetPost(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotAuth, gotMethod, gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			gotMethod = r.Method
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"number":99,"name":"example title"}`))
		}))
		defer srv.Close()

		post, err := testClient(srv).GetPost(context.Background(), 99)
		if err != nil {
			t.Fatalf("GetPost: %v", err)
		}
		if gotMethod != http.MethodGet || gotPath != "/teams/example-team/posts/99" {
			t.Errorf("request = %s %s", gotMethod, gotPath)
		}
		if gotAuth != "Bearer dummy-token" {
			t.Errorf("Authorization = %q", gotAuth)
		}
		if post.Number != 99 || post.Name != "example title" {
			t.Fatalf("post = %#v", post)
		}
	})

	t.Run("invalid number", func(t *testing.T) {
		client := NewClient("example-team", "dummy-token")
		for _, number := range []int{0, -1} {
			if _, err := client.GetPost(context.Background(), number); err == nil {
				t.Errorf("GetPost(%d): want error", number)
			}
		}
	})

	t.Run("404 is not found", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
		}))
		defer srv.Close()

		_, err := testClient(srv).GetPost(context.Background(), 404)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("decode error includes operation", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, "{")
		}))
		defer srv.Close()

		_, err := testClient(srv).GetPost(context.Background(), 7)
		if err == nil || !strings.Contains(err.Error(), "get post 7") || !strings.Contains(err.Error(), "decode response") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("non-404 non-2xx", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `{"error":"internal"}`)
		}))
		defer srv.Close()

		_, err := testClient(srv).GetPost(context.Background(), 500)
		if err == nil || !strings.Contains(err.Error(), "status 500") {
			t.Fatalf("err = %v, want status 500", err)
		}
		if errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want not ErrNotFound", err)
		}
	})
}

func TestHTTPErrorDetailsAndRedaction(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"https://upload.example.invalid/file?signature=body-secret","auth":"Bearer body-secret"}`))
	}))
	defer srv.Close()

	_, err := testClient(srv).SearchPosts(context.Background(), SearchPostsInput{})
	if err == nil {
		t.Fatal("SearchPosts: want error")
	}
	message := err.Error()
	for _, want := range []string{"search posts", "status 400", "body"} {
		if !strings.Contains(message, want) {
			t.Errorf("error %q does not contain %q", message, want)
		}
	}
	for _, secret := range []string{"body-secret", "signature="} {
		if strings.Contains(message, secret) {
			t.Errorf("error %q contains secret %q", message, secret)
		}
	}
}

func TestTransportErrorRedactsPresignedURLAndPreservesCause(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/teams/example-team/attachments/policies" {
			_, _ = io.WriteString(w, `{"attachment":{"endpoint":"`+srv.URL+`/upload?signature=upload-secret","url":"https://cdn.example.invalid/image"},"form":{}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	filePath := filepath.Join(t.TempDir(), "image.png")
	if err := os.WriteFile(filePath, []byte("image-data"), 0600); err != nil {
		t.Fatal(err)
	}
	client := NewClient("example-team", "dummy-token",
		WithBaseURL(srv.URL),
		WithHTTPClient(&http.Client{Transport: presignedErrorTransport{base: srv.Client().Transport}}),
	)
	_, err := client.UploadFile(context.Background(), filePath)
	if err == nil {
		t.Fatal("UploadFile: want transport error")
	}
	if strings.Contains(err.Error(), "upload-secret") || strings.Contains(err.Error(), "signature=") {
		t.Fatalf("error leaked presigned URL: %v", err)
	}
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		t.Fatalf("error = %T, want wrapped *url.Error", err)
	}
}

func TestRedactSecretsSchemeLessQuery(t *testing.T) {
	message := redactSecrets("request failed ?signature=fragment-secret")
	if strings.Contains(message, "fragment-secret") || !strings.Contains(message, "?signature=[REDACTED]") {
		t.Fatalf("redacted message = %q", message)
	}
}

func TestHTTPErrorBodyLimit(t *testing.T) {
	largeBody := strings.Repeat("x", 8192)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, largeBody)
	}))
	defer srv.Close()

	_, err := testClient(srv).SearchPosts(context.Background(), SearchPostsInput{})
	if err == nil {
		t.Fatal("SearchPosts: want error")
	}
	message := err.Error()
	index := strings.Index(message, "body: ")
	if index < 0 {
		t.Fatalf("error = %q, want body", message)
	}
	body, ok := strings.CutPrefix(message[index:], "body: ")
	if !ok || len(body) != 4096 {
		t.Fatalf("body length = %d, want 4096", len(body))
	}
}

func TestResponseBodyReadFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient("example-team", "dummy-token", WithBaseURL(srv.URL), WithHTTPClient(&http.Client{
		Transport: failingBodyTransport{base: srv.Client().Transport},
	}))
	_, err := client.SearchPosts(context.Background(), SearchPostsInput{})
	if err == nil || !strings.Contains(err.Error(), "read failed") || !strings.Contains(err.Error(), "status 500") {
		t.Fatalf("err = %v", err)
	}
}

func TestWritePayloads(t *testing.T) {
	t.Run("create normalizes nil tags", func(t *testing.T) {
		var payload map[string]any
		srv := httptest.NewServer(jsonCaptureHandler(&payload, `{"number":1}`))
		defer srv.Close()

		_, err := testClient(srv).CreatePost(context.Background(), CreatePostInput{
			Name: "example title", Category: "example/category", BodyMD: "body", Message: "create",
		})
		if err != nil {
			t.Fatalf("CreatePost: %v", err)
		}
		postPayload := payload["post"].(map[string]any)
		tags, ok := postPayload["tags"].([]any)
		if !ok || len(tags) != 0 {
			t.Fatalf("tags = %#v, want empty array", postPayload["tags"])
		}
	})

	t.Run("create includes expected fields only", func(t *testing.T) {
		var payload map[string]any
		srv := httptest.NewServer(jsonCaptureHandler(&payload, `{"number":1}`))
		defer srv.Close()

		_, err := testClient(srv).CreatePost(context.Background(), CreatePostInput{
			Name: "example title", Category: "example/category", BodyMD: "body",
			Message: "create", Tags: []string{"tag-a"}, WIP: true,
		})
		if err != nil {
			t.Fatalf("CreatePost: %v", err)
		}
		postPayload := payload["post"].(map[string]any)
		for key, want := range map[string]any{
			"name": "example title", "category": "example/category",
			"body_md": "body", "message": "create", "wip": true,
		} {
			if postPayload[key] != want {
				t.Errorf("%s = %#v, want %#v", key, postPayload[key], want)
			}
		}
		tags, ok := postPayload["tags"].([]any)
		if !ok || len(tags) != 1 || tags[0] != "tag-a" {
			t.Errorf("tags = %#v, want [tag-a]", postPayload["tags"])
		}
		for _, forbidden := range []string{"number", "url", "body_md_extra"} {
			if _, ok := postPayload[forbidden]; ok {
				t.Errorf("payload = %#v, want no %s", postPayload, forbidden)
			}
		}
	})

	t.Run("update and body only", func(t *testing.T) {
		var payloads []map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			payloads = append(payloads, payload)
			_, _ = w.Write([]byte(`{"number":2}`))
		}))
		defer srv.Close()

		client := testClient(srv)
		if _, err := client.UpdatePost(context.Background(), UpdatePostInput{
			PostNumber: 2, BodyMD: "body", Message: "update", Tags: nil,
		}); err != nil {
			t.Fatalf("UpdatePost: %v", err)
		}
		if _, err := client.UpdatePostBodyOnly(context.Background(), UpdatePostBodyOnlyInput{
			PostNumber: 2, BodyMD: "new body", Message: "body only",
		}); err != nil {
			t.Fatalf("UpdatePostBodyOnly: %v", err)
		}
		full := payloads[0]["post"].(map[string]any)
		if _, ok := full["tags"].([]any); !ok {
			t.Fatalf("tags = %#v, want array", full["tags"])
		}
		if full["body_md"] != "body" || full["message"] != "update" {
			t.Fatalf("payload = %#v, want body and message", full)
		}
		for _, forbidden := range []string{"name", "category", "wip"} {
			if _, ok := full[forbidden]; ok {
				t.Fatalf("payload = %#v, want no %s", full, forbidden)
			}
		}
		bodyOnly := payloads[1]["post"].(map[string]any)
		if _, ok := bodyOnly["tags"]; ok {
			t.Fatalf("body-only payload = %#v, want no tags", bodyOnly)
		}
		if bodyOnly["body_md"] != "new body" || bodyOnly["message"] != "body only" {
			t.Fatalf("payload = %#v, want body and message", bodyOnly)
		}
		for _, forbidden := range []string{"name", "category", "wip"} {
			if _, ok := bodyOnly[forbidden]; ok {
				t.Fatalf("payload = %#v, want no %s", bodyOnly, forbidden)
			}
		}
	})

	t.Run("title update omits body and tags", func(t *testing.T) {
		var payload map[string]any
		srv := httptest.NewServer(jsonCaptureHandler(&payload, `{"number":3}`))
		defer srv.Close()

		_, err := testClient(srv).UpdatePostName(context.Background(), UpdatePostNameInput{
			PostNumber: 3, Name: "new title", Message: "rename",
		})
		if err != nil {
			t.Fatalf("UpdatePostName: %v", err)
		}
		postPayload := payload["post"].(map[string]any)
		if _, ok := postPayload["body_md"]; ok {
			t.Fatalf("payload = %#v, want no body_md", postPayload)
		}
		if _, ok := postPayload["tags"]; ok {
			t.Fatalf("payload = %#v, want no tags", postPayload)
		}
		if postPayload["name"] != "new title" || postPayload["message"] != "rename" {
			t.Fatalf("payload = %#v, want name and message", postPayload)
		}
		if postPayload["wip"] != false {
			t.Fatalf("payload = %#v, want wip=false", postPayload)
		}
		if _, ok := postPayload["category"]; ok {
			t.Fatalf("payload = %#v, want no category", postPayload)
		}
	})
}

func TestCreatePostWIPPropagation(t *testing.T) {
	var payload map[string]any
	srv := httptest.NewServer(jsonCaptureHandler(&payload, `{"number":1}`))
	defer srv.Close()

	client := testClient(srv)
	for _, want := range []bool{true, false} {
		_, err := client.CreatePost(context.Background(), CreatePostInput{Name: "title", WIP: want})
		if err != nil {
			t.Fatalf("CreatePost(WIP=%t): %v", want, err)
		}
		postPayload := payload["post"].(map[string]any)
		if got := postPayload["wip"]; got != want {
			t.Errorf("wip = %v, want %t", got, want)
		}
	}
}

func TestWriteRequestMetadata(t *testing.T) {
	var gotMethod, gotPath, gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		_, _ = io.WriteString(w, `{"number":1}`)
	}))
	defer srv.Close()

	_, err := testClient(srv).CreatePost(context.Background(), CreatePostInput{Name: "title"})
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/teams/example-team/posts" {
		t.Errorf("request = %s %s", gotMethod, gotPath)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q", gotContentType)
	}
}

func TestUpdatePostNameValidation(t *testing.T) {
	client := NewClient("example-team", "dummy-token")
	for _, input := range []UpdatePostNameInput{{PostNumber: 1}, {Name: "title"}} {
		if _, err := client.UpdatePostName(context.Background(), input); err == nil {
			t.Errorf("UpdatePostName(%+v): want error", input)
		}
	}
}

func TestUpdateTagsRequestContract(t *testing.T) {
	var requests []struct {
		method string
		path   string
		auth   string
		body   map[string]any
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		requests = append(requests, struct {
			method string
			path   string
			auth   string
			body   map[string]any
		}{r.Method, r.URL.Path, r.Header.Get("Authorization"), payload["post"].(map[string]any)})
		_, _ = io.WriteString(w, `{"number":100}`)
	}))
	defer srv.Close()

	client := testClient(srv)
	for _, tags := range [][]string{{"tag-a"}, nil} {
		if err := client.UpdateTags(context.Background(), UpdateTagsInput{PostNumber: 100, Tags: tags}); err != nil {
			t.Fatalf("UpdateTags(%v): %v", tags, err)
		}
	}
	if len(requests) != 2 {
		t.Fatalf("request count = %d, want 2", len(requests))
	}
	for i, request := range requests {
		if request.method != http.MethodPatch || request.path != "/teams/example-team/posts/100" {
			t.Errorf("request[%d] = %s %s", i, request.method, request.path)
		}
		if request.auth != "Bearer dummy-token" {
			t.Errorf("Authorization[%d] = %q", i, request.auth)
		}
		if len(request.body) != 1 {
			t.Errorf("payload[%d] = %#v, want only tags", i, request.body)
		}
		tags, ok := request.body["tags"].([]any)
		if !ok {
			t.Errorf("tags[%d] = %#v, want array", i, request.body["tags"])
		}
		if i == 0 && (len(tags) != 1 || tags[0] != "tag-a") {
			t.Errorf("tags[%d] = %#v, want [tag-a]", i, tags)
		}
		if i == 1 && len(tags) != 0 {
			t.Errorf("tags[%d] = %#v, want empty array", i, tags)
		}
		for _, forbidden := range []string{"name", "body_md", "message", "category"} {
			if _, ok := request.body[forbidden]; ok {
				t.Errorf("payload[%d] = %#v, want no %s", i, request.body, forbidden)
			}
		}
	}
}

func TestUpdateTagsErrors(t *testing.T) {
	t.Run("non-2xx", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = io.WriteString(w, `{"error":"forbidden"}`)
		}))
		defer srv.Close()

		err := testClient(srv).UpdateTags(context.Background(), UpdateTagsInput{PostNumber: 100, Tags: []string{"tag-a"}})
		if err == nil || !strings.Contains(err.Error(), "status 403") {
			t.Fatalf("err = %v, want status 403", err)
		}
	})

	t.Run("decode error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, "{")
		}))
		defer srv.Close()

		err := testClient(srv).UpdateTags(context.Background(), UpdateTagsInput{PostNumber: 100, Tags: []string{"tag-a"}})
		if err == nil || !strings.Contains(err.Error(), "decode response") {
			t.Fatalf("err = %v, want decode response", err)
		}
	})
}

func TestWriteOperationsRejectInvalidPostNumbers(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer srv.Close()

	client := testClient(srv)
	for _, number := range []int{0, -1} {
		tests := []struct {
			name string
			call func() error
		}{
			{
				name: "UpdatePost",
				call: func() error {
					_, err := client.UpdatePost(context.Background(), UpdatePostInput{PostNumber: number, BodyMD: "body"})
					return err
				},
			},
			{
				name: "UpdatePostBodyOnly",
				call: func() error {
					_, err := client.UpdatePostBodyOnly(context.Background(), UpdatePostBodyOnlyInput{PostNumber: number, BodyMD: "body"})
					return err
				},
			},
			{
				name: "UpdatePostName",
				call: func() error {
					_, err := client.UpdatePostName(context.Background(), UpdatePostNameInput{PostNumber: number, Name: "title"})
					return err
				},
			},
			{
				name: "UpdateTags",
				call: func() error {
					return client.UpdateTags(context.Background(), UpdateTagsInput{PostNumber: number, Tags: []string{"tag-a"}})
				},
			},
		}
		for _, tt := range tests {
			if err := tt.call(); err == nil {
				t.Errorf("%s(%d): want error", tt.name, number)
			}
		}
	}
	if requests != 0 {
		t.Fatalf("request count = %d, want 0", requests)
	}
}

func TestDetectContentType(t *testing.T) {
	tests := map[string]string{
		"image.png": "image/png", "image.jpg": "image/jpeg", "image.jpeg": "image/jpeg",
		"image.gif": "image/gif", "image.webp": "image/webp", "image.bmp": "application/octet-stream",
		"video.mov": "video/quicktime", "video.mp4": "video/mp4", "video.m4v": "video/mp4",
		"video.webm": "video/webm", "video.ogv": "video/ogg", "video.avi": "application/octet-stream",
	}
	for path, want := range tests {
		if got := detectContentType(path); got != want {
			t.Errorf("detectContentType(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestUploadFile(t *testing.T) {
	tests := []struct {
		name         string
		fileName     string
		content      string
		wantType     string
		wantUploaded bool
	}{
		{name: "image", fileName: "image.PNG", content: "image-data", wantType: "image/png"},
		{name: "video", fileName: "video.MOV", content: "video-data", wantType: "video/quicktime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var policyType string
			var uploadContentType string
			var uploadedFile *multipart.FileHeader
			var srv *httptest.Server
			srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/teams/example-team/attachments/policies":
					var payload struct {
						Type string `json:"type"`
					}
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Error(err)
					}
					policyType = payload.Type
					_, _ = io.WriteString(w, `{"attachment":{"endpoint":"`+srv.URL+`/upload?signature=upload-secret","url":"https://cdn.example.invalid/attachment"},"form":{"policy":"form-value"}}`)
				case "/upload":
					uploadContentType = r.Header.Get("Content-Type")
					if err := r.ParseMultipartForm(1 << 20); err != nil {
						t.Error(err)
						return
					}
					f, fh, err := r.FormFile("file")
					if err != nil {
						t.Error(err)
						return
					}
					if f != nil {
						_ = f.Close()
					}
					uploadedFile = fh
					w.WriteHeader(http.StatusCreated)
				default:
					http.NotFound(w, r)
				}
			}))
			defer srv.Close()

			filePath := filepath.Join(t.TempDir(), tt.fileName)
			if err := os.WriteFile(filePath, []byte(tt.content), 0600); err != nil {
				t.Fatal(err)
			}
			gotURL, err := testClient(srv).UploadFile(context.Background(), filePath)
			if err != nil {
				t.Fatalf("UploadFile: %v", err)
			}
			if gotURL != "https://cdn.example.invalid/attachment" {
				t.Errorf("URL = %q", gotURL)
			}
			if policyType != tt.wantType {
				t.Errorf("policy type = %q, want %q", policyType, tt.wantType)
			}
			if uploadContentType == "" || !strings.HasPrefix(uploadContentType, "multipart/form-data; boundary=") {
				t.Errorf("upload Content-Type = %q", uploadContentType)
			}
			if uploadedFile == nil || uploadedFile.Filename != tt.fileName {
				t.Fatalf("uploaded file = %#v", uploadedFile)
			}
		})
	}
}

func TestUploadFileRejectsTooLarge(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "huge.bin")
	if err := os.WriteFile(filePath, make([]byte, MaxUploadFileSize+1), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := NewClient("example-team", "dummy-token").UploadFile(context.Background(), filePath)
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("UploadFile error = %v, want ErrFileTooLarge", err)
	}
}

func TestTeamNameAndOptions(t *testing.T) {
	customClient := &http.Client{}
	client := NewClient("example-team", "dummy-token", WithBaseURL("https://example.invalid/v1/"), WithHTTPClient(customClient))
	if client.TeamName() != "example-team" {
		t.Errorf("TeamName() = %q", client.TeamName())
	}
	if client.baseURL != "https://example.invalid/v1" {
		t.Errorf("baseURL = %q", client.baseURL)
	}
	if client.client != customClient {
		t.Error("WithHTTPClient did not install client")
	}

	defaultBase := NewClient("example-team", "dummy-token", WithBaseURL(""))
	if defaultBase.baseURL != defaultBaseURL {
		t.Errorf("empty baseURL option changed base URL to %q", defaultBase.baseURL)
	}
}

func TestTeamNameIsEscapedInPath(t *testing.T) {
	var gotEscapedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEscapedPath = r.URL.EscapedPath()
		_, _ = w.Write([]byte(`{"posts":[],"total_count":0}`))
	}))
	defer srv.Close()

	_, _ = NewClient("example/team", "dummy-token", WithBaseURL(srv.URL), WithHTTPClient(srv.Client())).SearchPosts(context.Background(), SearchPostsInput{})
	if !strings.Contains(gotEscapedPath, "example%2Fteam") {
		t.Errorf("escaped path = %q", gotEscapedPath)
	}
}

func testClient(srv *httptest.Server) *Client {
	return NewClient("example-team", "dummy-token", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
}

func jsonCaptureHandler(target *map[string]any, response string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(target); err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, response)
	})
}

type failingBodyTransport struct {
	base http.RoundTripper
}

func (t failingBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(failingReader{})
	return resp, nil
}

type presignedErrorTransport struct {
	base http.RoundTripper
}

func (t presignedErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/upload" {
		return nil, &url.Error{
			Op:  req.Method,
			URL: req.URL.String(),
			Err: errors.New("connection refused"),
		}
	}
	return t.base.RoundTrip(req)
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}
