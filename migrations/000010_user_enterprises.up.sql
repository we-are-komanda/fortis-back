CREATE TABLE user_enterprises (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    enterprise_id UUID NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, enterprise_id)
);

CREATE INDEX idx_user_enterprises_user_id ON user_enterprises(user_id);
CREATE INDEX idx_user_enterprises_enterprise_id ON user_enterprises(enterprise_id);
