package web

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
)

type youMightLikeQueryRow = struct {
	Uri    string
	TScore string
}

const youMightLikeAlgorithmId = "3l6te6ptx5l3c"

const youMightLikeQuery = `
select
  post.uri,
  (
    unixepoch ("indexed_at") - unixepoch ('now', '-7 days')
  ) / (24.0 * 3600) as t_score
from
  post_interacted_by_followed as post
  left join post_interacted_by_followed_author as author on post.author = author.author and post.user = author.user
where
  post.user = ?
  and post.author <> post.user
	and author.followed = 0
order by
  10 * post.followed_interaction_count + author.followed_interaction_count + 100 * t_score DESC
limit
  ?
offset
  ?
`

func youMightLike(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
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
	rows, err := database.QueryContext(ctx, youMightLikeQuery, session.UserDid, limit, offset)
	if err != nil {
		return output, fmt.Errorf("error executing youMightLike query: %s", err)
	}
	defer rows.Close()
	var row youMightLikeQueryRow
	for rows.Next() {
		err := rows.Scan(
			&row.Uri,
			&row.TScore,
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
