package web

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
)

type loveliesQueryRow = struct {
	Uri                  string
	IndexedAt            string
	Author               string
	UserInteractionRatio float64
	IScore               float64
	TScore               float64
	Rating               float64
}

const loveliesAlgorithmId = "o1s6niihick9"

const loveliesQuery = `
with
  scoredPost as (
    select
      "post"."uri",
      "post"."indexedAt",
      "post"."author",
      "following"."userInteractionRatio",
      "following"."followedBy",
      "post"."replyParent",
      (
        "userInteractionRatio" * (unixepoch ('now') - unixepoch ("indexedAt")) / (24.0 * 3600) * 0.20
      ) as "iScore",
      (
        unixepoch ("indexedAt") - unixepoch ('now', '-7 days')
      ) / (24.0 * 3600) as "tScore"
    from
      "post"
      inner join "following" on "post"."author" = "following"."following"
  )
select
  "uri",
  "indexedAt",
  "author",
  "userInteractionRatio",
  "iScore",
  "tScore",
  "iScore" + "tScore" as "rating"
from
  "scoredPost"
where
  "followedBy" = ?
  and "replyParent" is null
order by "rating" desc
limit
  ?
offset
  ?
`

func lovelies(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
	fmt.Printf("[FEED] Lovelies for %s\n", session.UserDid)
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
	rows, err := database.QueryContext(ctx, loveliesQuery, session.UserDid, limit, offset)
	if err != nil {
		return output, fmt.Errorf("error executing lovelies query: %s", err)
	}
	defer rows.Close()
	var row loveliesQueryRow
	for rows.Next() {
		err := rows.Scan(
			&row.Uri,
			&row.IndexedAt,
			&row.Author,
			&row.UserInteractionRatio,
			&row.IScore,
			&row.TScore,
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
