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
	Uri string
  IndexedAt string
  Rating float64
}

const catchupQueryAlgorithmId = "catchup"

const catchupQuery = `
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
order by
  "rating" desc,
  "indexedAt" desc
limit
  ?
offset
  ?
`

func catchup(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
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
	rows, err := database.QueryContext(ctx, catchupQuery, session.UserDid, session.PostsSince, limit, offset)
	if err != nil {
		return output, fmt.Errorf("error executing catchup query: %s", err)
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