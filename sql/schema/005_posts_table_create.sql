-- +goose Up
CREATE TABLE posts(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    description TEXT ,
    published_at TIMESTAMP,
    feed_id UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;


/*
A post is a single entry from a feed. It should have:
id - a unique identifier for the post
created_at - the time the record was created
updated_at - the time the record was last updated
title - the title of the post
url - the URL of the post (this should be unique)
description - the description of the post
published_at - the time the post was published
feed_id - the ID of the feed that the post came from
Some of these fields can probably be null, others you might want to be more strict about - it's up to you.
*/