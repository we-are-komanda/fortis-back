DROP TABLE IF EXISTS project_cost_projections;
DROP FUNCTION IF EXISTS fortis_cost_projection_immutable();
ALTER TABLE project_revisions DROP COLUMN IF EXISTS cost_calculation_version;
