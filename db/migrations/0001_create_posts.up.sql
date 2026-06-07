CREATE TYPE post_tag AS ENUM ('event', 'competition', 'news', 'announcement');

CREATE TABLE posts (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    tag post_tag NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);
