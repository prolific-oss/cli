package collection_test

import (
	"strings"
	"testing"

	"github.com/prolific-oss/cli/config"
	"github.com/prolific-oss/cli/ui/collection"
)

func TestGetCollectionPreviewPath(t *testing.T) {
	collectionID := "test-collection-123"

	path := collection.GetCollectionPreviewPath(collectionID)

	expectedPath := "data-collection-tool/collections/test-collection-123?preview=true"
	if path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, path)
	}
}

func TestGetCollectionPreviewPathContainsPreviewParam(t *testing.T) {
	collectionID := "abc123"

	path := collection.GetCollectionPreviewPath(collectionID)

	if !strings.Contains(path, "preview=true") {
		t.Fatalf("expected path to contain 'preview=true', got %q", path)
	}
}

func TestGetCollectionPreviewURL(t *testing.T) {
	collectionID := "test-collection-456"

	url := collection.GetCollectionPreviewURL(collectionID)

	expectedURL := config.GetApplicationURL() + "/data-collection-tool/collections/test-collection-456?preview=true"
	if url != expectedURL {
		t.Fatalf("expected URL %q, got %q", expectedURL, url)
	}
}

func TestGetCollectionPreviewURLUsesApplicationURL(t *testing.T) {
	collectionID := "xyz789"

	url := collection.GetCollectionPreviewURL(collectionID)

	if !strings.HasPrefix(url, config.GetApplicationURL()) {
		t.Fatalf("expected URL to start with application URL %q, got %q", config.GetApplicationURL(), url)
	}
}

func TestGetCollectionPreviewURLContainsCollectionID(t *testing.T) {
	collectionID := "my-unique-id"

	url := collection.GetCollectionPreviewURL(collectionID)

	if !strings.Contains(url, collectionID) {
		t.Fatalf("expected URL to contain collection ID %q, got %q", collectionID, url)
	}
}
