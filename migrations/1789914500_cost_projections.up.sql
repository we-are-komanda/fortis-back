ALTER TABLE project_revisions ADD COLUMN cost_calculation_version TEXT NOT NULL DEFAULT 'cost-rub-v1';

CREATE TABLE project_cost_projections (
    project_id UUID NOT NULL,
    project_version INTEGER NOT NULL,
    calculation_version TEXT NOT NULL,
    projection_json JSONB NOT NULL,
    known_subtotal_minor NUMERIC NOT NULL CHECK (known_subtotal_minor >= 0),
    total_minor NUMERIC CHECK (total_minor >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, project_version, calculation_version),
    FOREIGN KEY (project_id, project_version) REFERENCES project_revisions(project_id, version) ON DELETE CASCADE
);

CREATE FUNCTION fortis_cost_projection_immutable() RETURNS TRIGGER
LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'cost projections are immutable'; END; $$;
CREATE TRIGGER project_cost_projections_immutable BEFORE UPDATE ON project_cost_projections
FOR EACH ROW EXECUTE FUNCTION fortis_cost_projection_immutable();
