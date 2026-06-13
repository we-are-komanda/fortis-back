package domain

import (
	"fmt"
	"time"
)

// Document — сущность документа, прикреплённого к карточке средства защиты.
type Document struct {
	id          string
	assetID     string
	name        string
	mimeType    string
	sizeBytes   int64
	storageKey  string
	downloadURL string
	ownerID     *string
	createdAt   time.Time
	updatedAt   time.Time
}

// NewDocument создаёт новый Document с валидацией.
func NewDocument(
	id, assetID, name, mimeType, storageKey, downloadURL string,
	sizeBytes int64,
	ownerID *string,
	createdAt, updatedAt time.Time,
) (*Document, error) {
	if name == "" {
		return nil, ErrDocumentInvalidName
	}
	if assetID == "" {
		return nil, ErrDocumentInvalidAssetID
	}
	if sizeBytes < 0 {
		return nil, fmt.Errorf("%w: %d", ErrDocumentInvalidSizeBytes, sizeBytes)
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if storageKey == "" {
		return nil, fmt.Errorf("%w: storage_key must not be empty", ErrDocumentInvalidStorageKey)
	}

	return &Document{
		id:          id,
		assetID:     assetID,
		name:        name,
		mimeType:    mimeType,
		sizeBytes:   sizeBytes,
		storageKey:  storageKey,
		downloadURL: downloadURL,
		ownerID:     ownerID,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

// Getters.
func (d *Document) ID() string             { return d.id }
func (d *Document) AssetID() string        { return d.assetID }
func (d *Document) Name() string           { return d.name }
func (d *Document) MimeType() string       { return d.mimeType }
func (d *Document) SizeBytes() int64       { return d.sizeBytes }
func (d *Document) StorageKey() string     { return d.storageKey }
func (d *Document) DownloadURL() string    { return d.downloadURL }
func (d *Document) OwnerID() *string       { return d.ownerID }
func (d *Document) CreatedAt() time.Time   { return d.createdAt }
func (d *Document) UpdatedAt() time.Time   { return d.updatedAt }

// SetDownloadURL обновляет download URL документа.
func (d *Document) SetDownloadURL(url string) {
	d.downloadURL = url
	d.updatedAt = time.Now().UTC()
}

// SetUpdatedAt обновляет timestamp изменения.
func (d *Document) SetUpdatedAt(t time.Time) {
	d.updatedAt = t
}

// UpdateName обновляет название документа.
func (d *Document) UpdateName(name string) error {
	if name == "" {
		return ErrDocumentInvalidName
	}
	d.name = name
	d.updatedAt = time.Now().UTC()
	return nil
}
