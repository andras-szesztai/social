ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role_id SERIAL REFERENCES roles(id) DEFAULT 1;

UPDATE users
    SET role_id = (
        SELECT id
        FROM roles
        WHERE name = 'user'
    )
    WHERE role_id IS NULL;

ALTER TABLE users
    ALTER COLUMN role_id SET NOT NULL;

ALTER TABLE users
    ALTER COLUMN role_id DROP DEFAULT;