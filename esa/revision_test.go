package esa

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestListRevisions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotMethod, gotPath, gotAuth string
		var gotQuery url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			gotAuth = r.Header.Get("Authorization")
			gotQuery = r.URL.Query()
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"revisions":[{"number":5,"name":"example title","full_name":"example/category/example title","category":"example/category","tags":["tag-a"],"body_md":"body","body_html":"<p>body</p>","diff":"<div class=\"diff\"></div>","message":"append entry","created_at":"2026-08-19T14:44:09+09:00","updated_at":"2026-08-19T14:44:09+09:00","wip":false,"created_by":{"name":"Example User","screen_name":"example","icon":"https://img.example.invalid/icon.png"},"url":"https://example.esa.io/posts/42/revisions/5"}],"prev_page":null,"next_page":2,"total_count":25,"page":1,"per_page":20,"max_per_page":100}`)
		}))
		defer srv.Close()

		result, err := testClient(srv).ListRevisions(context.Background(), ListRevisionsInput{PostNumber: 42, Page: 1, PerPage: 20})
		if err != nil {
			t.Fatalf("ListRevisions: %v", err)
		}
		if gotMethod != http.MethodGet || gotPath != "/teams/example-team/posts/42/revisions" {
			t.Errorf("request = %s %s", gotMethod, gotPath)
		}
		if gotAuth != "Bearer dummy-token" {
			t.Errorf("Authorization = %q", gotAuth)
		}
		if gotQuery.Get("page") != "1" || gotQuery.Get("per_page") != "20" {
			t.Errorf("query = %v", gotQuery)
		}
		if result.TotalCount != 25 || result.Page != 1 || result.PerPage != 20 || result.MaxPerPage != 100 {
			t.Errorf("pagination = %#v", result)
		}
		if result.PrevPage != nil {
			t.Errorf("PrevPage = %#v, want nil", result.PrevPage)
		}
		if result.NextPage == nil || *result.NextPage != 2 {
			t.Errorf("NextPage = %#v, want 2", result.NextPage)
		}
		if len(result.Revisions) != 1 {
			t.Fatalf("revisions = %#v", result.Revisions)
		}
		revision := result.Revisions[0]
		if revision.Number != 5 || revision.Name != "example title" || revision.FullName != "example/category/example title" ||
			revision.Category != "example/category" || revision.BodyMD != "body" || revision.BodyHTML != "<p>body</p>" ||
			revision.Diff != `<div class="diff"></div>` || revision.Message != "append entry" ||
			revision.CreatedAt != "2026-08-19T14:44:09+09:00" || revision.UpdatedAt != "2026-08-19T14:44:09+09:00" ||
			revision.WIP || revision.URL != "https://example.esa.io/posts/42/revisions/5" {
			t.Errorf("revision = %#v", revision)
		}
		if len(revision.Tags) != 1 || revision.Tags[0] != "tag-a" {
			t.Errorf("tags = %#v", revision.Tags)
		}
		if revision.CreatedBy.ScreenName != "example" || revision.CreatedBy.Name != "Example User" {
			t.Errorf("created_by = %#v", revision.CreatedBy)
		}
	})

	t.Run("omits pagination query when unset", func(t *testing.T) {
		var gotRawQuery string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotRawQuery = r.URL.RawQuery
			_, _ = io.WriteString(w, `{"revisions":[]}`)
		}))
		defer srv.Close()

		if _, err := testClient(srv).ListRevisions(context.Background(), ListRevisionsInput{PostNumber: 42}); err != nil {
			t.Fatalf("ListRevisions: %v", err)
		}
		if gotRawQuery != "" {
			t.Errorf("query = %q, want empty", gotRawQuery)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		client := NewClient("example-team", "dummy-token")
		inputs := []ListRevisionsInput{
			{PostNumber: 0},
			{PostNumber: -1},
			{PostNumber: 42, Page: -1},
			{PostNumber: 42, PerPage: -1},
			{PostNumber: 42, PerPage: MaxSearchPerPage + 1},
		}
		for _, input := range inputs {
			if _, err := client.ListRevisions(context.Background(), input); err == nil {
				t.Errorf("ListRevisions(%+v): want error", input)
			}
		}
	})

	t.Run("404 is not found", func(t *testing.T) {
		srv := httptest.NewServer(notFoundHandler())
		defer srv.Close()

		_, err := testClient(srv).ListRevisions(context.Background(), ListRevisionsInput{PostNumber: 42})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("decode error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, "{")
		}))
		defer srv.Close()

		_, err := testClient(srv).ListRevisions(context.Background(), ListRevisionsInput{PostNumber: 42})
		if err == nil || !strings.Contains(err.Error(), "list revisions post 42") || !strings.Contains(err.Error(), "decode response") {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestGetRevision(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotMethod, gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"revision":{"number":3,"name":"example title","body_md":"body","message":"append entry","wip":true,"created_by":{"screen_name":"example"},"url":"https://example.esa.io/posts/42/revisions/3"}}`)
		}))
		defer srv.Close()

		revision, err := testClient(srv).GetRevision(context.Background(), GetRevisionInput{PostNumber: 42, RevisionNumber: 3})
		if err != nil {
			t.Fatalf("GetRevision: %v", err)
		}
		if gotMethod != http.MethodGet || gotPath != "/teams/example-team/posts/42/revisions/3" {
			t.Errorf("request = %s %s", gotMethod, gotPath)
		}
		if revision.Number != 3 || revision.BodyMD != "body" || revision.Message != "append entry" || !revision.WIP {
			t.Errorf("revision = %#v", revision)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		client := NewClient("example-team", "dummy-token")
		inputs := []GetRevisionInput{
			{PostNumber: 0, RevisionNumber: 1},
			{PostNumber: 42, RevisionNumber: 0},
			{PostNumber: 42, RevisionNumber: -1},
		}
		for _, input := range inputs {
			if _, err := client.GetRevision(context.Background(), input); err == nil {
				t.Errorf("GetRevision(%+v): want error", input)
			}
		}
	})

	t.Run("404 is not found", func(t *testing.T) {
		srv := httptest.NewServer(notFoundHandler())
		defer srv.Close()

		_, err := testClient(srv).GetRevision(context.Background(), GetRevisionInput{PostNumber: 42, RevisionNumber: 999})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("non-404 non-2xx", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `{"error":"internal"}`)
		}))
		defer srv.Close()

		_, err := testClient(srv).GetRevision(context.Background(), GetRevisionInput{PostNumber: 42, RevisionNumber: 3})
		if err == nil || !strings.Contains(err.Error(), "status 500") {
			t.Fatalf("err = %v, want status 500", err)
		}
		if errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want not ErrNotFound", err)
		}
	})
}

func TestCompareRevisions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotMethod, gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"diff":{"from_revision_number":4,"to_revision_number":5,"diff_html":"<div class=\"diff\"></div>","diff_text":" body"}}`)
		}))
		defer srv.Close()

		diff, err := testClient(srv).CompareRevisions(context.Background(), CompareRevisionsInput{
			PostNumber: 42, FromRevisionNumber: 4, ToRevisionNumber: 5,
		})
		if err != nil {
			t.Fatalf("CompareRevisions: %v", err)
		}
		if gotMethod != http.MethodGet || gotPath != "/teams/example-team/posts/42/revisions/compare/4...5" {
			t.Errorf("request = %s %s", gotMethod, gotPath)
		}
		if diff.FromRevisionNumber != 4 || diff.ToRevisionNumber != 5 ||
			diff.DiffHTML != `<div class="diff"></div>` || diff.DiffText != " body" {
			t.Errorf("diff = %#v", diff)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		client := NewClient("example-team", "dummy-token")
		inputs := []CompareRevisionsInput{
			{PostNumber: 0, FromRevisionNumber: 1, ToRevisionNumber: 2},
			{PostNumber: 42, FromRevisionNumber: 0, ToRevisionNumber: 2},
			{PostNumber: 42, FromRevisionNumber: 1, ToRevisionNumber: 0},
		}
		for _, input := range inputs {
			if _, err := client.CompareRevisions(context.Background(), input); err == nil {
				t.Errorf("CompareRevisions(%+v): want error", input)
			}
		}
	})

	t.Run("404 is not found", func(t *testing.T) {
		srv := httptest.NewServer(notFoundHandler())
		defer srv.Close()

		_, err := testClient(srv).CompareRevisions(context.Background(), CompareRevisionsInput{
			PostNumber: 42, FromRevisionNumber: 1, ToRevisionNumber: 999,
		})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}

func TestRollbackRevision(t *testing.T) {
	t.Run("success with wip and message", func(t *testing.T) {
		var gotMethod, gotPath, gotContentType string
		var payload map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			gotContentType = r.Header.Get("Content-Type")
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"number":42,"revision_number":6,"url":"https://example.esa.io/posts/42"}`)
		}))
		defer srv.Close()

		wip := true
		message := "revert oura summary"
		post, err := testClient(srv).RollbackRevision(context.Background(), RollbackRevisionInput{
			PostNumber: 42, RevisionNumber: 4, WIP: &wip, Message: &message,
		})
		if err != nil {
			t.Fatalf("RollbackRevision: %v", err)
		}
		if gotMethod != http.MethodPost || gotPath != "/teams/example-team/posts/42/revisions/4/rollback" {
			t.Errorf("request = %s %s", gotMethod, gotPath)
		}
		if gotContentType != "application/json" {
			t.Errorf("Content-Type = %q", gotContentType)
		}
		postPayload, ok := payload["post"].(map[string]any)
		if !ok {
			t.Fatalf("payload = %#v", payload)
		}
		if postPayload["wip"] != true || postPayload["message"] != "revert oura summary" {
			t.Errorf("post payload = %#v", postPayload)
		}
		if post.Number != 42 || post.RevisionNumber != 6 {
			t.Errorf("post = %#v", post)
		}
	})

	t.Run("omits wip and message when unset", func(t *testing.T) {
		var payload map[string]any
		srv := httptest.NewServer(jsonCaptureHandler(&payload, `{"number":42,"revision_number":6}`))
		defer srv.Close()

		if _, err := testClient(srv).RollbackRevision(context.Background(), RollbackRevisionInput{
			PostNumber: 42, RevisionNumber: 4,
		}); err != nil {
			t.Fatalf("RollbackRevision: %v", err)
		}
		postPayload, ok := payload["post"].(map[string]any)
		if !ok {
			t.Fatalf("payload = %#v", payload)
		}
		if len(postPayload) != 0 {
			t.Errorf("post payload = %#v, want no fields", postPayload)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		client := NewClient("example-team", "dummy-token")
		inputs := []RollbackRevisionInput{
			{PostNumber: 0, RevisionNumber: 1},
			{PostNumber: 42, RevisionNumber: 0},
			{PostNumber: 42, RevisionNumber: -1},
		}
		for _, input := range inputs {
			if _, err := client.RollbackRevision(context.Background(), input); err == nil {
				t.Errorf("RollbackRevision(%+v): want error", input)
			}
		}
	})

	t.Run("404 is not found", func(t *testing.T) {
		srv := httptest.NewServer(notFoundHandler())
		defer srv.Close()

		_, err := testClient(srv).RollbackRevision(context.Background(), RollbackRevisionInput{
			PostNumber: 42, RevisionNumber: 999,
		})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
		if errors.Is(err, ErrRollbackToLatestRevision) {
			t.Fatalf("err = %v, want not ErrRollbackToLatestRevision", err)
		}
	})

	t.Run("400 is rollback to latest revision", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":"bad_request","message":"cannot rollback to the latest revision"}`)
		}))
		defer srv.Close()

		_, err := testClient(srv).RollbackRevision(context.Background(), RollbackRevisionInput{
			PostNumber: 42, RevisionNumber: 6,
		})
		if !errors.Is(err, ErrRollbackToLatestRevision) {
			t.Fatalf("err = %v, want ErrRollbackToLatestRevision", err)
		}
		if errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want not ErrNotFound", err)
		}
		if !strings.Contains(err.Error(), "rollback post 42 to revision 6") || !strings.Contains(err.Error(), "status 400") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("other non-2xx", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"error":"unauthorized"}`)
		}))
		defer srv.Close()

		_, err := testClient(srv).RollbackRevision(context.Background(), RollbackRevisionInput{
			PostNumber: 42, RevisionNumber: 4,
		})
		if err == nil || !strings.Contains(err.Error(), "status 401") {
			t.Fatalf("err = %v, want status 401", err)
		}
		if errors.Is(err, ErrRollbackToLatestRevision) || errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want untyped error", err)
		}
	})
}

func notFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":"not_found"}`)
	})
}
