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
  1