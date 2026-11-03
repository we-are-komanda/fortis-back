CREATE TABLE issued_reports (
 id UUID PRIMARY KEY,
 project_id UUID NOT NULL,
 project_version INTEGER NOT NULL CHECK (project_version > 0),
 enterprise_id UUID NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
 report_schema_version TEXT NOT NULL,
 hide_cost BOOLEAN NOT NULL,
 map_view TEXT NOT NULL CHECK (map_view IN ('all','saved')),
 generated_at TIMESTAMPTZ NOT NULL,
 created_by UUID NOT NULL,
 payload_json JSONB NOT NULL,
 FOREIGN KEY (project_id,project_version) REFERENCES project_revisions(project_id,version) ON DELETE CASCADE,
 UNIQUE(project_id,project_version,report_schema_version,hide_cost,map_view)
);
CREATE FUNCTION fortis_issued_report_immutable() RETURNS TRIGGER
LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'issued reports are immutable'; END; $$;
CREATE TRIGGER issued_report_immutable BEFORE UPDATE ON issued_reports
FOR EACH ROW EXECUTE FUNCTION fortis_issued_report_immutable();
