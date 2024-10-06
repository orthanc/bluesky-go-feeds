// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: updates.sql

package write

import (
	"context"
	"database/sql"
)

const deleteFollowing = `-- name: DeleteFollowing :exec
delete from following
where
  uri = ?
`

func (q *Queries) DeleteFollowing(ctx context.Context, uri string) error {
	_, err := q.db.ExecContext(ctx, deleteFollowing, uri)
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