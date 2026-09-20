-- Identifiers are historical facts. No user/project cascade may erase the journal.
CREATE TABLE audit_events (
    event_id UUID PRIMARY KEY,
    actor_id UUID NOT NULL,
    enterprise_id UUID,
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    action TEXT NOT NULL,
    previous_version INTEGER,
    new_version INTEGER,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    request_id UUID NOT NULL
);
CREATE INDEX audit_events_entity_idx ON audit_events(entity_type, entity_id, occurred_at);
CREATE INDEX audit_events_enterprise_idx ON audit_events(enterprise_id, occurred_at);

CREATE FUNCTION fortis_audit_event_immutable() RETURNS TRIGGER
LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'audit events are append only'; END; $$;
CREATE TRIGGER audit_events_immutable BEFORE UPDATE OR DELETE ON audit_events
FOR EACH ROW EXECUTE FUNCTION fortis_audit_event_immutable();
