// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package read

import (
	"context"
	"database/sql"
	"strings"
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

const getFollowingFollowData = `-- name: GetFollowingFollowData :one
select
  (
    select
      count(*)
    from
      user as followed_by_user
    where
      followed_by_user."userDid" = ?1
  ) as follow_by_user,
  (
    select
      count(*)
    from
      user as following_user
    where
      following_user."userDid" = ?2
  ) as following_user
`

type GetFollowingFollowDataParams struct {
	FollowAuthor  string
	FollowSubject string
}

type GetFollowingFollowDataRow struct {
	FollowByUser  int64
	FollowingUser int64
}

func (q *Queries) GetFollowingFollowData(ctx context.Context, arg GetFollowingFollowDataParams) (GetFollowingFollowDataRow, error) {
	row := q.db.QueryRowContext(ctx, getFollowingFollowData, arg.FollowAuthor, arg.FollowSubject)
	var i GetFollowingFollowDataRow
	err := row.Scan(&i.FollowByUser, &i.FollowingUser)
	return i, err
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

const getLikeFollowData = `-- name: GetLikeFollowData :one
select
  (
    select
      count(*)
    from
      user as postUser
    where
      postUser."userDid" = ?1
  ) as post_by_user,
  (
    select
      count(*)
    from
      author
    where
      did = ?1
  ) as post_by_author,
  (
    select
      count(*)
    from
      user as likeUser
    where
      likeUser."userDid" = ?2
  ) as like_by_user
`

type GetLikeFollowDataParams struct {
	PostAuthor string
	LikeAuthor string
}

type GetLikeFollowDataRow struct {
	PostByUser   int64
	PostByAuthor int64
	LikeByUser   int64
}

func (q *Queries) GetLikeFollowData(ctx context.Context, arg GetLikeFollowDataParams) (GetLikeFollowDataRow, error) {
	row := q.db.QueryRowContext(ctx, getLikeFollowData, arg.PostAuthor, arg.LikeAuthor)
	var i GetLikeFollowDataRow
	err := row.Scan(&i.PostByUser, &i.PostByAuthor, &i.LikeByUser)
	return i, err
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

const getPostFollowData = `-- name: GetPostFollowData :one
select
  (
    select
      count(*)
    from
      author as post_author
    where
      post_author.did = ?1
  ) as post_by_author,
  (
    select
      count(*)
    from
      user as post_user
    where
      post_user.userDid = ?1
  ) as post_by_user,
  (
    select
      count(*)
    from
      author as reply_parent_author
    where
      reply_parent_author.did = ?2
  ) as reply_to_author,
  (
    select
      count(*)
    from
      user as reply_parent_user
    where
      reply_parent_user.userDid = ?2
  ) as reply_to_user,
  (
    select
      count(*)
    from
      author as reply_root_author
    where
      reply_root_author.did = ?3
  ) as reply_to_thread_author,
  (
    select
      count(*)
    from
      user as reply_root_user
    where
      reply_root_user.userDid = ?3
  ) as reply_to_thread_user,
  (
    select
      count(*)
    from
      posters_madness
    where
      stage = 'symptomatic'
      and poster_did = ?1
  ) as posters_madness_symptomatic,
  (
    select
      count(*)
    from
      posters_madness
    where
      stage = 'symptomatic'
      and poster_did = ?2
  ) as posters_madness_reply_to_symptomatic
`

type GetPostFollowDataParams struct {
	PostAuthor        string
	ReplyParentAuthor string
	ReplyRootAuthor   string
}

type GetPostFollowDataRow struct {
	PostByAuthor                     int64
	PostByUser                       int64
	ReplyToAuthor                    int64
	ReplyToUser                      int64
	ReplyToThreadAuthor              int64
	ReplyToThreadUser                int64
	PostersMadnessSymptomatic        int64
	PostersMadnessReplyToSymptomatic int64
}

func (q *Queries) GetPostFollowData(ctx context.Context, arg GetPostFollowDataParams) (GetPostFollowDataRow, error) {
	row := q.db.QueryRowContext(ctx, getPostFollowData, arg.PostAuthor, arg.ReplyParentAuthor, arg.ReplyRootAuthor)
	var i GetPostFollowDataRow
	err := row.Scan(
		&i.PostByAuthor,
		&i.PostByUser,
		&i.ReplyToAuthor,
		&i.ReplyToUser,
		&i.ReplyToThreadAuthor,
		&i.ReplyToThreadUser,
		&i.PostersMadnessSymptomatic,
		&i.PostersMadnessReplyToSymptomatic,
	)
	return i, err
}

const getPostersMadnessInStageNotUpdatedSince = `-- name: GetPostersMadnessInStageNotUpdatedSince :many
select
  poster_did,
  stage,
  last_checked
from
  posters_madness
where
  last_checked < ?
  and stage = ?
`

type GetPostersMadnessInStageNotUpdatedSinceParams struct {
	LastChecked string
	Stage       string
}

func (q *Queries) GetPostersMadnessInStageNotUpdatedSince(ctx context.Context, arg GetPostersMadnessInStageNotUpdatedSinceParams) ([]PostersMadness, error) {
	rows, err := q.db.QueryContext(ctx, getPostersMadnessInStageNotUpdatedSince, arg.LastChecked, arg.Stage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PostersMadness
	for rows.Next() {
		var i PostersMadness
		if err := rows.Scan(&i.PosterDid, &i.Stage, &i.LastChecked); err != nil {
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

const getPostersMadnessNotUpdatedSince = `-- name: GetPostersMadnessNotUpdatedSince :many
select
  poster_did,
  stage,
  last_checked
from
  posters_madness
where
  last_checked < ?
  and stage <> ?
`

type GetPostersMadnessNotUpdatedSinceParams struct {
	LastChecked string
	Stage       string
}

func (q *Queries) GetPostersMadnessNotUpdatedSince(ctx context.Context, arg GetPostersMadnessNotUpdatedSinceParams) ([]PostersMadness, error) {
	rows, err := q.db.QueryContext(ctx, getPostersMadnessNotUpdatedSince, arg.LastChecked, arg.Stage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PostersMadness
	for rows.Next() {
		var i PostersMadness
		if err := rows.Scan(&i.PosterDid, &i.Stage, &i.LastChecked); err != nil {
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

const getPostersMadnessStats = `-- name: GetPostersMadnessStats :many
select
  stage,
  cnt
from
  posters_madness_stats
`

func (q *Queries) GetPostersMadnessStats(ctx context.Context) ([]PostersMadnessStat, error) {
	rows, err := q.db.QueryContext(ctx, getPostersMadnessStats)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PostersMadnessStat
	for rows.Next() {
		var i PostersMadnessStat
		if err := rows.Scan(&i.Stage, &i.Cnt); err != nil {
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

const getPostersMadnessStatus = `-- name: GetPostersMadnessStatus :many
select
  poster_did,
  stage,
  last_checked
from
  posters_madness
where
  poster_did in (/*SLICE:dids*/?)
`

func (q *Queries) GetPostersMadnessStatus(ctx context.Context, dids []string) ([]PostersMadness, error) {
	query := getPostersMadnessStatus
	var queryParams []interface{}
	if len(dids) > 0 {
		for _, v := range dids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:dids*/?", strings.Repeat(",?", len(dids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:dids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PostersMadness
	for rows.Next() {
		var i PostersMadness
		if err := rows.Scan(&i.PosterDid, &i.Stage, &i.LastChecked); err != nil {
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

const isOnList = `-- name: IsOnList :one
select
  count(*)
from
  list_membership
where
  list_uri = ?
  and member_did = ?
`

type IsOnListParams struct {
	ListUri   string
	MemberDid string
}

func (q *Queries) IsOnList(ctx context.Context, arg IsOnListParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, isOnList, arg.ListUri, arg.MemberDid)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const listAllAuthors = `-- name: ListAllAuthors :many
select
  "did"
from
  author
order by
  hex (randomblob (16))
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
