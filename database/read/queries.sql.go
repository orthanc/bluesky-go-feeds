// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package read

import (
	"context"
	"database/sql"
)

const getCursor = `-- name: GetCursor :many
select
  service, cursor
from
  sub_state
where
  "service" = ?
limit
  1
`

func (q *Queries) GetCursor(ctx context.Context, service string) ([]SubState, error) {
	rows, err := q.db.QueryContext(ctx, getCursor, service)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SubState
	for rows.Next() {
		var i SubState
		if err := rows.Scan(&i.Service, &i.Cursor); err != nil {
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

const getLastSession = `-- name: GetLastSession :many
select
  sessionId, userDid, startedAt, postsSince, lastSeen, accessCount, algo
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
`

type GetLastSessionParams struct {
	UserDid string
	Algo    sql.NullString
}

func (q *Queries) GetLastSession(ctx context.Context, arg GetLastSessionParams) ([]Session, error) {
	rows, err := q.db.QueryContext(ctx, getLastSession, arg.UserDid, arg.Algo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Session
	for rows.Next() {
		var i Session
		if err := rows.Scan(
			&i.SessionId,
			&i.UserDid,
			&i.StartedAt,
			&i.PostsSince,
			&i.LastSeen,
			&i.AccessCount,
			&i.Algo,
		); err != nil {
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

const listAllFollowing = `-- name: ListAllFollowing :many
select
  uri, followedBy, "following", userInteractionRatio
from
  following
`

func (q *Queries) ListAllFollowing(ctx context.Context) ([]Following, error) {
	rows, err := q.db.QueryContext(ctx, listAllFollowing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Following
	for rows.Next() {
		var i Following
		if err := rows.Scan(
			&i.Uri,
			&i.FollowedBy,
			&i.Following,
			&i.UserInteractionRatio,
		); err != nil {
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

const listAllUsers = `-- name: ListAllUsers :many
select
  "userDid"
FROM
  user
`

func (q *Queries) ListAllUsers(ctx context.Context) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, listAllUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var userDid string
		if err := rows.Scan(&userDid); err != nil {
			return nil, err
		}
		items = append(items, userDid)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPosts = `-- name: ListPosts :many
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
  post
`

type ListPostsRow struct {
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

func (q *Queries) ListPosts(ctx context.Context) ([]ListPostsRow, error) {
	rows, err := q.db.QueryContext(ctx, listPosts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListPostsRow
	for rows.Next() {
		var i ListPostsRow
		if err := rows.Scan(
			&i.Author,
			&i.DirectReplyCount,
			&i.IndexedAt,
			&i.InteractionCount,
			&i.LikeCount,
			&i.ReplyCount,
			&i.Uri,
			&i.ReplyParent,
			&i.ReplyParentAuthor,
			&i.ReplyRoot,
			&i.ReplyRootAuthor,
		); err != nil {
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
