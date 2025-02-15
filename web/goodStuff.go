package web

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
)

type goodStuffQueryRow = struct {
	Uri                  string
	IndexedAt            string
	Author               string
	ExternalUri          sql.NullString
	QuotedPostUri        sql.NullString
	UserInteractionRatio float64
	PiScore              float64
	IScore               float64
	TScore               float64
	Rating               float64
}

const goodStuffAlgorithmId = "oi8ydnb44i8y"

const goodStuffQuery = `
with
  "authorScoredPost" as (
    select
      "post"."uri",
      "post"."indexedAt",
      "post"."author",
      "post"."replyParentAuthor",
      "post"."interactionCount",
      post.external_uri as external_uri,
      post.quoted_post_uri as quoted_post_uri,
      (unixepoch ('now') - unixepoch ("indexedAt")) / (24.0 * 3600) as "timeAgo",
      (
        unixepoch ("indexedAt") - unixepoch ('now', '-7 days')
      ) / (24.0 * 3600) as "tScore",
      min(
        3000,
        (
          "post"."interactionCount" - "author"."medianInteractionCount"
        ) * 1000 / (author."medianInteractionCount" + 1)
      ) as "piScore"
    from
      "post"
      inner join "author" on "post"."author" = "author"."did"
  ),
  "interactionScoredPost" as (
    select
      "authorScoredPost"."uri",
      "authorScoredPost"."indexedAt",
      "authorScoredPost"."author",
      "authorScoredPost"."replyParentAuthor",
      authorScoredPost.external_uri as external_uri,
      authorScoredPost.quoted_post_uri as quoted_post_uri,
      "following"."userInteractionRatio",
      "following"."followedBy",
      "authorScoredPost"."piScore",
      0.00005 * "authorScoredPost"."timeAgo" * "authorScoredPost"."piScore" as "pScore",
      0.20 * "following"."userInteractionRatio" * "authorScoredPost"."timeAgo" as "iScore",
      (
        unixepoch ("indexedAt") - unixepoch ('now', '-7 days')
      ) / (24.0 * 3600) as "tScore"
    from
      "authorScoredPost"
      inner join "following" on "authorScoredPost"."author" = "following"."following"
  )
select
  "interactionScoredPost"."uri",
  "interactionScoredPost"."indexedAt",
  "interactionScoredPost"."author",
  interactionScoredPost.external_uri as external_uri,
  interactionScoredPost.quoted_post_uri as quoted_post_uri,
  "interactionScoredPost"."userInteractionRatio",
  "interactionScoredPost"."piScore",
  "interactionScoredPost"."iScore",
  "interactionScoredPost"."tScore",
  "pScore" + "tScore" + "iScore" as "rating"
from
  "interactionScoredPost"
  left join "following" as "parentFollowing" on "interactionScoredPost"."replyParentAuthor" = "parentFollowing"."following"
where
  "interactionScoredPost"."followedBy" = ?
  and (
    "replyParentAuthor" is null
    or "parentFollowing"."followedBy" = ?
  )
  and "indexedAt" < ?
union all
select
  "authorScoredPost"."uri",
  "authorScoredPost"."indexedAt",
  "authorScoredPost"."author",
  authorScoredPost.external_uri as external_uri,
  authorScoredPost.quoted_post_uri as quoted_post_uri,
  0 as "userInteractionRatio",
  0 as "piScore",
  0 as "iScore",
  "authorScoredPost"."tScore",
  "authorScoredPost"."tScore" as "rating"
from
  "authorScoredPost"
where
  "authorScoredPost"."author" = ?
  and "indexedAt" < ?
order by
  "rating" desc
limit
  ?
offset
  ?
`

func goodStuff(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
	endLog := logFeedAccess("goodstuff", session)
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
	now := time.Now().UTC().Format(time.RFC3339)
	rows, err := database.QueryContext(ctx, goodStuffQuery, session.UserDid, session.UserDid, now, session.UserDid, now, limit, offset)
	if err != nil {
		return output, fmt.Errorf("error executing good stuff query: %s", err)
	}
	defer rows.Close()
	var row goodStuffQueryRow
	for rows.Next() {
		err := rows.Scan(
			&row.Uri,
			&row.IndexedAt,
			&row.Author,
			&row.ExternalUri,
			&row.QuotedPostUri,
			&row.UserInteractionRatio,
			&row.PiScore,
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
