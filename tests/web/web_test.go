package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"NDClasses/clients/web"
)

func TestNew(t *testing.T) {
	// Just test that New() doesn't panic
	_ = web.New()
}

func TestUpdates(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test content"))
	}))
	defer server.Close()

	client := web.New()

	updates, err := client.Updates(server.URL)
	if err != nil {
		t.Fatalf("Updates failed: %v", err)
	}

	if len(updates) != 1 {
		t.Errorf("Expected 1 update, got %d", len(updates))
	}

	if updates[0].ID != 1 {
		t.Errorf("Expected update ID 1, got %d", updates[0].ID)
	}

	if updates[0].Content != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", updates[0].Content)
	}
}

func TestSendMessage(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/x-www-form-urlencoded" {
			t.Errorf("Expected content type application/x-www-form-urlencoded, got %s", contentType)
		}

		// Check query parameters
		query := r.URL.RawQuery
		if !strings.Contains(query, "data=test+message") {
			t.Errorf("Expected query to contain 'data=test+message', got '%s'", query)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := web.New()

	err := client.SendMessage(server.URL, "test message")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
}
