DROP INDEX IF EXISTS idx_defense_projects_enterprise_id;
ALTER TABLE defense_projects DROP COLUMN IF EXISTS enterprise_id;
ALTER TABLE defense_projects DROP COLUMN IF EXISTS name;
