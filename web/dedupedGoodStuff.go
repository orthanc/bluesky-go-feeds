package web

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
)

const dedupedGoodStuffAlgorithmId = "3li6obdj4x5rp"

const deduplicationWindowSize = 5

type bufferEntry struct {
	rowNum int
	post   *goodStuffQueryRow
}
type postBuffer struct {
	startIndex int
	nextIndex  int
	lastIndex  int
	used       int
	posts      [deduplicationWindowSize]bufferEntry
}

func (buffer *postBuffer) removeIndex(index int) {
	if buffer.used == 0 {
		return
	}
	current := index
	for current != buffer.lastIndex {
		next := (current + 1) % len(buffer.posts)
		buffer.posts[current] = buffer.posts[next]
		current = next
	}
	buffer.used = buffer.used - 1
	buffer.nextIndex = (buffer.startIndex + buffer.used) % len(buffer.posts)
	buffer.lastIndex = (buffer.nextIndex + len(buffer.posts) - 1) % len(buffer.posts)
}

func (buffer *postBuffer) pushPost(post goodStuffQueryRow, rowNum int) bufferEntry {
	fmt.Printf("SKIPPING ---  %s : %s : %s : %d : %t : %t \n", post.Uri, post.QuotedPostUri.String, post.ExternalUri.String, buffer.used, post.QuotedPostUri.Valid, post.ExternalUri.Valid)
	if buffer.used > 0 && (post.ExternalUri.Valid || post.QuotedPostUri.Valid) {
		current := buffer.startIndex
		// Remove any posts from the buffer that have the same external uri or quote as the one we're adding
		for {
			fmt.Printf("\tSKIPPING POST: QUOTED URI %s: %s == %s\n", post.Uri, post.QuotedPostUri.String, buffer.posts[current].post.QuotedPostUri.String)
			if (post.ExternalUri.Valid && (
					post.ExternalUri.String == buffer.posts[current].post.ExternalUri.String ||
					post.ExternalUri.String == buffer.posts[current].post.RootExternalUri.String)) ||
				(post.RootExternalUri.Valid && (
					post.RootExternalUri.String == buffer.posts[current].post.ExternalUri.String ||
					post.RootExternalUri.String == buffer.posts[current].post.RootExternalUri.String)) ||
				(post.QuotedPostUri.Valid && (
					post.QuotedPostUri.String == buffer.posts[current].post.QuotedPostUri.String ||
					post.QuotedPostUri.String == buffer.posts[current].post.RootQuotedPostUri.String ||
					post.QuotedPostUri.String == buffer.posts[current].post.Uri)) ||
				(post.RootQuotedPostUri.Valid && (
					post.RootQuotedPostUri.String == buffer.posts[current].post.QuotedPostUri.String ||
					post.RootQuotedPostUri.String == buffer.posts[current].post.RootQuotedPostUri.String)) ||
				(post.Uri == buffer.posts[current].post.QuotedPostUri.String) {
				fmt.Printf("SKIPPING POST %s\n", buffer.posts[current].post.Uri)
				buffer.removeIndex(current)
			} else {
				current = (current + 1) % len(buffer.posts)
			}
			if current == buffer.nextIndex {
				break;
			}
		}
	}
	var entryToReturn bufferEntry
	if buffer.used == len(buffer.posts) {
		entryToReturn = buffer.posts[buffer.nextIndex]
		buffer.startIndex = (buffer.startIndex + 1) % len(buffer.posts)
		buffer.posts[buffer.nextIndex].post = &post
		buffer.posts[buffer.nextIndex].rowNum = rowNum
		buffer.lastIndex = buffer.nextIndex
		buffer.nextIndex = buffer.startIndex
	} else {
		buffer.posts[buffer.nextIndex].post = &post
		buffer.posts[buffer.nextIndex].rowNum = rowNum
		buffer.lastIndex = buffer.nextIndex
		buffer.nextIndex = (buffer.nextIndex + 1) % len(buffer.posts)
		buffer.used++
	}
	return entryToReturn
}

func dedupedGoodStuff(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error) {
	endLog := logFeedAccess("dedupedGoodStuff", session)
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
	loadLimit := 2 * limit
	rows, err := database.QueryContext(ctx, goodStuffQuery, session.UserDid, session.UserDid, now, session.UserDid, now, loadLimit, offset)
	if err != nil {
		return output, fmt.Errorf("error executing deduped good stuff query: %s", err)
	}
	defer rows.Close()
	var row goodStuffQueryRow
	buffer := &postBuffer{}
	rowNum := offset
	lastRownum := 0
	for rows.Next() && len(output.Feed) < limit {
		err := rows.Scan(
			&row.Uri,
			&row.IndexedAt,
			&row.Author,
			&row.ExternalUri,
			&row.QuotedPostUri,
			&row.RootExternalUri,
			&row.RootQuotedPostUri,
			&row.UserInteractionRatio,
			&row.PiScore,
			&row.IScore,
			&row.TScore,
			&row.Rating,
		)
		if err != nil {
			return output, err
		}
		toReturn := buffer.pushPost(row, rowNum)
		if toReturn.post != nil {
			output.Feed = append(output.Feed, &bsky.FeedDefs_SkeletonFeedPost{
				Post: toReturn.post.Uri,
			})
			lastRownum = toReturn.rowNum
		}
		rowNum++
	}
	if len(output.Feed) > 0 {
		nextCursor := strconv.Itoa(lastRownum + 1)
		output.Cursor = &nextCursor
	}
	return output, nil
}
