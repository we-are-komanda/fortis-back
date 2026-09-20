//go:build unit

package application

import (
	"context"
	"errors"
	"github.com/fortis/backend/internal/auth"
	"github.com/google/uuid"
	"testing"
	"time"

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

func TestDocumentService_Create_MetadataRejected(t *testing.T) {
	repo := newMockDocumentRepository()
	s := NewDocumentService(repo, testDocumentAssets(t))
	doc, err := s.Create(context.Background(), "actor", CreateDocumentInput{AssetID: "asset-id-123", Name: "sample.txt", StorageKey: "external", DownloadURL: "https://example.test/public"})
	require.ErrorIs(t, err, domain.ErrDocumentUploadRequired)
	require.Nil(t, doc)
	require.Empty(t, repo.documents)
}

func TestDocumentService_Create_InvalidName(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo, testDocumentAssets(t))

	input := CreateDocumentInput{
		AssetID:    "asset-id-123",
		Name:       "",
		MimeType:   "application/pdf",
		SizeBytes:  0,
		StorageKey: "storage/key/doc.pdf",
	}

	_, err := service.Create(context.Background(), "actor", input)
	assert.ErrorIs(t, err, domain.ErrDocumentUploadRequired)
}

func TestDocumentService_Create_InvalidAssetID(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo, testDocumentAssets(t))

	input := CreateDocumentInput{
		AssetID:    "",
		Name:       "doc.pdf",
		MimeType:   "application/pdf",
		SizeBytes:  0,
		StorageKey: "storage/key/doc.pdf",
	}

	_, err := service.Create(context.Background(), "actor", input)
	assert.ErrorIs(t, err, auth.ErrNotFound)
}

func TestDocumentService_GetByID_Success(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo, testDocumentAssets(t))

	input := CreateDocumentInput{
		AssetID:    "asset-id-123",
		Name:       "doc.pdf",
		MimeType:   "application/pdf",
		SizeBytes:  1024,
		StorageKey: "storage/key/doc.pdf",
	}

	created, err := seedDocument(mockRepo, input)
	require.NoError(t, err)

	doc, err := service.GetByID(context.Background(), "actor", created.ID())
	require.NoError(t, err)
	assert.Equal(t, created.ID(), doc.ID())
	assert.Equal(t, "doc.pdf", doc.Name())
}

func TestDocumentService_GetByID_NotFound(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo, testDocumentAssets(t))

	_, err := service.GetByID(context.Background(), "actor", "non-existent-id")
	assert.ErrorIs(t, err, auth.ErrNotFound)
}

func TestDocumentService_ListByAssetID_Success(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo, testDocumentAssets(t))

	// Create 2 documents for asset-1 and 1 for asset-2
	docs := []CreateDocumentInput{
		{AssetID: "asset-1", Name: "doc1.pdf", MimeType: "application/pdf", SizeBytes: 100, StorageKey: "key1"},
		{AssetID: "asset-1", Name: "doc2.pdf", MimeType: "application/pdf", SizeBytes: 200, StorageKey: "key2"},
		{AssetID: "asset-2", Name: "doc3.pdf", MimeType: "application/pdf", SizeBytes: 300, StorageKey: "key3"},
	}

	for _, d := range docs {
		_, err := seedDocument(mockRepo, d)
		require.NoError(t, err)
	}

	result, err := service.ListByAssetID(context.Background(), "actor", "asset-1")
	require.NoError(t, err)
	assert.Len(t, result, 2)

	result2, err := service.ListByAssetID(context.Background(), "actor", "asset-2")
	require.NoError(t, err)
	assert.Len(t, result2, 1)

	result3, err := service.ListByAssetID(context.Background(), "actor", "non-existent")
	require.ErrorIs(t, err, auth.ErrNotFound)
	assert.Empty(t, result3)
}

func TestDocumentService_Delete_Success(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo, testDocumentAssets(t))

	input := CreateDocumentInput{
		AssetID:    "asset-id",
		Name:       "doc.pdf",
		MimeType:   "application/pdf",
		SizeBytes:  100,
		StorageKey: "storage/key/doc.pdf",
	}

	created, err := seedDocument(mockRepo, input)
	require.NoError(t, err)

	err = service.Delete(context.Background(), "actor", created.ID())
	assert.NoError(t, err)

	_, err = service.GetByID(context.Background(), "actor", created.ID())
	assert.ErrorIs(t, err, auth.ErrNotFound)
}

func TestDocumentService_Delete_NotFound(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	service := NewDocumentService(mockRepo, testDocumentAssets(t))

	err := service.Delete(context.Background(), "actor", "non-existent")
	assert.ErrorIs(t, err, auth.ErrNotFound)
}

func TestDocumentService_Create_RepoError(t *testing.T) {
	mockRepo := newMockDocumentRepository()
	mockRepo.err = errors.New("db error")
	service := NewDocumentService(mockRepo, testDocumentAssets(t))

	input := CreateDocumentInput{
		AssetID:    "asset-id",
		Name:       "doc.pdf",
		MimeType:   "application/pdf",
		SizeBytes:  100,
		StorageKey: "storage/key/doc.pdf",
	}

	_, err := service.Create(context.Background(), "actor", input)
	assert.ErrorIs(t, err, domain.ErrDocumentUploadRequired)
}

func seedDocument(repo *mockDocumentRepository, input CreateDocumentInput) (*domain.Document, error) {
	now := time.Now().UTC()
	doc, err := domain.NewDocument(uuid.NewString(), input.AssetID, input.Name, input.MimeType, input.StorageKey, input.DownloadURL, input.SizeBytes, nil, now, now)
	if err != nil {
		return nil, err
	}
	return doc, repo.Save(context.Background(), doc)
}
func (m *mockDocumentRepository) CommitDocument(ctx context.Context, d *domain.Document, actor, action string) error {
	return m.Save(ctx, d)
}
func (m *mockDocumentRepository) DeleteDocument(ctx context.Context, id, actor string) error {
	return m.Delete(ctx, id)
}
func (m *mockDocumentRepository) ListDocuments(ctx context.Context, id string, limit, offset int) ([]*domain.Document, int64, error) {
	docs, err := m.FindByAssetID(ctx, id)
	return docs, int64(len(docs)), err
}
