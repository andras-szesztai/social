CREATE TABLE IF NOT EXISTS user_invitations (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    token bytea PRIMARY KEY,
);