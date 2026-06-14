CREATE TYPE post_tag AS ENUM ('event', 'competition', 'news', 'announcement');

CREATE TABLE posts (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    tag post_tag NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    wca_id VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100),
    email VARCHAR(100),
    roles JSONB NOT NULL DEFAULT '[]'::jsonb,
    competition_count INTEGER DEFAULT 0,
    delegated_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE records (
    region VARCHAR(50) NOT NULL,
    event_id VARCHAR(20) NOT NULL,
    single INTEGER,
    average INTEGER,
    PRIMARY KEY (region, event_id)
);