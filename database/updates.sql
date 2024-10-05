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

-- name: SaveUser :exec
insert into
  user (
    "userDid",
    "lastSeen"
  )
values
  (?, ?) on conflict do update set lastSeen = excluded.lastSeen;

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
delete from following where uri = ?;

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