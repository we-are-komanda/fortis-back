-- Anonymous cache entries have no account; rollback removes only replay cache entries.
DELETE FROM project_idempotency WHERE actor_id NOT IN (SELECT id FROM users);
ALTER TABLE project_idempotency ADD CONSTRAINT project_idempotency_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE CASCADE;
