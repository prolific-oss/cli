package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const fileTestToken = "test-token"

func TestGetFileUploadURL(t *testing.T) {
	var gotPath, gotWorkspace string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		gotWorkspace = r.URL.Query().Get("workspace_id")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"upload_url":"https://s3/u","public_url":"https://cdn/p.mp4","content_type":"video/mp4"}`))
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: fileTestToken}

	resp, err := c.GetFileUploadURL("ws-1", "my video.mp4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v1/data-collection/files/upload-url/my%20video.mp4" || gotWorkspace != "ws-1" {
		t.Fatalf("unexpected request: path=%q workspace=%q", gotPath, gotWorkspace)
	}
	if resp.PublicURL != "https://cdn/p.mp4" || resp.ContentType != "video/mp4" || resp.UploadURL != "https://s3/u" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestDeleteFile(t *testing.T) {
	var gotMethod, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := Client{Client: server.Client(), BaseURL: server.URL, Token: fileTestToken}

	if err := c.DeleteFile("ws-1", "abc123", "intro.mp4"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/api/v1/data-collection/files/ws-1/abc123/intro.mp4" {
		t.Fatalf("unexpected request: %s %s", gotMethod, gotPath)
	}
}
