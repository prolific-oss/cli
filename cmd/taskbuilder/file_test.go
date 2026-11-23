package taskbuilder_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/cmd/taskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

const testFilePublicURL = "https://cdn.example.com/files/ws-1/0123456789abcdef0123456789abcdef/intro.mp4"

func runFileUpload(t *testing.T, c *mock_client.MockAPI, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	cmd := taskbuilder.NewFileUploadCommand(c, &out)
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

func writeTempFile(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
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

	out, errOut, err := runFileUpload(t, c, "-w", "ws-1", "-f", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != testFilePublicURL {
		t.Fatalf("expected only the public URL on stdout, got %q", out)
	}
	if !strings.Contains(errOut, "virus scanned") {
		t.Fatalf("expected the scan notice on stderr, got %q", errOut)
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
	_, _, err := runFileUpload(t, c, "-w", "ws-1", "-f", path)
	if err == nil || !strings.Contains(err.Error(), "unsupported file type") || !strings.Contains(err.Error(), ".mp4, .webm") {
		t.Fatalf("expected unsupported file type error, got %v", err)
	}
}

func TestFileUploadMissingFile(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	_, _, err := runFileUpload(t, c, "-w", "ws-1", "-f", filepath.Join(t.TempDir(), "missing.mp4"))
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

	_, _, err := runFileUpload(t, c, "-w", "ws-1", "-f", path)
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

func TestFileUploadValidation(t *testing.T) {
	dir := t.TempDir()
	big := filepath.Join(dir, "big.mp4")
	f, err := os.Create(big)
	if err != nil {
		t.Fatal(err)
	}
	// Sparse file: one byte over the 5 GiB limit without writing the data.
	if err := f.Truncate(5<<30 + 1); err != nil {
		t.Fatal(err)
	}
	f.Close()

	subdir := filepath.Join(dir, "folder.mp4")
	if err := os.Mkdir(subdir, 0o700); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"empty file", []string{"-w", "ws-1", "-f", writeTempFile(t, "empty.mp4", nil)}, "file is empty"},
		{"directory", []string{"-w", "ws-1", "-f", subdir}, "is a directory"},
		{"over 5 GiB", []string{"-w", "ws-1", "-f", big}, "the limit is"},
		{"uppercase extension is accepted by the type check", []string{"-w", "ws-1", "-f", filepath.Join(dir, "missing.MP4")}, "does not exist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			_, _, err := runFileUpload(t, c, tt.args...)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestFileUploadRequiresFlags(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	for _, args := range [][]string{{}, {"-w", "ws-1"}, {"-f", "clip.mp4"}} {
		if _, _, err := runFileUpload(t, c, args...); err == nil {
			t.Fatalf("expected an error for args %v", args)
		}
	}
}

func TestFileUploadStorageRejectsUpload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("SignatureDoesNotMatch"))
	}))
	defer srv.Close()

	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().GetFileUploadURL("ws-1", "intro.mp4").Return(&client.GetFileUploadURLResponse{
		UploadURL: srv.URL, PublicURL: testFilePublicURL, ContentType: "video/mp4",
	}, nil)

	out, errOut, err := runFileUpload(t, c, "-w", "ws-1", "-f", writeTempFile(t, "intro.mp4", []byte("x")))
	if err == nil || !strings.Contains(err.Error(), "SignatureDoesNotMatch") {
		t.Fatalf("expected the storage error, got %v", err)
	}
	if out != "" || strings.Contains(errOut, "virus scanned") {
		t.Fatalf("expected no URL and no scan notice after a failed upload, got out=%q err=%q", out, errOut)
	}
}

func TestTaskBuilderListsFileCommands(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	var out bytes.Buffer
	cmd := taskbuilder.NewTaskBuilderCommand(c, &out)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"file", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"upload", "delete"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("expected help to list %q, got %q", want, out.String())
		}
	}
}

func fileHostingGateError() error {
	return fmt.Errorf("unable to fulfil request: %w", &client.UnrecognizedAPIError{
		StatusCode: http.StatusForbidden,
		Body:       []byte(`{"message":"File uploads and deletes are not enabled for this workspace."}`),
	})
}

func TestFileUploadGateDenied(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().GetFileUploadURL("ws-1", "intro.mp4").Return(nil, fileHostingGateError())

	_, _, err := runFileUpload(t, c, "-w", "ws-1", "-f", writeTempFile(t, "intro.mp4", []byte("x")))
	if err == nil || !strings.Contains(err.Error(), shared.ErrFileHostingNotEnabled) {
		t.Fatalf("expected the gate message, got %v", err)
	}
}

func TestFileDeleteGateDenied(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().DeleteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(fileHostingGateError())

	cmd := taskbuilder.NewFileDeleteCommand(c, &bytes.Buffer{})
	if err := cmd.RunE(cmd, []string{testFilePublicURL}); err == nil || !strings.Contains(err.Error(), shared.ErrFileHostingNotEnabled) {
		t.Fatalf("expected the gate message, got %v", err)
	}
}

func TestFileUploadErrorDoesNotShowPresignedURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	uploadURL := srv.URL + "/video.mp4?X-Amz-Signature=secret-signature"
	srv.Close() // connection refused

	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().GetFileUploadURL("ws-1", "intro.mp4").Return(&client.GetFileUploadURLResponse{
		UploadURL: uploadURL, PublicURL: testFilePublicURL, ContentType: "video/mp4",
	}, nil)

	_, _, err := runFileUpload(t, c, "-w", "ws-1", "-f", writeTempFile(t, "intro.mp4", []byte("x")))
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "secret-signature") {
		t.Fatalf("error leaks the presigned URL: %v", err)
	}
}
