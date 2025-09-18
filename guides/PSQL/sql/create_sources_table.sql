CREATE TABLE IF NOT EXISTS sources (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    link TEXT NOT NULL UNIQUE,
    date TIMESTAMP NOT NULL,
    summary TEXT,
    importance_bool BOOLEAN,
    importance_reasoning TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed BOOLEAN DEFAULT FALSE,
    relevant_per_human_check TEXT DEFAULT 'maybe'
);
