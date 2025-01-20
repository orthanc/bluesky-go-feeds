package web

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
)

type followerRouletteQueryRow = struct {
	Rnd               string
	Author            string
	Uri               string
	ReplyParentAuthor sql.NullString
	InteractionCount  float64
}

const followerRouletteAlgorithmId = "3lazy424fbmri"

const followerRouletteQuery = `
SELECT hex(randomblob(16)) rnd, follower.followed_by, post.uri, post.replyParentAuthor, max(post.interactionCount)
FROM follower
JOIN post on post.author = follower.followed_by
LEFT JOIN following on post.replyParentAuthor = following.following AND following.followedBy = ?
WHERE follower.following = ?
AND follower.mutual = 0
AND post.indexedAt > ?
AND (post.replyParentAuthor IS NULL OR following.followedBy IS NOT NULL)
GROUP BY follower.followed_by
ORDER BY rnd
LIMIT ?
`

func followerRoulette(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
	endLog := logFeedAccess("followerRoulette", session)
	defer endLog()
	output := bsky.FeedGetFeedSkeleton_Output{
		Feed: make([]*bsky.FeedDefs_SkeletonFeedPost, 0, limit),
	}

	isFollowFarmerResult, err := database.Queries.IsOnList(ctx, schema.IsOnListParams{
		ListUri:   os.Getenv("FOLLOW_FARMERS_LIST"),
		MemberDid: session.UserDid,
	})
	if err != nil {
		return output, err
	}
	if isFollowFarmerResult > 0 {
		output.Feed = append(output.Feed, &bsky.FeedDefs_SkeletonFeedPost{
			Post: "at://did:plc:crngjmsdh3zpuhmd5gtgwx6q/app.bsky.feed.post/3lg4lxcvzlk2b",
		})
		output.Feed = append(output.Feed, &bsky.FeedDefs_SkeletonFeedPost{
			Post: "at://did:plc:crngjmsdh3zpuhmd5gtgwx6q/app.bsky.feed.post/3lg4lxcwgbs2b",
		})
		output.Feed = append(output.Feed, &bsky.FeedDefs_SkeletonFeedPost{
			Post: "at://did:plc:crngjmsdh3zpuhmd5gtgwx6q/app.bsky.feed.post/3lg4lxcwj7k2b",
		})
		output.Feed = append(output.Feed, &bsky.FeedDefs_SkeletonFeedPost{
			Post: "at://did:plc:crngjmsdh3zpuhmd5gtgwx6q/app.bsky.feed.post/3lg4lxcwk6s2b",
		})
		output.Feed = append(output.Feed, &bsky.FeedDefs_SkeletonFeedPost{
			Post: "at://did:plc:crngjmsdh3zpuhmd5gtgwx6q/app.bsky.feed.post/3lg4lxcwm5c2b",
		})
		return output, nil
	}

	oneDayAgo := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	rows, err := database.QueryContext(ctx, followerRouletteQuery, session.UserDid, session.UserDid, oneDayAgo, limit)
	if err != nil {
		return output, fmt.Errorf("error executing followerRoulette query: %s", err)
	}
	defer rows.Close()
	var row followerRouletteQueryRow
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
