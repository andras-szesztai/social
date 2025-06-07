CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    level int NOT NULL DEFAULT 0,
    description TEXT,
);

INSERT INTO roles (name, level, description) VALUES ('admin', 20, 'Admin role');
INSERT INTO roles (name, level, description) VALUES ('moderator', 10, 'Moderator role');
INSERT INTO roles (name, level, description) VALUES ('user', 1, 'User role');


