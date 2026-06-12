ALTER TABLE defense_projects ADD COLUMN name VARCHAR(512) NOT NULL DEFAULT '';
ALTER TABLE defense_projects ADD COLUMN enterprise_id UUID REFERENCES enterprises(id) ON DELETE CASCADE;
CREATE INDEX idx_defense_projects_enterprise_id ON defense_projects(enterprise_id);
