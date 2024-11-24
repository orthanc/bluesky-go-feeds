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
	writeSchema "github.com/orthanc/feedgenerator/database/write"
)

type LikeProcessor struct {
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

		interest, err := processor.Database.Queries.GetLikeFollowData(ctx, read.GetLikeFollowDataParams{
			PostAuthor: postAuthor,
			LikeAuthor: event.Did,
		})
		if err != nil {
			return fmt.Errorf("unable to load like follow data %s", err)
		}
		// Quick return for likes that we have no interest in so that we can avoid starting transactions for them
		// authorFollowedBy := processor.AllFollowing.FollowedBy(event.Did)
		// authorIsFollowed := len(authorFollowedBy) > 0
		if !(interest.PostByUser > 0 || interest.PostByAuthor > 0) {
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
		if interest.PostByAuthor > 0 {
			err := updates.IncrementPostLike(ctx, postUri)
			if err != nil {
				return err
			}

			if interest.LikeByUser > 0 && event.Did != postAuthor {
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
		if interest.PostByUser > 0 {
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
