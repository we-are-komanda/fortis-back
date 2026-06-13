CREATE TABLE defense_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enterprise_id UUID REFERENCES enterprises(id) ON DELETE CASCADE,
    asset_data JSONB NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_defense_assets_enterprise_id ON defense_assets(enterprise_id);
CREATE INDEX idx_defense_assets_is_public ON defense_assets(is_public);
CREATE INDEX idx_defense_assets_updated_at ON defense_assets(updated_at);
