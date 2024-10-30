CREATE TABLE users_missions (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    mission_id INTEGER NOT NULL REFERENCES missions (id),
    progress INTEGER NOT NULL DEFAULT 0,
    claimed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
);