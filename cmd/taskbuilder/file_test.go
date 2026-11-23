package taskbuilder_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/taskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

const testFilePublicURL = "https://cdn.example.com/files/ws-1/0123456789abcdef0123456789abcdef/intro.mp4"

func runFileUpload(t *testing.T, c *mock_client.MockAPI, args ...string) (string, error) {
	t.Helper()
	var b bytes.Buffer
	cmd := taskbuilder.NewFileUploadCommand(c, &b)
	cmd.SetArgs(args)
	cmd.SetOut(&b)
	cmd.SetErr(&b)
	err := cmd.Execute()
	return b.String(), err
}

func TestFileUploadPrintsPublicURL(t *testing.T) {
	var gotType, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotType = r.Header.Get("Content-Type")
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		gotBody = buf.String()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "intro.mp4")
	if err := os.WriteFile(path, []byte("video-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	c := mock_client.NewMockAPI(ctrl)
	c.EXPECT().GetFileUploadURL("ws-1", "intro.mp4").Return(&client.GetFileUploadURLResponse{
		UploadURL: srv.URL, PublicURL: testFilePublicURL, ContentType: "video/mp4",
	}, nil)

	out, err := runFileUpload(t, c, "-w", "ws-1", "-f", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != testFilePublicURL {
		t.Fatalf("expected public URL, got %q", out)
	}
	if gotType != "video/mp4" || gotBody != "video-bytes" {
		t.Fatalf("unexpected upload: type=%q body=%q", gotType, gotBody)
	}
}

func TestFileUploadRejectsNonVideo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	c := mock_client.NewMockAPI(gomock.NewController(t))
	_, err := runFileUpload(t, c, "-w", "ws-1", "-f", path)
	if err == nil || !strings.Contains(err.Error(), "unsupported file type") || !strings.Contains(err.Error(), ".mp4, .webm") {
		t.Fatalf("expected unsupported file type error, got %v", err)
	}
}

func TestFileUploadMissingFile(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	_, err := runFileUpload(t, c, "-w", "ws-1", "-f", filepath.Join(t.TempDir(), "missing.mp4"))
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected missing file error, got %v", err)
	}
}

func TestFileUploadAPIError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "intro.mp4")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().GetFileUploadURL("ws-1", "intro.mp4").Return(nil, errors.New("forbidden"))

	_, err := runFileUpload(t, c, "-w", "ws-1", "-f", path)
	if err == nil || !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("expected API error, got %v", err)
	}
}

func TestFileDeleteParsesPublicURL(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().DeleteFile("ws-1", "0123456789abcdef0123456789abcdef", "intro.mp4").Return(nil)

	var b bytes.Buffer
	cmd := taskbuilder.NewFileDeleteCommand(c, &b)
	if err := cmd.RunE(cmd, []string{testFilePublicURL}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(b.String(), "intro.mp4 is no longer hosted") {
		t.Fatalf("unexpected output %q", b.String())
	}
}

func TestFileDeleteRejectsNonFileURL(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	cmd := taskbuilder.NewFileDeleteCommand(c, &bytes.Buffer{})

	for _, u := range []string{"not a url", "https://cdn.example.com/other/a/b/c.mp4", "https://cdn.example.com/files/ws-1/intro.mp4"} {
		if err := cmd.RunE(cmd, []string{u}); err == nil {
			t.Fatalf("expected error for %q", u)
		}
	}
}

func TestFileDeleteAPIError(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().DeleteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("boom"))

	cmd := taskbuilder.NewFileDeleteCommand(c, &bytes.Buffer{})
	if err := cmd.RunE(cmd, []string{testFilePublicURL}); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected API error, got %v", err)
	}
}
