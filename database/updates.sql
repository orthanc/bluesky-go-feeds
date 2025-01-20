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
    created_at,
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
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on conflict do nothing;

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
  rowid in (
    select
      rowid
    from
      post
    where
      post."indexedAt" < ?
    limit
      ?
  );

-- name: DeleteRepostsBefore :execrows
delete from repost
where
  rowid in (
    select
      rowid
    from
      repost
    where
      repost."indexedAt" < ?
    limit
      ?
  );

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
  rowid in (
    select
      rowid
    from
      userInteraction
    where
      "userInteraction"."indexedAt" < ?
    limit
      ?
  );

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
  rowid in (
    select
      rowid
    from
      interactionWithUser
    where
      "interactionWithUser"."indexedAt" < ?
    limit
      ?
  );

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
  rowid in (
    select
      rowid
    from
      session
    where
      session."lastSeen" < ?
    limit
      ?
  );

-- name: DeletePostInteractedByFollowedBefore :execrows
delete from post_interacted_by_followed
where
  rowid in (
    select
      rowid
    from
      post_interacted_by_followed
    where
      post_interacted_by_followed.indexed_at < ?
    limit
      ?
  );

-- name: SaveFollowing :exec
insert into
  following (
    "uri",
    "followedBy",
    "following",
    "userInteractionRatio",
    last_recorded
  )
values
  (?, ?, ?, ?, ?) on conflict do
update
set
  last_recorded = excluded.last_recorded;

-- name: DeleteFollowing :exec
delete from following
where
  uri = ?;

-- name: SaveFollower :exec
insert into
  follower (followed_by, following, last_recorded)
values
  (?, ?, ?) on conflict do
update
set
  last_recorded = excluded.last_recorded;

-- name: DeleteFollower :exec
delete from follower
where
  following = ?
  and followed_by = ?
  and last_recorded < ?;

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
  indexed_at = excluded.indexed_at,
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
  indexed_at = excluded.indexed_at,
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
  indexed_at = excluded.indexed_at,
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
  indexed_at = excluded.indexed_at,
  followed_repost_count = followed_repost_count + 1,
  followed_interaction_count = followed_interaction_count + 1;

-- name: SaveListMembership :exec
insert into
  list_membership (list_uri, member_did, last_recorded)
values
  (?, ?, ?) on conflict do
update
set
  last_recorded = excluded.last_recorded;

-- name: DeleteListMembershipNotRecordedBefore :exec
delete from list_membership
where
  list_uri = ?
  and last_recorded < ?;
