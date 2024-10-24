// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package write

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
	items := []SubState{}
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
	items := []Session{}
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

const listAllAuthors = `-- name: ListAllAuthors :many
select
  "did"
from
  author
`

func (q *Queries) ListAllAuthors(ctx context.Context) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, listAllAuthors)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
	for rows.Next() {
		var did string
		if err := rows.Scan(&did); err != nil {
			return nil, err
		}
		items = append(items, did)
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
	items := []Following{}
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
from
  user
`

func (q *Queries) ListAllUsers(ctx context.Context) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, listAllUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
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

const listPostInteractionsForAuthor = `-- name: ListPostInteractionsForAuthor :many
select
  "directReplyCount",
  "interactionCount",
  "likeCount",
  "replyCount"
from
  post
where
  "author" = ?
`

type ListPostInteractionsForAuthorRow struct {
	DirectReplyCount float64
	InteractionCount float64
	LikeCount        float64
	ReplyCount       float64
}

func (q *Queries) ListPostInteractionsForAuthor(ctx context.Context, author string) ([]ListPostInteractionsForAuthorRow, error) {
	rows, err := q.db.QueryContext(ctx, listPostInteractionsForAuthor, author)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListPostInteractionsForAuthorRow{}
	for rows.Next() {
		var i ListPostInteractionsForAuthorRow
		if err := rows.Scan(
			&i.DirectReplyCount,
			&i.InteractionCount,
			&i.LikeCount,
			&i.ReplyCount,
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

const listUsersNotSeenSince = `-- name: ListUsersNotSeenSince :many
select
  "userDid"
from
  user
where
  "lastSeen" < ?1
`

func (q *Queries) ListUsersNotSeenSince(ctx context.Context, purgebefore string) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, listUsersNotSeenSince, purgebefore)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
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
