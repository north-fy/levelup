DROP INDEX IF EXISTS idx_users_github_id;
CREATE UNIQUE INDEX idx_users_github_id ON users (github_id);
