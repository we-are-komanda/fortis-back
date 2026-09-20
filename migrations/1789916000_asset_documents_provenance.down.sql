DROP TRIGGER IF EXISTS document_identity_immutable ON defense_asset_documents;
DROP FUNCTION IF EXISTS fortis_document_identity_immutable();
DROP INDEX IF EXISTS defense_asset_documents_active_idx;
ALTER TABLE defense_asset_documents DROP COLUMN deleted_at,DROP COLUMN commercial,DROP COLUMN status,DROP COLUMN checksum,DROP COLUMN revision;
