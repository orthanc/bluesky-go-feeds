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

-- name: GetPostFollowData :one
select
  (
    select
      count(*)
    from
      author as post_author
    where
      post_author.did = sqlc.arg (post_author)
  ) as post_by_author,
  (
    select
      count(*)
    from
      user as post_user
    where
      post_user.userDid = sqlc.arg (post_author)
  ) as post_by_user,
  (
    select
      count(*)
    from
      author as reply_parent_author
    where
      reply_parent_author.did = sqlc.arg (reply_parent_author)
  ) as reply_to_author,
  (
    select
      count(*)
    from
      user as reply_parent_user
    where
      reply_parent_user.userDid = sqlc.arg (reply_parent_author)
  ) as reply_to_user,
  (
    select
      count(*)
    from
      author as reply_root_author
    where
      reply_root_author.did = sqlc.arg (reply_root_author)
  ) as reply_to_thread_author,
  (
    select
      count(*)
    from
      user as reply_root_user
    where
      reply_root_user.userDid = sqlc.arg (reply_root_author)
  ) as reply_to_thread_user,
  (
    select
      count(*)
    from
      posters_madness
    where
      stage in ('pre-infectious', 'infectious', 'post-infectious')
      and poster_did = sqlc.arg (post_author)
  ) as posters_madness_symptomatic,
  (
    select
      count(*)
    from
      posters_madness
    where
      stage in ('pre-infectious', 'infectious', 'post-infectious')
      and poster_did = sqlc.arg (reply_parent_author)
  ) as posters_madness_reply_to_symptomatic;

-- name: GetFollowingFollowData :one
select
  (
    select
      count(*)
    from
      user as followed_by_user
    where
      followed_by_user."userDid" = sqlc.arg (follow_author)
  ) as follow_by_user,
  (
    select
      count(*)
    from
      user as following_user
    where
      following_user."userDid" = sqlc.arg (follow_subject)
  ) as following_user;

-- name: IsOnList :one
select
  count(*)
from
  list_membership
where
  list_uri = ?
  and member_did = ?;

-- name: GetPostersMadnessStatus :many
select
  poster_did,
  stage,
  last_checked
from
  posters_madness
where
  poster_did in (sqlc.slice ('dids'));

-- name: GetPostersMadnessNotUpdatedSince :many
select
  poster_did,
  stage,
  last_checked
from
  posters_madness
where
  last_checked < ?
  and stage <> ?;

-- name: GetPostersMadnessInStageNotUpdatedSince :many
select
  poster_did,
  stage,
  last_checked
from
  posters_madness
where
  last_checked < ?
  and stage = ?;

-- name: GetPostersMadnessStats :many
select
  stage,
  cnt
from
  posters_madness_stats;