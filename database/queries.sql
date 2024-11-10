-- name: GetCursor :many
select
  *
from
  sub_state
where
  "service" = ?
limit
  1;

-- name: ListPostInteractionsForAuthor :many
select
  "directReplyCount",
  "interactionCount",
  "likeCount",
  "replyCount"
from
  post
where
  "author" = ?;

-- name: GetPostDates :one
select
  "indexedAt",
  created_at
from
  post
where
  uri = ?;

-- name: ListAllUsers :many
select
  "userDid"
from
  user;

-- name: ListUsersNotSeenSince :many
select
  "userDid"
from
  user
where
  "lastSeen" < sqlc.arg ('purgeBefore');

-- name: ListAllAuthors :many
select
  "did"
from
  author;

-- name: ListAllFollowing :many
select
  *
from
  following;

-- name: ListAllFollowers :many
select
  *
from
  follower;

-- name: GetLastSession :many
select
  *
from
  session
where
  "userDid" = ?
  and (
    "algo" = ?
    OR "algo" is null
  )
order by
  "lastSeen" desc
limit
  1;

-- name: ListFollowerLastRecordedBefore :many
select
  following,
  followed_by,
  last_recorded
from
  follower
where
  following = ?
  and last_recorded < ?;

-- name: ListFollowingLastRecordedBefore :many
select
  uri,
  last_recorded
from
  following
where
  "followedBy" = ?
  and (
    last_recorded is NULL
    or last_recorded < ?
  );