-- Migration: create users table
-- users depends on user_roles, departments, sectors

CREATE TABLE IF NOT EXISTS users (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(100) NOT NULL UNIQUE,
    email           VARCHAR(255) NOT NULL UNIQUE,
    phone           VARCHAR(30)  NOT NULL DEFAULT '',
    firstname       VARCHAR(100) NOT NULL DEFAULT '',
    lastname        VARCHAR(100) NOT NULL DEFAULT '',
    nickname        VARCHAR(100) NOT NULL DEFAULT '',
    password        TEXT        NOT NULL,
    role_id         UUID        REFERENCES user_roles(id) ON DELETE SET NULL,
    department_id   UUID        REFERENCES departments(id) ON DELETE SET NULL,
    sector_id       UUID        REFERENCES sectors(id) ON DELETE SET NULL,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    profile_picture TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_role_id       ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_users_department_id ON users(department_id);
CREATE INDEX IF NOT EXISTS idx_users_sector_id     ON users(sector_id);

COMMENT ON TABLE users IS 'Stores all system users with their roles and organizational units';
