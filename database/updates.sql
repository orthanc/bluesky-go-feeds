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

-- name: IncrementPostRepost :exec
update post
set
  repost_count = repost_count + 1,
  "interactionCount" = "interactionCount" + 1
where
  "uri" = ?;

-- name: DeletePostsBefore :execrows
delete from post
where
  "indexedAt" < ?;

-- name: DeleteRepostsBefore :execrows
delete from repost
where
  "indexedAt" < ?;

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
  (?, ?, ?, ?, ?, ?) on conflict do nothing;

-- name: DeleteUserInteractionsBefore :execrows
delete from userInteraction
where
  "indexedAt" < ?;

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
  (?, ?, ?, ?, ?, ?) on conflict do nothing;

-- name: DeleteInteractionWithUsersBefore :execrows
delete from interactionWithUser
where
  "indexedAt" < ?;

-- name: SaveUser :exec
insert into
  user ("userDid", "lastSeen", "lastSynced")
values
  (?, ?, ?) on conflict do
update
set
  "lastSeen" = excluded."lastSeen",
  "lastSynced" = excluded."lastSynced";

-- name: UpdateUserLastSeen :many
update user
set
  "lastSeen" = ?
where
  "userDid" = ? returning *;

-- name: DeleteUserWhenNotSeen :execrows
delete from user
where
  "userDid" = ?
  and "lastSeen" < sqlc.arg ('purgeBefore');

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

-- name: DeleteSessionsBefore :execrows
delete from session
where
  "lastSeen" < ?;

-- name: DeletePostInteractedByFollowedBefore :execrows
delete from post_interacted_by_followed
where indexed_at < ?;

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

-- name: UpdateAuthorMedians :exec
update author
set
  "medianReplyCount" = ?,
  "medianDirectReplyCount" = ?,
  "medianLikeCount" = ?,
  "medianInteractionCount" = ?
where
  "did" = ?;

-- name: DeleteAuthorsByDid :execrows
delete from author
where
  "did" in (sqlc.slice ('dids'));

-- name: SavePostDirectRepliedToByFollowing :exec
insert into
  post_interacted_by_followed (
    user,
    uri,
    author,
    indexed_at,
    followed_reply_count,
    followed_direct_reply_count,
    followed_interaction_count
  )
values
  (?, ?, ?, ?, 1, 1, 1) on conflict do
update
set
  followed_reply_count = followed_reply_count + 1,
  followed_direct_reply_count = followed_direct_reply_count + 1,
  followed_interaction_count = followed_interaction_count + 1;

-- name: SavePostRepliedToByFollowing :exec
insert into
  post_interacted_by_followed (
    user,
    uri,
    author,
    indexed_at,
    followed_reply_count,
    followed_interaction_count
  )
values
  (?, ?, ?, ?, 1, 1) on conflict do
update
set
  followed_reply_count = followed_reply_count + 1,
  followed_interaction_count = followed_interaction_count + 1;

-- name: SavePostLikedByFollowing :exec
insert into
  post_interacted_by_followed (
    user,
    uri,
    author,
    indexed_at,
    followed_like_count,
    followed_interaction_count
  )
values
  (?, ?, ?, ?, 1, 1) on conflict do
update
set
  followed_like_count = followed_like_count + 1,
  followed_interaction_count = followed_interaction_count + 1;

-- name: SavePostRepostedByFollowing :exec
insert into
  post_interacted_by_followed (
    user,
    uri,
    author,
    indexed_at,
    followed_repost_count,
    followed_interaction_count
  )
values
  (?, ?, ?, ?, 1, 1) on conflict do
update
set
  followed_repost_count = followed_repost_count + 1,
  followed_interaction_count = followed_interaction_count + 1;