package domain

import (
	"context"
	"errors"
)

const MaxDocumentBytes = 10 * 1024 * 1024

var ErrDocumentUnavailable = errors.New("document storage or scanner unavailable")
var ErrDocumentRejected = errors.New("document rejected by scanner")
var ErrDocumentTooLarge = errors.New("document exceeds 10 MiB")
var ErrDocumentMediaType = errors.New("unsupported document content type")
var ErrDocumentUploadRequired = errors.New("file upload required")
var ErrDocumentNotReady = errors.New("document is not available for download")

type DocumentStorage interface {
	Ready() bool
	Put(context.Context, string, []byte) (string, error)
	Read(context.Context, string) ([]byte, error)
	Remove(context.Context, string) error
}
type DocumentScanner interface {
	Ready() bool
	Scan(context.Context, string) error
}
type DocumentWriteRepository interface {
	CommitDocument(context.Context, *Document, string, string) error
	DeleteDocument(context.Context, string, string) error
	ListDocuments(context.Context, string, int, int) ([]*Document, int64, error)
}
