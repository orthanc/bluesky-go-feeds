package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/database/read"
)

type RepostProcessor struct {
	Database     *database.Database
	PostUrisChan chan ReferencedPost
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

		interest, err := processor.Database.Queries.GetRepostFollowData(ctx, read.GetRepostFollowDataParams{
			PostAuthor:   postAuthor,
			RepostAuthor: event.Did,
		})
		if err != nil {
			return fmt.Errorf("unable to load like follow data %s", err)
		}
		// Quick return for reposts that we have no interest in so that we can avoid starting transactions for them
		// authorFollowedBy := processor.AllFollowing.FollowedBy(event.Did)
		// authorIsFollowed := len(authorFollowedBy) > 0
		if !(interest.PostByAuthor > 0 || interest.RepostByAuthor >= 0) {
			// ||
			// authorIsFollowed) {
			return nil
		}

		if interest.RepostByAuthor > 0 {
			processor.PostUrisChan <- ReferencedPost{
				PostUri:           postUri,
				SourceEventAuthor: event.Did,
				SourceIndexedAt:   time.Now().UTC().Format(time.RFC3339),
				SourcePostUri:     postUri,
				ReferenceType:     RepostedPost,
			}
		}

		updates, tx, err := processor.Database.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		// indexedAt := time.Now().UTC().Format(time.RFC3339)
		if interest.PostByAuthor > 0 {
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
