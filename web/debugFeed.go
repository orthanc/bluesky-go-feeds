package web

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
)

type debugQueryRow = struct {
	Uri                  string
	IndexedAt            string
	Author               string
}

const debugAlgorithmId = "replies-foll"

const debugQuery = `
select
  post."uri",
  "indexedAt",
  "author"
from post
inner join "following" on "post"."author" = "following"."following"
left join "following" as "parentFollowing" on "post"."replyParentAuthor" = "parentFollowing"."following" and parentFollowing.followedBy = ?
where
  following.followedBy = ?
	and (
		"replyParent" is null
		or parentFollowing.following is not null
	)
union 
select
  post."uri",
  "indexedAt",
  "author"
from post
where post.author = ?
order by indexedAt desc
	limit
  ?
offset
  ?
`

func debugFeed(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
	endLog := logFeedAccess("debug", session)
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
	rows, err := database.QueryContext(ctx, debugQuery, session.UserDid, session.UserDid, session.UserDid, limit, offset)
	if err != nil {
		return output, fmt.Errorf("error executing debug query: %s", err)
	}
	defer rows.Close()
	var row debugQueryRow
	for rows.Next() {
		err := rows.Scan(
			&row.Uri,
			&row.IndexedAt,
			&row.Author,
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
