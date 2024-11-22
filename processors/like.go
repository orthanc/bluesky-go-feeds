package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/orthanc/feedgenerator/database"
	writeSchema "github.com/orthanc/feedgenerator/database/write"
	"github.com/orthanc/feedgenerator/following"
)

type LikeProcessor struct {
	AllFollowing *following.AllFollowing
	Database     *database.Database
}

func (processor *LikeProcessor) Process(ctx context.Context, event *models.Event, likeUri string) error {
	switch event.Commit.Operation {
	case models.CommitOperationCreate:
	var like bsky.FeedLike
		if err := json.Unmarshal(event.Commit.Record, &like); err != nil {
			fmt.Printf("failed to unmarshal like: %s : at://%s/%s/%s\n", err, event.Did, event.Commit.Collection, event.Commit.RKey)
			return nil
		}
		postUri := like.Subject.Uri
		postAuthor := getAuthorFromPostUri(postUri)
		if postAuthor == "" {
			return nil
		}

		// Quick return for likes that we have no interest in so that we can avoid starting transactions for them
		// authorFollowedBy := processor.AllFollowing.FollowedBy(event.Did)
		// authorIsFollowed := len(authorFollowedBy) > 0
		if !(processor.AllFollowing.IsUser(postAuthor) ||
			processor.AllFollowing.IsAuthor(postAuthor)) {
			// ||
			// authorIsFollowed) {
			return nil
		}

		updates, tx, err := processor.Database.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		indexedAt := time.Now().UTC().Format(time.RFC3339)
		if processor.AllFollowing.IsAuthor(postAuthor) {
			err := updates.IncrementPostLike(ctx, postUri)
			if err != nil {
				return err
			}

			if processor.AllFollowing.IsUser(event.Did) && event.Did != postAuthor {
				err := updates.SaveUserInteraction(ctx, writeSchema.SaveUserInteractionParams{
					InteractionUri: likeUri,
					AuthorDid:      postAuthor,
					UserDid:        event.Did,
					PostUri:        postUri,
					Type:           "like",
					IndexedAt:      indexedAt,
				})
				if err != nil {
					return err
				}
			}
		}
		if processor.AllFollowing.IsUser(postAuthor) {
			// Someone liking a post by one of the users
			err := updates.SaveInteractionWithUser(ctx, writeSchema.SaveInteractionWithUserParams{
				InteractionUri:       likeUri,
				InteractionAuthorDid: event.Did,
				UserDid:              postAuthor,
				PostUri:              postUri,
				Type:                 "like",
				IndexedAt:            indexedAt,
			})
			if err != nil {
				return err
			}
		}

		// for _, followedBy := range authorFollowedBy {
		// 	err := updates.SavePostLikedByFollowing(ctx, writeSchema.SavePostLikedByFollowingParams{
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
