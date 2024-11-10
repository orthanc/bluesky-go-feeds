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

const getPostDates = `-- name: GetPostDates :one
select
  "indexedAt",
  created_at
from
  post
where
  uri = ?
`

type GetPostDatesRow struct {
	IndexedAt string
	CreatedAt sql.NullString
}

func (q *Queries) GetPostDates(ctx context.Context, uri string) (GetPostDatesRow, error) {
	row := q.db.QueryRowContext(ctx, getPostDates, uri)
	var i GetPostDatesRow
	err := row.Scan(&i.IndexedAt, &i.CreatedAt)
	return i, err
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
	var items []string
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

const listAllFollowers = `-- name: ListAllFollowers :many
select
  followed_by, "following", mutual, last_recorded
from
  follower
`

func (q *Queries) ListAllFollowers(ctx context.Context) ([]Follower, error) {
	rows, err := q.db.QueryContext(ctx, listAllFollowers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Follower
	for rows.Next() {
		var i Follower
		if err := rows.Scan(
			&i.FollowedBy,
			&i.Following,
			&i.Mutual,
			&i.LastRecorded,
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
  uri, followedBy, "following", userInteractionRatio, mutual, last_recorded
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
			&i.Mutual,
			&i.LastRecorded,
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

const listFollowerLastRecordedBefore = `-- name: ListFollowerLastRecordedBefore :many
select
  following,
  followed_by,
  last_recorded
from
  follower
where
  following = ?
  and last_recorded < ?
`

type ListFollowerLastRecordedBeforeParams struct {
	Following    string
	LastRecorded string
}

type ListFollowerLastRecordedBeforeRow struct {
	Following    string
	FollowedBy   string
	LastRecorded string
}

func (q *Queries) ListFollowerLastRecordedBefore(ctx context.Context, arg ListFollowerLastRecordedBeforeParams) ([]ListFollowerLastRecordedBeforeRow, error) {
	rows, err := q.db.QueryContext(ctx, listFollowerLastRecordedBefore, arg.Following, arg.LastRecorded)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListFollowerLastRecordedBeforeRow
	for rows.Next() {
		var i ListFollowerLastRecordedBeforeRow
		if err := rows.Scan(&i.Following, &i.FollowedBy, &i.LastRecorded); err != nil {
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

const listFollowingLastRecordedBefore = `-- name: ListFollowingLastRecordedBefore :many
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
  )
`

type ListFollowingLastRecordedBeforeParams struct {
	FollowedBy   string
	LastRecorded sql.NullString
}

type ListFollowingLastRecordedBeforeRow struct {
	Uri          string
	LastRecorded sql.NullString
}

func (q *Queries) ListFollowingLastRecordedBefore(ctx context.Context, arg ListFollowingLastRecordedBeforeParams) ([]ListFollowingLastRecordedBeforeRow, error) {
	rows, err := q.db.QueryContext(ctx, listFollowingLastRecordedBefore, arg.FollowedBy, arg.LastRecorded)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListFollowingLastRecordedBeforeRow
	for rows.Next() {
		var i ListFollowingLastRecordedBeforeRow
		if err := rows.Scan(&i.Uri, &i.LastRecorded); err != nil {
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
	var items []ListPostInteractionsForAuthorRow
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
