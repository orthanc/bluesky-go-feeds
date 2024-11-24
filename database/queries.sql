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
  author
order by
  hex (randomblob (16));

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

-- name: GetLikeFollowData :one
select
  (
    select
      count(*)
    from
      user as postUser
    where
      postUser."userDid" = sqlc.arg (post_author)
  ) as post_by_user,
  (
    select
      count(*)
    from
      author
    where
      did = sqlc.arg (post_author)
  ) as post_by_author,
  (
    select
      count(*)
    from
      user as likeUser
    where
      likeUser."userDid" = sqlc.arg (like_author)
  ) as like_by_user;