//go:build unit

package ui

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/fortis/backend/internal/modules/defense_asset/application"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
)

// mockDocumentService — мок сервиса для тестирования контроллера документов.
type mockDocumentService struct {
	createFn       func(ctx context.Context, input application.CreateDocumentInput) (*domain.Document, error)
	getByIDFn      func(ctx context.Context, id string) (*domain.Document, error)
	listByAssetIDFn func(ctx context.Context, assetID string) ([]*domain.Document, error)
	deleteFn       func(ctx context.Context, id string) error
}

func (m *mockDocumentService) Create(ctx context.Context, input application.CreateDocumentInput) (*domain.Document, error) {
	return m.createFn(ctx, input)
}

func (m *mockDocumentService) GetByID(ctx context.Context, id string) (*domain.Document, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockDocumentService) ListByAssetID(ctx context.Context, assetID string) ([]*domain.Document, error) {
	return m.listByAssetIDFn(ctx, assetID)
}

func (m *mockDocumentService) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

// testDocument возвращает тестовый доменный объект Document.
func testDocument(id, assetID, name string) *domain.Document {
	now := time.Now().UTC()
	mimeType := "application/pdf"
	storageKey := "storage/key/" + name
	doc, _ := domain.NewDocument(id, assetID, name, mimeType, storageKey, "", 1024, nil, now, now)
	return doc
}

func TestDocumentCreate_Success(t *testing.T) {
	svc := &mockDocumentService{
		createFn: func(ctx context.Context, input application.CreateDocumentInput) (*domain.Document, error) {
			return testDocument("doc-123", input.AssetID, input.Name), nil
		},
	}

	controller := NewDocumentController(svc)

	body := `{"assetId":"asset-123","name":"doc.pdf","storageKey":"storage/key/doc.pdf"}`
	req := fasthttp.AcquireRequest()
	req.SetBody([]byte(body))
	req.Header.SetContentType("application/json")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Create(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusCreated {
		t.Errorf("expected status %d, got %d", fasthttp.StatusCreated, ctx.Response.StatusCode())
	}

	var resp AssetDocumentDTO
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Name != "doc.pdf" {
		t.Errorf("expected name 'doc.pdf', got %q", resp.Name)
	}
	if resp.AssetID != "asset-123" {
		t.Errorf("expected assetId 'asset-123', got %q", resp.AssetID)
	}
}

func TestDocumentCreate_MissingAssetID(t *testing.T) {
	controller := NewDocumentController(&mockDocumentService{})

	body := `{"name":"doc.pdf","storageKey":"storage/key/doc.pdf"}`
	req := fasthttp.AcquireRequest()
	req.SetBody([]byte(body))
	req.Header.SetContentType("application/json")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Create(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}

func TestDocumentCreate_MissingName(t *testing.T) {
	controller := NewDocumentController(&mockDocumentService{})

	body := `{"assetId":"asset-123","storageKey":"storage/key/doc.pdf"}`
	req := fasthttp.AcquireRequest()
	req.SetBody([]byte(body))
	req.Header.SetContentType("application/json")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Create(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}

func TestDocumentCreate_MissingStorageKey(t *testing.T) {
	controller := NewDocumentController(&mockDocumentService{})

	body := `{"assetId":"asset-123","name":"doc.pdf"}`
	req := fasthttp.AcquireRequest()
	req.SetBody([]byte(body))
	req.Header.SetContentType("application/json")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Create(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}

func TestDocumentCreate_InvalidJSON(t *testing.T) {
	controller := NewDocumentController(&mockDocumentService{})

	body := `{invalid json}`
	req := fasthttp.AcquireRequest()
	req.SetBody([]byte(body))
	req.Header.SetContentType("application/json")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Create(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}

func TestDocumentList_Success(t *testing.T) {
	svc := &mockDocumentService{
		listByAssetIDFn: func(ctx context.Context, assetID string) ([]*domain.Document, error) {
			return []*domain.Document{
				testDocument("doc-1", assetID, "doc1.pdf"),
				testDocument("doc-2", assetID, "doc2.pdf"),
			}, nil
		},
	}

	controller := NewDocumentController(svc)

	req := fasthttp.AcquireRequest()
	req.URI().SetQueryString("assetId=asset-123")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.List(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status %d, got %d", fasthttp.StatusOK, ctx.Response.StatusCode())
	}

	var resp AssetDocumentListResponse
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.TotalItems != 2 {
		t.Errorf("expected totalItems 2, got %d", resp.TotalItems)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
}

func TestDocumentList_MissingAssetID(t *testing.T) {
	controller := NewDocumentController(&mockDocumentService{})

	req := fasthttp.AcquireRequest()
	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.List(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}

func TestDocumentGet_Success(t *testing.T) {
	svc := &mockDocumentService{
		getByIDFn: func(ctx context.Context, id string) (*domain.Document, error) {
			return testDocument(id, "asset-123", "doc.pdf"), nil
		},
	}

	controller := NewDocumentController(svc)

	req := fasthttp.AcquireRequest()
	req.URI().SetQueryString("id=doc-123")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Get(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status %d, got %d", fasthttp.StatusOK, ctx.Response.StatusCode())
	}

	var resp AssetDocumentDTO
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Name != "doc.pdf" {
		t.Errorf("expected name 'doc.pdf', got %q", resp.Name)
	}
}

func TestDocumentGet_NotFound(t *testing.T) {
	svc := &mockDocumentService{
		getByIDFn: func(ctx context.Context, id string) (*domain.Document, error) {
			return nil, domain.ErrDocumentNotFound
		},
	}

	controller := NewDocumentController(svc)

	req := fasthttp.AcquireRequest()
	req.URI().SetQueryString("id=non-existent")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Get(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("expected status %d, got %d", fasthttp.StatusNotFound, ctx.Response.StatusCode())
	}
}

func TestDocumentGet_MissingID(t *testing.T) {
	controller := NewDocumentController(&mockDocumentService{})

	req := fasthttp.AcquireRequest()
	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Get(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}

func TestDocumentDownload_Success(t *testing.T) {
	svc := &mockDocumentService{
		getByIDFn: func(ctx context.Context, id string) (*domain.Document, error) {
			doc := testDocument(id, "asset-123", "doc.pdf")
			doc.SetDownloadURL("https://storage.example.com/documents/doc.pdf")
			return doc, nil
		},
	}

	controller := NewDocumentController(svc)

	req := fasthttp.AcquireRequest()
	req.URI().SetQueryString("id=doc-123")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Download(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status %d, got %d", fasthttp.StatusOK, ctx.Response.StatusCode())
	}

	var resp AssetDocumentDTO
	if err := json.Unmarshal(ctx.Response.Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.DownloadURL != "https://storage.example.com/documents/doc.pdf" {
		t.Errorf("expected downloadUrl to be set, got %q", resp.DownloadURL)
	}
}

func TestDocumentDownload_MissingID(t *testing.T) {
	controller := NewDocumentController(&mockDocumentService{})

	req := fasthttp.AcquireRequest()
	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Download(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}

func TestDocumentDelete_Success(t *testing.T) {
	svc := &mockDocumentService{
		deleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}

	controller := NewDocumentController(svc)

	req := fasthttp.AcquireRequest()
	req.URI().SetQueryString("id=doc-123")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Delete(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("expected status %d, got %d", fasthttp.StatusOK, ctx.Response.StatusCode())
	}
}

func TestDocumentDelete_NotFound(t *testing.T) {
	svc := &mockDocumentService{
		deleteFn: func(ctx context.Context, id string) error {
			return domain.ErrDocumentNotFound
		},
	}

	controller := NewDocumentController(svc)

	req := fasthttp.AcquireRequest()
	req.URI().SetQueryString("id=non-existent")

	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Delete(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("expected status %d, got %d", fasthttp.StatusNotFound, ctx.Response.StatusCode())
	}
}

func TestDocumentDelete_MissingID(t *testing.T) {
	controller := NewDocumentController(&mockDocumentService{})

	req := fasthttp.AcquireRequest()
	var ctx fasthttp.RequestCtx
	ctx.Init(req, nil, nil)

	controller.Delete(&ctx)

	if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
	}
}
