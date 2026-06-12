CREATE TABLE defense_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_defense_projects_updated_at ON defense_projects(updated_at);
