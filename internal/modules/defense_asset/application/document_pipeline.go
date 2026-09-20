package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/fortis/backend/internal/auth"
	"github.com/fortis/backend/internal/modules/defense_asset/domain"
	"github.com/google/uuid"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type DocumentPipeline struct {
	Storage domain.DocumentStorage
	Scanner domain.DocumentScanner
}
type UploadDocumentInput struct {
	AssetID, Name, MimeType string
	Body                    []byte
	Commercial              bool
}

func validatedDocument(in UploadDocumentInput) (string, string, error) {
	if len(in.Body) > domain.MaxDocumentBytes {
		return "", "", domain.ErrDocumentTooLarge
	}
	if len(in.Body) == 0 {
		return "", "", domain.ErrDocumentMediaType
	}
	kind, _, err := mime.ParseMediaType(in.MimeType)
	if err != nil {
		return "", "", domain.ErrDocumentMediaType
	}
	actual, _, _ := mime.ParseMediaType(http.DetectContentType(in.Body))
	if kind != actual {
		return "", "", domain.ErrDocumentMediaType
	}
	switch kind {
	case "text/plain":
		if !utf8.Valid(in.Body) {
			return "", "", domain.ErrDocumentMediaType
		}
		for _, r := range string(in.Body) {
			if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
				return "", "", domain.ErrDocumentMediaType
			}
		}
	case "application/pdf":
		if !bytes.HasPrefix(in.Body, []byte("%PDF-")) || !bytes.Contains(in.Body, []byte("%%EOF")) {
			return "", "", domain.ErrDocumentMediaType
		}
	case "image/png", "image/jpeg":
		config, _, err := image.DecodeConfig(bytes.NewReader(in.Body))
		if err != nil || config.Width < 1 || config.Height < 1 || int64(config.Width)*int64(config.Height) > 40000000 {
			return "", "", domain.ErrDocumentMediaType
		}
	default:
		return "", "", domain.ErrDocumentMediaType
	}
	name := filepath.Base(strings.ReplaceAll(in.Name, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == "/" || utf8.RuneCountInString(name) > 200 {
		return "", "", domain.ErrDocumentInvalidName
	}
	return name, kind, nil
}
func (s *DocumentService) Upload(ctx context.Context, actor string, in UploadDocumentInput) (*domain.Document, error) {
	if _, err := s.assets.GetForMutation(ctx, actor, in.AssetID); err != nil {
		return nil, err
	}
	name, kind, err := validatedDocument(in)
	if err != nil {
		return nil, err
	}
	repo, ok := s.repo.(domain.DocumentWriteRepository)
	if !ok || s.pipeline.Storage == nil || !s.pipeline.Storage.Ready() || s.pipeline.Scanner == nil || !s.pipeline.Scanner.Ready() {
		return nil, domain.ErrDocumentUnavailable
	}
	id := uuid.NewString()
	path, err := s.pipeline.Storage.Put(ctx, id, in.Body)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(in.Body)
	now := time.Now().UTC()
	doc, err := domain.NewDocument(id, in.AssetID, name, kind, id, "", int64(len(in.Body)), &actor, now, now)
	if err != nil {
		s.pipeline.Storage.Remove(ctx, id)
		return nil, err
	}
	doc.SetFileMetadata("quarantined", "1", hex.EncodeToString(sum[:]), in.Commercial, nil)
	if err = repo.CommitDocument(ctx, doc, actor, "document.upload"); err != nil {
		s.pipeline.Storage.Remove(context.WithoutCancel(ctx), id)
		return nil, err
	}
	err = s.pipeline.Scanner.Scan(ctx, path)
	status, action := "ready", "document.ready"
	if err != nil {
		if !errors.Is(err, domain.ErrDocumentRejected) {
			return doc, domain.ErrDocumentUnavailable
		}
		status, action = "rejected", "document.reject"
	}
	doc.SetFileMetadata(status, "1", doc.Checksum(), in.Commercial, nil)
	if saveErr := repo.CommitDocument(ctx, doc, actor, action); saveErr != nil {
		return nil, saveErr
	}
	if err != nil {
		return doc, domain.ErrDocumentRejected
	}
	return doc, nil
}
func (s *DocumentService) Download(ctx context.Context, actor, id, parent string) (*domain.Document, []byte, error) {
	doc, err := s.GetByID(ctx, actor, id)
	if err != nil {
		return nil, nil, err
	}
	if parent != "" && parent != doc.AssetID() {
		return nil, nil, auth.ErrNotFound
	}
	if doc.Status() != "ready" || doc.Checksum() == "" {
		return nil, nil, domain.ErrDocumentNotReady
	}
	if s.pipeline.Storage == nil || !s.pipeline.Storage.Ready() {
		return nil, nil, domain.ErrDocumentUnavailable
	}
	body, err := s.pipeline.Storage.Read(ctx, doc.StorageKey())
	if err != nil {
		return nil, nil, err
	}
	sum := sha256.Sum256(body)
	if hex.EncodeToString(sum[:]) != doc.Checksum() || int64(len(body)) != doc.SizeBytes() {
		return nil, nil, domain.ErrDocumentUnavailable
	}
	return doc, body, nil
}
func (s *DocumentService) ListPage(ctx context.Context, actor, asset string, limit, offset int) ([]*domain.Document, int64, error) {
	if _, err := s.assets.GetByID(ctx, actor, asset); err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	if repo, ok := s.repo.(domain.DocumentWriteRepository); ok {
		return repo.ListDocuments(ctx, asset, limit, offset)
	}
	docs, err := s.repo.FindByAssetID(ctx, asset)
	if err != nil {
		return nil, 0, err
	}
	count := int64(len(docs))
	if offset >= len(docs) {
		return []*domain.Document{}, count, nil
	}
	end := offset + limit
	if end > len(docs) {
		end = len(docs)
	}
	return docs[offset:end], count, nil
}
