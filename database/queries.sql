-- name: GetCursor :many
select
  *
from
  sub_state
where
  "service" = ?
limit
  1;

-- name: ListPosts :many
SELECT
  "author",
  "directReplyCount",
  "indexedAt",
  "interactionCount",
  "likeCount",
  "replyCount",
  "uri",
  "replyParent",
  "replyParentAuthor",
  "replyRoot",
  "replyRootAuthor"
FROM
  post;

-- name: ListAllUsers :many
select
  "userDid"
FROM
  user;

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