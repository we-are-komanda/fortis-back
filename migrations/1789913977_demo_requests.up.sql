CREATE TABLE demo_requests (
 id UUID PRIMARY KEY,
 event_id UUID NOT NULL UNIQUE,
 name TEXT NOT NULL,
 organization TEXT NOT NULL,
 email TEXT NOT NULL,
 comment TEXT NOT NULL,
 consent_version TEXT NOT NULL,
 consent_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX demo_requests_created_idx ON demo_requests(consent_at DESC, id);
CREATE TABLE demo_request_outbox (
 id UUID PRIMARY KEY,
 request_id UUID NOT NULL UNIQUE REFERENCES demo_requests(id) ON DELETE CASCADE,
 status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','sent','failed')),
 attempts INTEGER NOT NULL DEFAULT 0 CHECK(attempts >= 0),
 next_attempt_at TIMESTAMPTZ NOT NULL,
 lease_token TEXT NOT NULL DEFAULT '',
 lease_until TIMESTAMPTZ,
 last_error TEXT NOT NULL DEFAULT ''
);
CREATE INDEX demo_request_outbox_due_idx ON demo_request_outbox(next_attempt_at,lease_until) WHERE status='pending';
CREATE TABLE demo_request_receipts (
 key TEXT PRIMARY KEY,
 digest TEXT NOT NULL DEFAULT '',
 request_id UUID REFERENCES demo_requests(id) ON DELETE CASCADE,
 expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX demo_request_receipts_expiry_idx ON demo_request_receipts(expires_at);
CREATE TABLE demo_request_rates (
 key TEXT PRIMARY KEY,
 timestamps JSONB NOT NULL DEFAULT '[]'
);
