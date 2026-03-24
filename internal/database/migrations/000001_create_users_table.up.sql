CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL PRIMARY KEY,
    username   VARCHAR   NOT NULL,
    password   VARCHAR   NOT NULL,
    email      VARCHAR   NOT NULL UNIQUE,
    full_name  VARCHAR,
    cpf        VARCHAR,
    birth_date DATE,
    type       VARCHAR,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
