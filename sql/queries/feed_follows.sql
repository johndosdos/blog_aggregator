-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO
        feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES
        ($1, $2, $3, $4, $5)
    RETURNING *
) SELECT
    inserted_feed_follow.*,
    users.name,
    feeds.name
FROM 
    inserted_feed_follow
INNER JOIN users ON users.id = inserted_feed_follow.user_id
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id;

-- name: GetFeedFollowsForUser :many
SELECT
    feed_follows.*,
    users.name,
    feeds.name
FROM feed_follows
INNER JOIN users ON users.id = feed_follows.user_id
INNER JOIN feeds ON feeds.id = feed_follows.feed_id
WHERE users.name = $1;

-- name: DeleteUsersFeedFollows :exec
DELETE FROM feed_follows;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE
    feed_follows.user_id = $1 AND feed_follows.feed_id = (
        SELECT id 
        FROM feeds
        WHERE url = $2
    );