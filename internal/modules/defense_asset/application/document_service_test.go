//go:build unit

package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

// mockDocumentRepository — мок-реализация DocumentRepositoryInterface.
type mockDocumentRepository struct {
	documents map[string]*domain.Document
	err       error
}

func newMockDocumentRepository() *mockDocumentRepository {
	return &mockDocumentRepository{
		documents: make(map[string]*domain.Document),
	}
}

func (m *mockDocumentRepository) Save(_ context.Context, doc *domain.Document) error {
	if m.err != nil {
		return m.err
	}
	m.documents[doc.ID()] = doc
	return nil
}

func (m *mockDocumentRepository) FindByID(_ context.Context, id string) (*domain.Document, error) {
	if m.err != nil {
		return nil, m.err
	}
	doc, ok := m.documents[id]
	if !ok {
		return nil, domain.ErrDocumentNotFound
	}
	return doc, nil
}

func (m *mockDocumentRepository) FindByAssetID(_ context.Context, assetID string) ([]*domain.Document, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []*domain.Document
	for _, doc := range m.documents {
		if doc.AssetID() == assetID {
			result = append(result, doc)
		}
	}
	return result, nil
}

func (m *mockDocumentRepository) Delete(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.documents[id]; !ok {
		return domain.ErrDocumentNotFound
	}
	delete(m.documents, id)
	return nil
}

func TestDocumentService_Create_Success(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo)

	input := CreateDocumentInput{
		AssetID:     "asset-id-123",
		Name:        "document.pdf",
		MimeType:    "application/pdf",
		SizeBytes:   2048,
		StorageKey:  "storage/key/document.pdf",
		DownloadURL: "https://download.url/document.pdf",
		OwnerID:     nil,
	}

	doc, err := service.Create(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, "asset-id-123", doc.AssetID())
	assert.Equal(t, "document.pdf", doc.Name())
	assert.Equal(t, "application/pdf", doc.MimeType())
	assert.Equal(t, int64(2048), doc.SizeBytes())
	assert.Equal(t, "storage/key/document.pdf", doc.StorageKey())
	assert.NotEmpty(t, doc.ID())
}

func TestDocumentService_Create_InvalidName(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo)

	input := CreateDocumentInput{
		AssetID:    "asset-id-123",
		Name:       "",
		MimeType:   "application/pdf",
		SizeBytes:  0,
		StorageKey: "storage/key/doc.pdf",
	}

	_, err := service.Create(context.Background(), input)
	assert.ErrorIs(t, err, domain.ErrDocumentInvalidName)
}

func TestDocumentService_Create_InvalidAssetID(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo)

	input := CreateDocumentInput{
		AssetID:    "",
		Name:       "doc.pdf",
		MimeType:   "application/pdf",
		SizeBytes:  0,
		StorageKey: "storage/key/doc.pdf",
	}

	_, err := service.Create(context.Background(), input)
	assert.ErrorIs(t, err, domain.ErrDocumentInvalidAssetID)
}

func TestDocumentService_GetByID_Success(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo)

	input := CreateDocumentInput{
		AssetID:    "asset-id-123",
		Name:       "doc.pdf",
		MimeType:   "application/pdf",
		SizeBytes:  1024,
		StorageKey: "storage/key/doc.pdf",
	}

	created, err := service.Create(context.Background(), input)
	require.NoError(t, err)

	doc, err := service.GetByID(context.Background(), created.ID())
	require.NoError(t, err)
	assert.Equal(t, created.ID(), doc.ID())
	assert.Equal(t, "doc.pdf", doc.Name())
}

func TestDocumentService_GetByID_NotFound(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo)

	_, err := service.GetByID(context.Background(), "non-existent-id")
	assert.ErrorIs(t, err, domain.ErrDocumentNotFound)
}

func TestDocumentService_ListByAssetID_Success(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo)

	// Create 2 documents for asset-1 and 1 for asset-2
	docs := []CreateDocumentInput{
		{AssetID: "asset-1", Name: "doc1.pdf", MimeType: "application/pdf", SizeBytes: 100, StorageKey: "key1"},
		{AssetID: "asset-1", Name: "doc2.pdf", MimeType: "application/pdf", SizeBytes: 200, StorageKey: "key2"},
		{AssetID: "asset-2", Name: "doc3.pdf", MimeType: "application/pdf", SizeBytes: 300, StorageKey: "key3"},
	}

	for _, d := range docs {
		_, err := service.Create(context.Background(), d)
		require.NoError(t, err)
	}

	result, err := service.ListByAssetID(context.Background(), "asset-1")
	require.NoError(t, err)
	assert.Len(t, result, 2)

	result2, err := service.ListByAssetID(context.Background(), "asset-2")
	require.NoError(t, err)
	assert.Len(t, result2, 1)

	result3, err := service.ListByAssetID(context.Background(), "non-existent")
	require.NoError(t, err)
	assert.Len(t, result3, 0)
}

func TestDocumentService_Delete_Success(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo)

	input := CreateDocumentInput{
		AssetID:    "asset-id",
		Name:       "doc.pdf",
		MimeType:   "application/pdf",
		SizeBytes:  100,
		StorageKey: "storage/key/doc.pdf",
	}

	created, err := service.Create(context.Background(), input)
	require.NoError(t, err)

	err = service.Delete(context.Background(), created.ID())
	assert.NoError(t, err)

	_, err = service.GetByID(context.Background(), created.ID())
	assert.ErrorIs(t, err, domain.ErrDocumentNotFound)
}

func TestDocumentService_Delete_NotFound(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo)

	err := service.Delete(context.Background(), "non-existent")
	assert.ErrorIs(t, err, domain.ErrDocumentNotFound)
}

func TestDocumentService_Create_RepoError(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	mockRepo.err = errors.New("db error")
	service := NewDocumentService(mockRepo)

	input := CreateDocumentInput{
		AssetID:    "asset-id",
		Name:       "doc.pdf",
		MimeType:   "application/pdf",
		SizeBytes:  100,
		StorageKey: "storage/key/doc.pdf",
	}

	_, err := service.Create(context.Background(), input)
	assert.ErrorContains(t, err, "save document")
}
