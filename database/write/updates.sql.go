// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: updates.sql

package write

import (
	"context"
	"database/sql"
	"strings"
)

const deleteAuthorsByDid = `-- name: DeleteAuthorsByDid :execrows
delete from author
where
  "did" in (/*SLICE:dids*/?)
`

func (q *Queries) DeleteAuthorsByDid(ctx context.Context, dids []string) (int64, error) {
	query := deleteAuthorsByDid
	var queryParams []interface{}
	if len(dids) > 0 {
		for _, v := range dids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:dids*/?", strings.Repeat(",?", len(dids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:dids*/?", "NULL", 1)
	}
	result, err := q.db.ExecContext(ctx, query, queryParams...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteFollower = `-- name: DeleteFollower :exec
delete from follower
where
  following = ?
  and followed_by = ?
  and last_recorded < ?
`

type DeleteFollowerParams struct {
	Following    string
	FollowedBy   string
	LastRecorded string
}

func (q *Queries) DeleteFollower(ctx context.Context, arg DeleteFollowerParams) error {
	_, err := q.db.ExecContext(ctx, deleteFollower, arg.Following, arg.FollowedBy, arg.LastRecorded)
	return err
}

const deleteFollowing = `-- name: DeleteFollowing :exec
delete from following
where
  uri = ?
`

func (q *Queries) DeleteFollowing(ctx context.Context, uri string) error {
	_, err := q.db.ExecContext(ctx, deleteFollowing, uri)
	return err
}

const deleteInteractionWithUsersBefore = `-- name: DeleteInteractionWithUsersBefore :execrows
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
  )
`

type DeleteInteractionWithUsersBeforeParams struct {
	IndexedAt string
	Limit     int64
}

func (q *Queries) DeleteInteractionWithUsersBefore(ctx context.Context, arg DeleteInteractionWithUsersBeforeParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteInteractionWithUsersBefore, arg.IndexedAt, arg.Limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteListMembershipNotRecordedBefore = `-- name: DeleteListMembershipNotRecordedBefore :exec
delete from list_membership
where
  list_uri = ?
  and last_recorded < ?
`

type DeleteListMembershipNotRecordedBeforeParams struct {
	ListUri      string
	LastRecorded string
}

func (q *Queries) DeleteListMembershipNotRecordedBefore(ctx context.Context, arg DeleteListMembershipNotRecordedBeforeParams) error {
	_, err := q.db.ExecContext(ctx, deleteListMembershipNotRecordedBefore, arg.ListUri, arg.LastRecorded)
	return err
}

const deletePostInteractedByFollowedBefore = `-- name: DeletePostInteractedByFollowedBefore :execrows
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
  )
`

type DeletePostInteractedByFollowedBeforeParams struct {
	IndexedAt string
	Limit     int64
}

func (q *Queries) DeletePostInteractedByFollowedBefore(ctx context.Context, arg DeletePostInteractedByFollowedBeforeParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, deletePostInteractedByFollowedBefore, arg.IndexedAt, arg.Limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deletePostsBefore = `-- name: DeletePostsBefore :execrows
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
  )
`

type DeletePostsBeforeParams struct {
	IndexedAt string
	Limit     int64
}

func (q *Queries) DeletePostsBefore(ctx context.Context, arg DeletePostsBeforeParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, deletePostsBefore, arg.IndexedAt, arg.Limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteRepostsBefore = `-- name: DeleteRepostsBefore :execrows
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
  )
`

type DeleteRepostsBeforeParams struct {
	IndexedAt string
	Limit     int64
}

func (q *Queries) DeleteRepostsBefore(ctx context.Context, arg DeleteRepostsBeforeParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteRepostsBefore, arg.IndexedAt, arg.Limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteSessionsBefore = `-- name: DeleteSessionsBefore :execrows
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
  )
`

type DeleteSessionsBeforeParams struct {
	LastSeen string
	Limit    int64
}

func (q *Queries) DeleteSessionsBefore(ctx context.Context, arg DeleteSessionsBeforeParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteSessionsBefore, arg.LastSeen, arg.Limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteUserInteractionsBefore = `-- name: DeleteUserInteractionsBefore :execrows
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
  )
`

type DeleteUserInteractionsBeforeParams struct {
	IndexedAt string
	Limit     int64
}

func (q *Queries) DeleteUserInteractionsBefore(ctx context.Context, arg DeleteUserInteractionsBeforeParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteUserInteractionsBefore, arg.IndexedAt, arg.Limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteUserLinksBefore = `-- name: DeleteUserLinksBefore :execrows
delete from user_link
where
  rowid in (
    select
      rowid
    from
      user_link
    where
      user_link.last_seen <= ?
    limit
      ?
  )
`

type DeleteUserLinksBeforeParams struct {
	LastSeen string
	Limit    int64
}

func (q *Queries) DeleteUserLinksBefore(ctx context.Context, arg DeleteUserLinksBeforeParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteUserLinksBefore, arg.LastSeen, arg.Limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteUserWhenNotSeen = `-- name: DeleteUserWhenNotSeen :execrows
delete from user
where
  "userDid" = ?
  and "lastSeen" < ?2
`

type DeleteUserWhenNotSeenParams struct {
	UserDid     string
	PurgeBefore string
}

func (q *Queries) DeleteUserWhenNotSeen(ctx context.Context, arg DeleteUserWhenNotSeenParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteUserWhenNotSeen, arg.UserDid, arg.PurgeBefore)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const incrementPostDirectReply = `-- name: IncrementPostDirectReply :exec
update post
set
  "directReplyCount" = "directReplyCount" + 1,
  "replyCount" = "replyCount" + 1,
  "interactionCount" = "interactionCount" + 1
where
  "uri" = ?
`

func (q *Queries) IncrementPostDirectReply(ctx context.Context, uri string) error {
	_, err := q.db.ExecContext(ctx, incrementPostDirectReply, uri)
	return err
}

const incrementPostIndirectReply = `-- name: IncrementPostIndirectReply :exec
update post
set
  "replyCount" = "replyCount" + 1,
  "interactionCount" = "interactionCount" + 1
where
  "uri" = ?
`

func (q *Queries) IncrementPostIndirectReply(ctx context.Context, uri string) error {
	_, err := q.db.ExecContext(ctx, incrementPostIndirectReply, uri)
	return err
}

const incrementPostLike = `-- name: IncrementPostLike :exec
update post
set
  "likeCount" = "likeCount" + 1,
  "interactionCount" = "interactionCount" + 1
where
  "uri" = ?
`

func (q *Queries) IncrementPostLike(ctx context.Context, uri string) error {
	_, err := q.db.ExecContext(ctx, incrementPostLike, uri)
	return err
}

const incrementPostRepost = `-- name: IncrementPostRepost :exec
update post
set
  repost_count = repost_count + 1,
  "interactionCount" = "interactionCount" + 1
where
  "uri" = ?
`

func (q *Queries) IncrementPostRepost(ctx context.Context, uri string) error {
	_, err := q.db.ExecContext(ctx, incrementPostRepost, uri)
	return err
}

const saveAuthor = `-- name: SaveAuthor :exec
insert into
  author (
    "did",
    "medianDirectReplyCount",
    "medianInteractionCount",
    "medianLikeCount",
    "medianReplyCount"
  )
values
  (?, ?, ?, ?, ?) on conflict do nothing
`

type SaveAuthorParams struct {
	Did                    string
	MedianDirectReplyCount float64
	MedianInteractionCount float64
	MedianLikeCount        float64
	MedianReplyCount       float64
}

func (q *Queries) SaveAuthor(ctx context.Context, arg SaveAuthorParams) error {
	_, err := q.db.ExecContext(ctx, saveAuthor,
		arg.Did,
		arg.MedianDirectReplyCount,
		arg.MedianInteractionCount,
		arg.MedianLikeCount,
		arg.MedianReplyCount,
	)
	return err
}

const saveCursor = `-- name: SaveCursor :exec
insert into
  sub_state ("service", "cursor")
values
  (?, ?) on conflict do
update
set
  "cursor" = excluded."cursor"
`

type SaveCursorParams struct {
	Service string
	Cursor  int64
}

func (q *Queries) SaveCursor(ctx context.Context, arg SaveCursorParams) error {
	_, err := q.db.ExecContext(ctx, saveCursor, arg.Service, arg.Cursor)
	return err
}

const saveFollower = `-- name: SaveFollower :exec
insert into
  follower (followed_by, following, last_recorded)
values
  (?, ?, ?) on conflict do
update
set
  last_recorded = excluded.last_recorded
`

type SaveFollowerParams struct {
	FollowedBy   string
	Following    string
	LastRecorded string
}

func (q *Queries) SaveFollower(ctx context.Context, arg SaveFollowerParams) error {
	_, err := q.db.ExecContext(ctx, saveFollower, arg.FollowedBy, arg.Following, arg.LastRecorded)
	return err
}

const saveFollowing = `-- name: SaveFollowing :exec
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
  last_recorded = excluded.last_recorded
`

type SaveFollowingParams struct {
	Uri                  string
	FollowedBy           string
	Following            string
	UserInteractionRatio sql.NullFloat64
	LastRecorded         sql.NullString
}

func (q *Queries) SaveFollowing(ctx context.Context, arg SaveFollowingParams) error {
	_, err := q.db.ExecContext(ctx, saveFollowing,
		arg.Uri,
		arg.FollowedBy,
		arg.Following,
		arg.UserInteractionRatio,
		arg.LastRecorded,
	)
	return err
}

const saveInteractionWithUser = `-- name: SaveInteractionWithUser :exec
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
  (?, ?, ?, ?, ?, ?) on conflict do nothing
`

type SaveInteractionWithUserParams struct {
	InteractionUri       string
	UserDid              string
	Type                 string
	InteractionAuthorDid string
	PostUri              string
	IndexedAt            string
}

func (q *Queries) SaveInteractionWithUser(ctx context.Context, arg SaveInteractionWithUserParams) error {
	_, err := q.db.ExecContext(ctx, saveInteractionWithUser,
		arg.InteractionUri,
		arg.UserDid,
		arg.Type,
		arg.InteractionAuthorDid,
		arg.PostUri,
		arg.IndexedAt,
	)
	return err
}

const saveListMembership = `-- name: SaveListMembership :exec
insert into
  list_membership (list_uri, member_did, last_recorded)
values
  (?, ?, ?) on conflict do
update
set
  last_recorded = excluded.last_recorded
`

type SaveListMembershipParams struct {
	ListUri      string
	MemberDid    string
	LastRecorded string
}

func (q *Queries) SaveListMembership(ctx context.Context, arg SaveListMembershipParams) error {
	_, err := q.db.ExecContext(ctx, saveListMembership, arg.ListUri, arg.MemberDid, arg.LastRecorded)
	return err
}

const savePost = `-- name: SavePost :exec
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
    "replyRootAuthor",
    posters_madness,
    external_uri,
    quoted_post_uri
  )
values
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on conflict do nothing
`

type SavePostParams struct {
	Author            string
	DirectReplyCount  float64
	IndexedAt         string
	CreatedAt         sql.NullString
	InteractionCount  float64
	LikeCount         float64
	ReplyCount        float64
	Uri               string
	ReplyParent       sql.NullString
	ReplyParentAuthor sql.NullString
	ReplyRoot         sql.NullString
	ReplyRootAuthor   sql.NullString
	PostersMadness    sql.NullInt64
	ExternalUri       sql.NullString
	QuotedPostUri     sql.NullString
}

func (q *Queries) SavePost(ctx context.Context, arg SavePostParams) error {
	_, err := q.db.ExecContext(ctx, savePost,
		arg.Author,
		arg.DirectReplyCount,
		arg.IndexedAt,
		arg.CreatedAt,
		arg.InteractionCount,
		arg.LikeCount,
		arg.ReplyCount,
		arg.Uri,
		arg.ReplyParent,
		arg.ReplyParentAuthor,
		arg.ReplyRoot,
		arg.ReplyRootAuthor,
		arg.PostersMadness,
		arg.ExternalUri,
		arg.QuotedPostUri,
	)
	return err
}

const savePostDirectRepliedToByFollowing = `-- name: SavePostDirectRepliedToByFollowing :exec
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
  followed_interaction_count = followed_interaction_count + 1
`

type SavePostDirectRepliedToByFollowingParams struct {
	User      string
	Uri       string
	Author    string
	IndexedAt string
}

func (q *Queries) SavePostDirectRepliedToByFollowing(ctx context.Context, arg SavePostDirectRepliedToByFollowingParams) error {
	_, err := q.db.ExecContext(ctx, savePostDirectRepliedToByFollowing,
		arg.User,
		arg.Uri,
		arg.Author,
		arg.IndexedAt,
	)
	return err
}

const savePostLikedByFollowing = `-- name: SavePostLikedByFollowing :exec
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
  followed_interaction_count = followed_interaction_count + 1
`

type SavePostLikedByFollowingParams struct {
	User      string
	Uri       string
	Author    string
	IndexedAt string
}

func (q *Queries) SavePostLikedByFollowing(ctx context.Context, arg SavePostLikedByFollowingParams) error {
	_, err := q.db.ExecContext(ctx, savePostLikedByFollowing,
		arg.User,
		arg.Uri,
		arg.Author,
		arg.IndexedAt,
	)
	return err
}

const savePostRepliedToByFollowing = `-- name: SavePostRepliedToByFollowing :exec
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
  followed_interaction_count = followed_interaction_count + 1
`

type SavePostRepliedToByFollowingParams struct {
	User      string
	Uri       string
	Author    string
	IndexedAt string
}

func (q *Queries) SavePostRepliedToByFollowing(ctx context.Context, arg SavePostRepliedToByFollowingParams) error {
	_, err := q.db.ExecContext(ctx, savePostRepliedToByFollowing,
		arg.User,
		arg.Uri,
		arg.Author,
		arg.IndexedAt,
	)
	return err
}

const savePostRepostedByFollowing = `-- name: SavePostRepostedByFollowing :exec
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
  followed_interaction_count = followed_interaction_count + 1
`

type SavePostRepostedByFollowingParams struct {
	User      string
	Uri       string
	Author    string
	IndexedAt string
}

func (q *Queries) SavePostRepostedByFollowing(ctx context.Context, arg SavePostRepostedByFollowingParams) error {
	_, err := q.db.ExecContext(ctx, savePostRepostedByFollowing,
		arg.User,
		arg.Uri,
		arg.Author,
		arg.IndexedAt,
	)
	return err
}

const saveSession = `-- name: SaveSession :exec
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
  (?, ?, ?, ?, ?, ?)
`

type SaveSessionParams struct {
	UserDid     string
	StartedAt   string
	PostsSince  string
	LastSeen    string
	AccessCount sql.NullFloat64
	Algo        sql.NullString
}

func (q *Queries) SaveSession(ctx context.Context, arg SaveSessionParams) error {
	_, err := q.db.ExecContext(ctx, saveSession,
		arg.UserDid,
		arg.StartedAt,
		arg.PostsSince,
		arg.LastSeen,
		arg.AccessCount,
		arg.Algo,
	)
	return err
}

const saveUser = `-- name: SaveUser :exec
insert into
  user ("userDid", "lastSeen", "lastSynced")
values
  (?, ?, ?) on conflict do
update
set
  "lastSeen" = excluded."lastSeen",
  "lastSynced" = excluded."lastSynced"
`

type SaveUserParams struct {
	UserDid    string
	LastSeen   string
	LastSynced sql.NullString
}

func (q *Queries) SaveUser(ctx context.Context, arg SaveUserParams) error {
	_, err := q.db.ExecContext(ctx, saveUser, arg.UserDid, arg.LastSeen, arg.LastSynced)
	return err
}

const saveUserInteraction = `-- name: SaveUserInteraction :exec
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
  (?, ?, ?, ?, ?, ?) on conflict do nothing
`

type SaveUserInteractionParams struct {
	InteractionUri string
	UserDid        string
	Type           string
	AuthorDid      string
	PostUri        string
	IndexedAt      string
}

func (q *Queries) SaveUserInteraction(ctx context.Context, arg SaveUserInteractionParams) error {
	_, err := q.db.ExecContext(ctx, saveUserInteraction,
		arg.InteractionUri,
		arg.UserDid,
		arg.Type,
		arg.AuthorDid,
		arg.PostUri,
		arg.IndexedAt,
	)
	return err
}

const saveUserLink = `-- name: SaveUserLink :exec
insert into
  user_link (
    user_did,
    link_uri,
    share_count,
    first_seen,
    last_seen,
    first_seen_post_uri,
    last_seen_post_uri
  )
select
  followedBy,
  ?1,
  1,
  ?2,
  ?2,
  ?3,
  ?3
from
  following
where
  following.following = ?4 on conflict do
update
set
  share_count = share_count + 1,
  last_seen = excluded.last_seen,
  last_seen_post_uri = excluded.last_seen_post_uri
`

type SaveUserLinkParams struct {
	LinkUri    string
	SeenAt     string
	PostUri    string
	PostAuthor string
}

func (q *Queries) SaveUserLink(ctx context.Context, arg SaveUserLinkParams) error {
	_, err := q.db.ExecContext(ctx, saveUserLink,
		arg.LinkUri,
		arg.SeenAt,
		arg.PostUri,
		arg.PostAuthor,
	)
	return err
}

const updateSessionLastSeen = `-- name: UpdateSessionLastSeen :exec
update session
set
  "lastSeen" = ?,
  "accessCount" = "accessCount" + 1
where
  "sessionId" = ?
`

type UpdateSessionLastSeenParams struct {
	LastSeen  string
	SessionId int64
}

func (q *Queries) UpdateSessionLastSeen(ctx context.Context, arg UpdateSessionLastSeenParams) error {
	_, err := q.db.ExecContext(ctx, updateSessionLastSeen, arg.LastSeen, arg.SessionId)
	return err
}

const updateUserLastSeen = `-- name: UpdateUserLastSeen :many
update user
set
  "lastSeen" = ?
where
  "userDid" = ? returning userDid, lastSeen, lastSynced
`

type UpdateUserLastSeenParams struct {
	LastSeen string
	UserDid  string
}

func (q *Queries) UpdateUserLastSeen(ctx context.Context, arg UpdateUserLastSeenParams) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, updateUserLastSeen, arg.LastSeen, arg.UserDid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(&i.UserDid, &i.LastSeen, &i.LastSynced); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
