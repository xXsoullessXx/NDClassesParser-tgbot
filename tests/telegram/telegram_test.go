package telegram_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"NDClasses/clients/telegram"
)

func createTestClient(serverURL string) telegram.Client {
	testURL, _ := url.Parse(serverURL)
	// Use http scheme for testing
	httpHost := "http://" + testURL.Host
	client := telegram.New(httpHost, "test_token")
	return client
}

func TestNew(t *testing.T) {
	host := "api.telegram.org"
	token := "test_token"

	client := telegram.New(host, token)

	// Can't test private fields from external package
	// Just test that it doesn't panic
	_ = client
}

func TestUpdates(t *testing.T) {
	// Mock response
	mockResponse := telegram.UpdatesResponse{
		Ok: true,
		Result: []telegram.Update{
			{
				ID: 1,
				Message: telegram.Message{
					MessageID: 123,
					Chat: telegram.Chat{
						ID:   456,
						Type: "private",
					},
					Text: "test message",
				},
			},
		},
	}

	responseJSON, _ := json.Marshal(mockResponse)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("offset") != "0" {
			t.Errorf("Expected offset '0', got '%s'", query.Get("offset"))
		}
		if query.Get("limit") != "10" {
			t.Errorf("Expected limit '10', got '%s'", query.Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	}))
	defer server.Close()

	client := createTestClient(server.URL)

	updates, err := client.Updates(0, 10)
	if err != nil {
		t.Fatalf("Updates failed: %v", err)
	}

	if len(updates) != 1 {
		t.Errorf("Expected 1 update, got %d", len(updates))
	}

	if updates[0].ID != 1 {
		t.Errorf("Expected update ID 1, got %d", updates[0].ID)
	}

	if updates[0].Message.Text != "test message" {
		t.Errorf("Expected message text 'test message', got '%s'", updates[0].Message.Text)
	}
}

func TestSendMessage(t *testing.T) {
	// Mock response
	mockResponse := map[string]interface{}{
		"ok": true,
	}
	responseJSON, _ := json.Marshal(mockResponse)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("chat_id") != "123456" {
			t.Errorf("Expected chat_id '123456', got '%s'", query.Get("chat_id"))
		}
		if query.Get("text") != "test message" {
			t.Errorf("Expected text 'test message', got '%s'", query.Get("text"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	}))
	defer server.Close()

	client := createTestClient(server.URL)

	err := client.SendMessage(123456, "test message")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
}

func TestUpdatesInvalidJSON(t *testing.T) {
	// Create test server with invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := createTestClient(server.URL)

	_, err := client.Updates(0, 10)
	if err == nil {
		t.Error("Expected JSON parsing error, but got none")
	}

	if !strings.Contains(err.Error(), "can't parse json") {
		t.Errorf("Expected JSON parsing error, got: %v", err)
	}
}
