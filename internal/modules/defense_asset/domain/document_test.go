//go:build unit

package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDocument_ValidInput(t *testing.T) {
	now := time.Now().UTC()
	ownerID := "550e8400-e29b-41d4-a716-446655440000"

	doc, err := NewDocument(
		"doc-id-123",
		"asset-id-456",
		"test-document.pdf",
		"application/pdf",
		"storage/key/file.pdf",
		"https://download.url/file.pdf",
		1024,
		&ownerID,
		now,
		now,
	)

	assert.NoError(t, err)
	assert.NotNil(t, doc)
	assert.Equal(t, "doc-id-123", doc.ID())
	assert.Equal(t, "asset-id-456", doc.AssetID())
	assert.Equal(t, "test-document.pdf", doc.Name())
	assert.Equal(t, "application/pdf", doc.MimeType())
	assert.Equal(t, int64(1024), doc.SizeBytes())
	assert.Equal(t, "storage/key/file.pdf", doc.StorageKey())
	assert.Equal(t, "https://download.url/file.pdf", doc.DownloadURL())
	assert.Equal(t, &ownerID, doc.OwnerID())
}

func TestNewDocument_EmptyName(t *testing.T) {
	now := time.Now().UTC()
	_, err := NewDocument("id", "asset-id", "", "application/pdf", "storage/key", "", 0, nil, now, now)
	assert.ErrorIs(t, err, ErrDocumentInvalidName)
}

func TestNewDocument_EmptyAssetID(t *testing.T) {
	now := time.Now().UTC()
	_, err := NewDocument("id", "", "doc.pdf", "application/pdf", "storage/key", "", 0, nil, now, now)
	assert.ErrorIs(t, err, ErrDocumentInvalidAssetID)
}

func TestNewDocument_NegativeSize(t *testing.T) {
	now := time.Now().UTC()
	_, err := NewDocument("id", "asset-id", "doc.pdf", "application/pdf", "storage/key", "", -1, nil, now, now)
	assert.ErrorIs(t, err, ErrDocumentInvalidSizeBytes)
}

func TestNewDocument_EmptyStorageKey(t *testing.T) {
	now := time.Now().UTC()
	_, err := NewDocument("id", "asset-id", "doc.pdf", "application/pdf", "", "", 0, nil, now, now)
	assert.ErrorIs(t, err, ErrDocumentInvalidStorageKey)
}

func TestNewDocument_DefaultMimeType(t *testing.T) {
	now := time.Now().UTC()
	doc, err := NewDocument("id", "asset-id", "doc.pdf", "", "storage/key", "", 0, nil, now, now)
	assert.NoError(t, err)
	assert.Equal(t, "application/octet-stream", doc.MimeType())
}

func TestDocument_SetDownloadURL(t *testing.T) {
	now := time.Now().UTC()
	doc, err := NewDocument("id", "asset-id", "doc.pdf", "application/pdf", "storage/key", "", 0, nil, now, now)
	assert.NoError(t, err)

	initialUpdatedAt := doc.UpdatedAt()
	doc.SetDownloadURL("https://new-download.url/file.pdf")
	assert.Equal(t, "https://new-download.url/file.pdf", doc.DownloadURL())
	assert.True(t, doc.UpdatedAt().After(initialUpdatedAt) || doc.UpdatedAt().Equal(initialUpdatedAt))
}

func TestDocument_UpdateName(t *testing.T) {
	now := time.Now().UTC()
	doc, err := NewDocument("id", "asset-id", "doc.pdf", "application/pdf", "storage/key", "", 0, nil, now, now)
	assert.NoError(t, err)

	err = doc.UpdateName("new-name.pdf")
	assert.NoError(t, err)
	assert.Equal(t, "new-name.pdf", doc.Name())

	err = doc.UpdateName("")
	assert.ErrorIs(t, err, ErrDocumentInvalidName)
}
