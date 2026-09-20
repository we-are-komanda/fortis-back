CREATE TABLE project_revisions (
    project_id UUID NOT NULL REFERENCES defense_projects(id) ON DELETE CASCADE,
    version INTEGER NOT NULL CHECK (version > 0),
    snapshot_json JSONB NOT NULL,
    snapshot_digest TEXT NOT NULL,
    -- Historical actor identifier, not a live ownership reference. Account deletion
    -- must neither rewrite an immutable revision nor be blocked by its history.
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, version)
);

CREATE TABLE project_idempotency (
    actor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    operation TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    payload_digest TEXT NOT NULL,
    project_id UUID REFERENCES defense_projects(id) ON DELETE SET NULL,
    project_version INTEGER NOT NULL CHECK (project_version > 0),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (actor_id, operation, idempotency_key)
);
CREATE INDEX project_idempotency_expires_at_idx ON project_idempotency(expires_at);

-- The canonical form matches the server: recursive lexical object keys,
-- preserved array order, UTF-8, no insignificant whitespace or HTML escaping.
CREATE FUNCTION fortis_project_canonical_json(value JSONB) RETURNS TEXT
LANGUAGE plpgsql IMMUTABLE STRICT AS $$
DECLARE result TEXT;
BEGIN
    CASE jsonb_typeof(value)
    WHEN 'object' THEN
        SELECT '{' || COALESCE(string_agg(fortis_project_canonical_json(to_jsonb(e.key)) || ':' || fortis_project_canonical_json(e.value), ',' ORDER BY e.key COLLATE "C"), '') || '}'
        INTO result FROM jsonb_each(value) e;
    WHEN 'array' THEN
        SELECT '[' || COALESCE(string_agg(fortis_project_canonical_json(e.value), ',' ORDER BY e.ordinality), '') || ']'
        INTO result FROM jsonb_array_elements(value) WITH ORDINALITY e;
    ELSE result := replace(replace(value::text, chr(8232), '\u2028'), chr(8233), '\u2029');
    END CASE;
    RETURN result;
END;
$$;

-- Only the current known version is recoverable. Author and earlier states are unknown.
WITH snapshots AS (
    SELECT p.id, p.version, jsonb_build_object(
        'project', (p.project_data - 'requestId' - 'generatedAt') || jsonb_build_object(
            'projectId', p.id, 'name', p.name, 'enterpriseId', p.enterprise_id, 'version', p.version,
            'updatedAt', COALESCE(p.project_data->'updatedAt', to_jsonb(p.updated_at))),
        'budget', b.config_data,
        'inputDataVersions', '{}'::jsonb
    ) AS snapshot
    FROM defense_projects p LEFT JOIN budget_configs b ON b.project_id = p.id
)
INSERT INTO project_revisions (project_id, version, snapshot_json, snapshot_digest)
SELECT id, version, snapshot,
    encode(sha256(convert_to(fortis_project_canonical_json(snapshot), 'UTF8')), 'hex')
FROM snapshots;

CREATE FUNCTION fortis_project_revision_immutable() RETURNS TRIGGER
LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'project revisions are immutable'; END; $$;
CREATE TRIGGER project_revision_immutable BEFORE UPDATE ON project_revisions
FOR EACH ROW EXECUTE FUNCTION fortis_project_revision_immutable();
