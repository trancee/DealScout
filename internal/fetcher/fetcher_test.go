package fetcher_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trancee/DealScout/internal/fetcher"
)

func TestGetReturnsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello world"))
	}))
	defer server.Close()

	f := fetcher.New(0, 1)
	body, err := f.Get(server.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(body) != "hello world" {
		t.Errorf("body = %q, want %q", string(body), "hello world")
	}
}

func TestPostWithTemplateReplacements(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	f := fetcher.New(0, 1)
	template := `{"offset": {offset}, "limit": {limit}}`
	replacements := map[string]string{"{offset}": "24", "{limit}": "12"}

	_, err := f.Post(server.URL, template, replacements, nil)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}

	want := `{"offset": 24, "limit": 12}`
	if receivedBody != want {
		t.Errorf("body = %q, want %q", receivedBody, want)
	}
}

func TestUserAgentRotation(t *testing.T) {
	var agents []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agents = append(agents, r.Header.Get("User-Agent"))
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	f := fetcher.New(0, 1)
	for range 3 {
		_, _ = f.Get(server.URL)
	}

	if len(agents) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(agents))
	}

	// At least one pair should differ (rotation).
	allSame := agents[0] == agents[1] && agents[1] == agents[2]
	if allSame {
		t.Error("all User-Agents are the same, expected rotation")
	}

	// All should be non-empty.
	for i, ua := range agents {
		if ua == "" {
			t.Errorf("User-Agent[%d] is empty", i)
		}
	}
}

func TestRetryOn429ThenSucceeds(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	f := fetcher.New(0, 3).WithRetryBaseDelay(0)
	body, err := f.Get(server.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(body) != "success" {
		t.Errorf("body = %q, want %q", string(body), "success")
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestRetryExhausted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	f := fetcher.New(0, 2).WithRetryBaseDelay(0)
	_, err := f.Get(server.URL)
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
}

func TestNoRetryOn4xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	f := fetcher.New(0, 3).WithRetryBaseDelay(0)
	_, err := f.Get(server.URL)
	if err == nil {
		t.Fatal("expected error for 403")
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (no retry on 403)", attempts)
	}
}
