package httpclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testRequest struct {
	Message string `json:"message"`
}

type testResponse struct {
	Reply string `json:"reply"`
}

func TestPost_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("content-type") != "application/json" {
			t.Errorf("expected content-type application/json, got %s", r.Header.Get("content-type"))
		}

		var req testRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(testResponse{Reply: "pong"}) //nolint:errcheck
	}))
	defer srv.Close()

	client := NewHTTPClient(srv.Client(), srv.URL)
	resp, err := Post[testRequest, testResponse](client, "/test", testRequest{Message: "ping"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Reply != "pong" {
		t.Errorf("expected Reply=pong, got %q", resp.Reply)
	}
}

func TestPost_NonOKStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewHTTPClient(srv.Client(), srv.URL)
	_, err := Post[testRequest, testResponse](client, "/test", testRequest{Message: "ping"})
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}

func TestPost_InvalidResponseBodyReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`)) //nolint:errcheck
	}))
	defer srv.Close()

	client := NewHTTPClient(srv.Client(), srv.URL)
	_, err := Post[testRequest, testResponse](client, "/test", testRequest{Message: "ping"})
	if err == nil {
		t.Fatal("expected error for invalid response JSON, got nil")
	}
}

func TestPost_ServerUnavailableReturnsError(t *testing.T) {
	// サーバーを起動してすぐに閉じることで接続失敗をシミュレートする
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	client := NewHTTPClient(srv.Client(), srv.URL)
	_, err := Post[testRequest, testResponse](client, "/test", testRequest{Message: "ping"})
	if err == nil {
		t.Fatal("expected error for unavailable server, got nil")
	}
}

func TestPost_RequestSentToCorrectPath(t *testing.T) {
	receivedPath := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(testResponse{Reply: "ok"}) //nolint:errcheck
	}))
	defer srv.Close()

	client := NewHTTPClient(srv.Client(), srv.URL)
	_, err := Post[testRequest, testResponse](client, "/my-endpoint", testRequest{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if receivedPath != "/my-endpoint" {
		t.Errorf("expected path=/my-endpoint, got %q", receivedPath)
	}
}

func TestPost_StatusNotFoundReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := NewHTTPClient(srv.Client(), srv.URL)
	_, err := Post[testRequest, testResponse](client, "/missing", testRequest{})
	if err == nil {
		t.Fatal("expected error for 404 status, got nil")
	}
}
