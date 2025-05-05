-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;


-- name: GetUserByName :one
SELECT * FROM users WHERE name = $1;

-- name: Reset :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT * FROM users;

-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    $4
) RETURNING *;

-- name: GetFeeds :many
SELECT feeds.name as rss_name, feeds.url as rss_url, users.name as user_name FROM feeds
INNER JOIN users ON feeds.user_id = users.id;

-- name: CreateFeedFollow :many
WITH inserted_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, feed_id, user_id)
    VALUES (
        $1,
        NOW(),
        NOW(),
        $2,
        $3
    ) RETURNING *
) SELECT 
inserted_follow.*,
feeds.name as feed_name,
feeds.url as feed_url,
users.name as user_name
FROM  inserted_follow
INNER JOIN users ON users.id = inserted_follow.user_id
INNER JOIN feeds ON feeds.id = inserted_follow.feed_id;

-- name: GetFeedByURL :one
SELECT * FROM feeds WHERE url = $1;

-- name: GetFeedFollowForUser :many
SELECT 
    users.name AS user_name,
    feeds.name AS feed_name,
    feeds.url AS feed_url,
    feeds.last_fetched_at as last_fetched_at
FROM feed_follows
INNER JOIN feeds ON feeds.id = feed_follows.feed_id
INNER JOIN users ON users.id = feed_follows.user_id
WHERE feed_follows.user_id = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows USING feeds
WHERE feeds.url = $2 AND feed_follows.feed_id = feeds.id AND feed_follows.user_id = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT id, url FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST;

-- name: CreatePost :exec
INSERT INTO posts(id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    $4,
    $5,
    $6
);

-- name: GetPostsForUser :many
SELECT * FROM posts;
