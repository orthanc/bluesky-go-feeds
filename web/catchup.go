package web

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
)

type catchupQueryRow = struct {
	Uri       string
	IndexedAt string
	Rating    float64
}

const catchupAlgorithmId = "catchup"
const catchupMutualsAlgorithmId = "3l7hnqx43hpea"
const catchupFollowersAlgorithmId = "3l7hnrrppswb3"

const catchupQueryPrefix = `
select
  "post"."uri",
  "post"."indexedAt",
  (
    "post"."interactionCount" - "author"."medianInteractionCount"
  ) * 1000 / ("author"."medianInteractionCount" + 1) as "rating"
from
  "post"
  inner join "author" on "post"."author" = "author"."did"
  inner join "following" on "post"."author" = "following"."following"
where
  "following"."followedBy" = ?
  and "post"."indexedAt" >= ?
`
const catchupQuerySuffix = `
order by
  "rating" desc,
  "indexedAt" desc
limit
  ?
offset
  ?
`

const catchupQuery = catchupQueryPrefix + catchupQuerySuffix
const catchupMutualsQuery = catchupQueryPrefix + `
  and following.mutual > 0
` + catchupQuerySuffix
const catchupFollowersQuery = `
select
	"post"."uri",
	"post"."indexedAt",
	(
		"post"."interactionCount" - "author"."medianInteractionCount"
	) * 1000 / ("author"."medianInteractionCount" + 1) as "rating"
from
	"post"
	inner join "author" on "post"."author" = "author"."did"
	inner join follower on "post"."author" = follower.followed_by
where
	follower.following = ?
	and "post"."indexedAt" >= ?
  and follower.mutual = 0
` + catchupQuerySuffix

func catchupVariant(name string, query string, ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
	endLog := logFeedAccess(name, session)
	defer endLog()
	output := bsky.FeedGetFeedSkeleton_Output{
		Feed: make([]*bsky.FeedDefs_SkeletonFeedPost, 0, limit),
	}
	offset := 0
	if cursor != "" {
		parsedOffset, err := strconv.Atoi(cursor)
		if err != nil {
			return output, fmt.Errorf("unable to parse cursor %s: %w", cursor, err)
		}
		offset = parsedOffset
	}
	rows, err := database.QueryContext(ctx, query, session.UserDid, session.PostsSince, limit, offset)
	if err != nil {
		return output, fmt.Errorf("error executing %s query: %s", name, err)
	}
	defer rows.Close()
	var row catchupQueryRow
	for rows.Next() {
		err := rows.Scan(
			&row.Uri,
			&row.IndexedAt,
			&row.Rating,
		)
		if err != nil {
			return output, err
		}
		// w, _ := json.Marshal(row)
		// fmt.Println(string(w))
		output.Feed = append(output.Feed, &bsky.FeedDefs_SkeletonFeedPost{
			Post: row.Uri,
		})
	}
	if len(output.Feed) > 0 {
		nextCursor := strconv.Itoa(offset + len(output.Feed))
		output.Cursor = &nextCursor
	}
	return output, nil
}

func catchup(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
	return catchupVariant("catchup", catchupQuery, ctx, database, session, cursor, limit)
}

func catchupMutuals(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
	return catchupVariant("catchupMutuals", catchupMutualsQuery, ctx, database, session, cursor, limit)
}

func catchupFollowers(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
	return catchupVariant("catchupFollowers", catchupFollowersQuery, ctx, database, session, cursor, limit)
}
