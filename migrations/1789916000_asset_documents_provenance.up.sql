ALTER TABLE defense_asset_documents
 ADD COLUMN revision TEXT NOT NULL DEFAULT '1',
 ADD COLUMN checksum TEXT,
 ADD COLUMN status TEXT NOT NULL DEFAULT 'legacy_unavailable' CHECK(status IN ('legacy_unavailable','quarantined','ready','rejected')),
 ADD COLUMN commercial BOOLEAN NOT NULL DEFAULT TRUE,
 ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX defense_asset_documents_active_idx ON defense_asset_documents(asset_id,created_at) WHERE deleted_at IS NULL;
-- File identity cannot change after upload. A new file always has a new UUID.
CREATE FUNCTION fortis_document_identity_immutable() RETURNS TRIGGER LANGUAGE plpgsql AS $$ BEGIN
 IF ROW(NEW.id,NEW.asset_id,NEW.name,NEW.mime_type,NEW.size_bytes,NEW.storage_key,NEW.checksum,NEW.revision,NEW.owner_id,NEW.created_at,NEW.commercial)
 IS DISTINCT FROM ROW(OLD.id,OLD.asset_id,OLD.name,OLD.mime_type,OLD.size_bytes,OLD.storage_key,OLD.checksum,OLD.revision,OLD.owner_id,OLD.created_at,OLD.commercial)
 THEN RAISE EXCEPTION 'document identity is immutable'; END IF;
 RETURN NEW;
END; $$;
CREATE TRIGGER document_identity_immutable BEFORE UPDATE ON defense_asset_documents FOR EACH ROW EXECUTE FUNCTION fortis_document_identity_immutable();
