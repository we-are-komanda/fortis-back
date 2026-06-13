CREATE TABLE defense_asset_documents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id    UUID NOT NULL REFERENCES defense_assets(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    mime_type   TEXT NOT NULL DEFAULT 'application/octet-stream',
    size_bytes  BIGINT NOT NULL DEFAULT 0,
    storage_key TEXT NOT NULL,
    download_url TEXT,
    owner_id    UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_defense_asset_documents_asset_id ON defense_asset_documents(asset_id);
