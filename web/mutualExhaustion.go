package web

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
)

type mutualExhaustionQueryRow = struct {
	Uri              string
	IndexedAt        string
	InteractionCount string
}

const mutualExhaustionAlgorithmId = "3lpoa7sy3hfvj"

const mutualExhaustionQuery = `
select
  "post"."uri",
  "post"."indexedAt",
  (select count("postUri") from "userInteraction" where "userInteraction"."userDid" = ? and "userInteraction"."postUri" = post.uri) as interaction_count
from
  "post"
  inner join "following" on "post"."author" = "following"."following" and following.mutual > 0
where
  "following"."followedBy" = ?
  and interaction_count = 0
  and "post"."replyParent" is null
order by indexedAt desc
limit
  ?
offset
  ?
`

func mutualExhaustion(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
	endLog := logFeedAccess("mutualExhaustion", session)
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
	rows, err := database.QueryContext(ctx, mutualExhaustionQuery, session.UserDid, session.UserDid, limit, offset)
	if err != nil {
		return output, fmt.Errorf("error executing mutualExhaustion query: %s", err)
	}
	defer rows.Close()
	var row mutualExhaustionQueryRow
	for rows.Next() {
		err := rows.Scan(
			&row.Uri,
			&row.IndexedAt,
			&row.InteractionCount,
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
