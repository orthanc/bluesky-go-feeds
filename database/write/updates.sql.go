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
  "indexedAt" < ?
`

func (q *Queries) DeleteInteractionWithUsersBefore(ctx context.Context, indexedat string) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteInteractionWithUsersBefore, indexedat)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deletePostsBefore = `-- name: DeletePostsBefore :execrows
delete from post
where
  "indexedAt" < ?
`

func (q *Queries) DeletePostsBefore(ctx context.Context, indexedat string) (int64, error) {
	result, err := q.db.ExecContext(ctx, deletePostsBefore, indexedat)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteRepostsBefore = `-- name: DeleteRepostsBefore :execrows
delete from repost
where
  "indexedAt" < ?
`

func (q *Queries) DeleteRepostsBefore(ctx context.Context, indexedat string) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteRepostsBefore, indexedat)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteSessionsBefore = `-- name: DeleteSessionsBefore :execrows
delete from session
where
  "lastSeen" < ?
`

func (q *Queries) DeleteSessionsBefore(ctx context.Context, lastseen string) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteSessionsBefore, lastseen)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteUserInteractionsBefore = `-- name: DeleteUserInteractionsBefore :execrows
delete from userInteraction
where
  "indexedAt" < ?
`

func (q *Queries) DeleteUserInteractionsBefore(ctx context.Context, indexedat string) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteUserInteractionsBefore, indexedat)
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

const saveFollowing = `-- name: SaveFollowing :exec
insert into
  following (
    "uri",
    "followedBy",
    "following",
    "userInteractionRatio"
  )
values
  (?, ?, ?, ?) on conflict do nothing
`

type SaveFollowingParams struct {
	Uri                  string
	FollowedBy           string
	Following            string
	UserInteractionRatio sql.NullFloat64
}

func (q *Queries) SaveFollowing(ctx context.Context, arg SaveFollowingParams) error {
	_, err := q.db.ExecContext(ctx, saveFollowing,
		arg.Uri,
		arg.FollowedBy,
		arg.Following,
		arg.UserInteractionRatio,
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

const savePost = `-- name: SavePost :exec
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
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on conflict do nothing
`

type SavePostParams struct {
	Author            string
	DirectReplyCount  float64
	IndexedAt         string
	InteractionCount  float64
	LikeCount         float64
	ReplyCount        float64
	Uri               string
	ReplyParent       sql.NullString
	ReplyParentAuthor sql.NullString
	ReplyRoot         sql.NullString
	ReplyRootAuthor   sql.NullString
}

func (q *Queries) SavePost(ctx context.Context, arg SavePostParams) error {
	_, err := q.db.ExecContext(ctx, savePost,
		arg.Author,
		arg.DirectReplyCount,
		arg.IndexedAt,
		arg.InteractionCount,
		arg.LikeCount,
		arg.ReplyCount,
		arg.Uri,
		arg.ReplyParent,
		arg.ReplyParentAuthor,
		arg.ReplyRoot,
		arg.ReplyRootAuthor,
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
  user ("userDid", "lastSeen")
values
  (?, ?) on conflict do
update
set
  "lastSeen" = excluded."lastSeen"
`

type SaveUserParams struct {
	UserDid  string
	LastSeen string
}

func (q *Queries) SaveUser(ctx context.Context, arg SaveUserParams) error {
	_, err := q.db.ExecContext(ctx, saveUser, arg.UserDid, arg.LastSeen)
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

const updateAuthorMedians = `-- name: UpdateAuthorMedians :exec
update author
set
  "medianReplyCount" = ?,
  "medianDirectReplyCount" = ?,
  "medianLikeCount" = ?,
  "medianInteractionCount" = ?
where
  "did" = ?
`

type UpdateAuthorMediansParams struct {
	MedianReplyCount       float64
	MedianDirectReplyCount float64
	MedianLikeCount        float64
	MedianInteractionCount float64
	Did                    string
}

func (q *Queries) UpdateAuthorMedians(ctx context.Context, arg UpdateAuthorMediansParams) error {
	_, err := q.db.ExecContext(ctx, updateAuthorMedians,
		arg.MedianReplyCount,
		arg.MedianDirectReplyCount,
		arg.MedianLikeCount,
		arg.MedianInteractionCount,
		arg.Did,
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

const updateUserLastSeen = `-- name: UpdateUserLastSeen :execrows
update user
set
  "lastSeen" = ?
where
  "userDid" = ?
`

type UpdateUserLastSeenParams struct {
	LastSeen string
	UserDid  string
}

func (q *Queries) UpdateUserLastSeen(ctx context.Context, arg UpdateUserLastSeenParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateUserLastSeen, arg.LastSeen, arg.UserDid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
