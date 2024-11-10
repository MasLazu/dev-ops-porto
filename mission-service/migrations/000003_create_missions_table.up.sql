CREATE TABLE missions (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    image_path VARCHAR(255) NOT NULL,
    event_encreasor_id INTEGER NOT NULL REFERENCES events (id),
    event_decreasor_id INTEGER REFERENCES events (id),
    goal INTEGER NOT NULL,
    reward INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);