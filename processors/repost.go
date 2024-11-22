package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/following"
)

type RepostProcessor struct {
	AllFollowing *following.AllFollowing
	Database     *database.Database
}

func (processor *RepostProcessor) Process(ctx context.Context, event *models.Event, repostUri string) error {
	switch event.Commit.Operation {
	case models.CommitOperationCreate:
		var repost bsky.FeedRepost
		if err := json.Unmarshal(event.Commit.Record, &repost); err != nil {
			fmt.Printf("failed to unmarshal repost: %s : at://%s/%s/%s\n", err, event.Did, event.Commit.Collection, event.Commit.RKey)
			return nil
		}
		postUri := repost.Subject.Uri
		postAuthor := getAuthorFromPostUri(postUri)
		if postAuthor == "" {
			return nil
		}

		// Quick return for likes that we have no interest in so that we can avoid starting transactions for them
		// authorFollowedBy := processor.AllFollowing.FollowedBy(event.Did)
		// authorIsFollowed := len(authorFollowedBy) > 0
		if !(processor.AllFollowing.IsAuthor(postAuthor)) { 
			// ||
			// authorIsFollowed) {
			return nil
		}

		updates, tx, err := processor.Database.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		// indexedAt := time.Now().UTC().Format(time.RFC3339)
		if processor.AllFollowing.IsAuthor(postAuthor) {
			err := updates.IncrementPostRepost(ctx, postUri)
			if err != nil {
				return err
			}
		}

		// for _, followedBy := range authorFollowedBy {
		// 	err := updates.SavePostRepostedByFollowing(ctx, writeSchema.SavePostRepostedByFollowingParams{
		// 		User:      followedBy,
		// 		Uri:       postUri,
		// 		Author:    postAuthor,
		// 		IndexedAt: indexedAt,
		// 	})
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		tx.Commit()
	}
	return nil
}
