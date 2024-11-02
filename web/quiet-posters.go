package web

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
)

type quietPostersQueryRow = struct {
	Uri        string
	IndexedAt  string
	Author     string
	BoostScore sql.NullFloat64
	TScore     float64
}

const quietPostersAlgorithmId = "3l6bvwqxjuvit"

const quietPostersQuery = `
with
	"maxPostCount" as (
		select max("postCount") as cnt, ifnull(log(max("postCount")),1) as logCnt
		from "following"
		inner join "author" on "following"."following" = "author"."did"
		where "following"."followedBy" = ?
	),
	"authorBoost" as (
		select "author"."did", 7 - 14 * ifnull(log("postCount"),0) / (select "logCnt" from "maxPostCount") as "boostScore"
		from "following"
		inner join "author" on "following"."following" = "author"."did"
		where "following"."followedBy" = ?
	)
select
  "post"."uri",
  "indexedAt",
  "author",
  "boostScore",
	(
		unixepoch ("indexedAt") - unixepoch ('now', '-7 days')
	) / (24.0 * 3600) as "tScore"
	from "post"
	inner join "authorBoost" on "post"."author" = "authorBoost"."did"
  left join "following" as "parentFollowing" on "post"."replyParentAuthor" = "parentFollowing"."following"
	where (
    "replyParentAuthor" is null
    or "parentFollowing"."followedBy" = ?
  )
	and post."indexedAt" >= ?
	order by "tScore" + "boostScore" / 7 desc
limit
  ?
offset
  ?
`

func quietPosters(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
	fmt.Printf("[FEED] Quiet People for %s since %s\n", session.UserDid, session.PostsSince)
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
	rows, err := database.QueryContext(ctx, quietPostersQuery, session.UserDid, session.UserDid, session.UserDid, session.PostsSince, limit, offset)
	if err != nil {
		return output, fmt.Errorf("error executing quietPosters query: %s", err)
	}
	defer rows.Close()
	var row quietPostersQueryRow
	for rows.Next() {
		err := rows.Scan(
			&row.Uri,
			&row.IndexedAt,
			&row.Author,
			&row.BoostScore,
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
