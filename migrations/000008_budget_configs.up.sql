CREATE TABLE budget_configs (
    project_id  UUID PRIMARY KEY REFERENCES defense_projects(id) ON DELETE CASCADE,
    config_data JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budget_configs_project_id ON budget_configs(project_id);
