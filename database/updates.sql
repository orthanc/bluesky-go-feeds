-- name: SaveCursor :exec
insert into
  sub_state ("service", "cursor")
values
  (?, ?) on conflict do
update
set
  "cursor" = excluded."cursor";

-- name: SavePost :exec
insert into
  post (
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
  )
values
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on conflict do nothing;

-- name: IncrementPostDirectReply :exec
update post
set
  "directReplyCount" = "directReplyCount" + 1,
  "replyCount" = "replyCount" + 1,
  "interactionCount" = "interactionCount" + 1
where
  "uri" = ?;

-- name: IncrementPostIndirectReply :exec
update post
set
  "replyCount" = "replyCount" + 1,
  "interactionCount" = "interactionCount" + 1
where
  "uri" = ?;

-- name: IncrementPostLike :exec
update post
set
  "likeCount" = "likeCount" + 1,
  "interactionCount" = "interactionCount" + 1
where
  "uri" = ?;

-- name: SaveUserInteraction :exec
insert into
  userInteraction (
    "interactionUri",
    "userDid",
    "type",
    "authorDid",
    "postUri",
    "indexedAt"
  )
values
  (?, ?, ?, ?, ?, ?);

-- name: SaveInteractionWithUser :exec
insert into
  interactionWithUser (
    "interactionUri",
    "userDid",
    "type",
    "interactionAuthorDid",
    "postUri",
    "indexedAt"
  )
values
  (?, ?, ?, ?, ?, ?);

-- name: SaveUser :exec
insert into
  user ("userDid", "lastSeen")
values
  (?, ?) on conflict do
update
set
  "lastSeen" = excluded."lastSeen";

-- name: UpdateUserLastSeen :execrows
update user
set
  "lastSeen" = ?
where
  "userDid" = ?;

-- name: SaveSession :exec
insert into
  session (
    "userDid",
    "startedAt",
    "postsSince",
    "lastSeen",
    "accessCount",
    "algo"
  )
values
  (?, ?, ?, ?, ?, ?);

-- name: UpdateSessionLastSeen :exec
update session
set
  "lastSeen" = ?,
  "accessCount" = "accessCount" + 1
where
  "sessionId" = ?;

-- name: SaveFollowing :exec
insert into
  following (
    "uri",
    "followedBy",
    "following",
    "userInteractionRatio"
  )
values
  (?, ?, ?, ?) on conflict do nothing;

-- name: DeleteFollowing :exec
delete from following
where
  uri = ?;

-- name: SaveAuthor :exec
insert into
  author (
    "did",
    "medianDirectReplyCount",
    "medianInteractionCount",
    "medianLikeCount",
    "medianReplyCount"
  )
values
  (?, ?, ?, ?, ?) on conflict do nothing;