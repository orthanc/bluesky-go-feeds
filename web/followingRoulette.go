package web

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
)

type followingRouletteQueryRow = struct {
	Rnd               string
	Author            string
	Uri               string
	ReplyParentAuthor sql.NullString
	InteractionCount  float64
}

const followingRouletteAlgorithmId = "3lbjni6x7fmkz"

const followingRouletteQuery = `
SELECT hex(randomblob(16)) rnd, following.following, post.uri, post.replyParentAuthor, max(post.interactionCount)
FROM following
JOIN post on post.author = following.following
LEFT JOIN following as parentFollowing on post.replyParentAuthor = parentFollowing.following AND parentFollowing.followedBy = ?
WHERE following.followedBy = ?
AND post.indexedAt > ?
AND (post.replyParentAuthor IS NULL OR parentFollowing.followedBy IS NOT NULL)
GROUP BY following.following
ORDER BY rnd
LIMIT ?
`

func followingRoulette(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
	endLog := logFeedAccess("followingRoulette", session)
	defer endLog()
	output := bsky.FeedGetFeedSkeleton_Output{
		Feed: make([]*bsky.FeedDefs_SkeletonFeedPost, 0, limit),
	}
	oneDayAgo := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	rows, err := database.QueryContext(ctx, followingRouletteQuery, session.UserDid, session.UserDid, oneDayAgo, limit)
	if err != nil {
		return output, fmt.Errorf("error executing followingRoulette query: %s", err)
	}
	defer rows.Close()
	var row followingRouletteQueryRow
	for rows.Next() {
		err := rows.Scan(
			&row.Rnd,
			&row.Author,
			&row.Uri,
			&row.ReplyParentAuthor,
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
	if len(output.Feed) == limit {
		nextCursor := "more"
		output.Cursor = &nextCursor
	}
	return output, nil
}
