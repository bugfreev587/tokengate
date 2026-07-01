ALTER TABLE accounts
  ADD COLUMN IF NOT EXISTS owner_user_id BIGINT NULL REFERENCES users(id) ON DELETE CASCADE,
  ADD COLUMN IF NOT EXISTS credentials_encrypted BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE groups
  ADD COLUMN IF NOT EXISTS owner_user_id BIGINT NULL REFERENCES users(id) ON DELETE CASCADE,
  ADD COLUMN IF NOT EXISTS capacity_source VARCHAR(32) NOT NULL DEFAULT 'tokengate';

ALTER TABLE usage_logs
  ADD COLUMN IF NOT EXISTS capacity_source VARCHAR(32) NOT NULL DEFAULT 'tokengate';

CREATE INDEX IF NOT EXISTS idx_accounts_owner_user_id ON accounts(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_groups_owner_capacity_source ON groups(owner_user_id, capacity_source);
CREATE INDEX IF NOT EXISTS idx_usage_logs_capacity_source ON usage_logs(capacity_source);
