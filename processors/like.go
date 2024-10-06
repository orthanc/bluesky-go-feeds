package processor

import (
	"context"
	"time"

	"github.com/bluesky-social/indigo/repomgr"
	"github.com/orthanc/feedgenerator/database"
	writeSchema "github.com/orthanc/feedgenerator/database/write"
	"github.com/orthanc/feedgenerator/following"
	"github.com/orthanc/feedgenerator/subscription"
)

type LikeProcessor struct {
	AllFollowing *following.AllFollowing
	Database     *database.Database
}

func (processor *LikeProcessor) Process(ctx context.Context, event subscription.FirehoseEvent) error {
	switch event.EventKind {
	case repomgr.EvtKindCreateRecord:
		postUri := event.Record["subject"].(map[string]any)["uri"].(string)
		postAuthor := getAuthorFromPostUri(postUri)

		// Quick return for likes that we have no interest in so that we can avoid starting transactions for them
		if !(processor.AllFollowing.IsUser(postAuthor) ||
			processor.AllFollowing.IsFollowed(postAuthor)) {
			return nil
		}

		updates, tx, err := processor.Database.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		indexedAt := time.Now().Format(time.RFC3339)
		if processor.AllFollowing.IsFollowed(postAuthor) {
			err := updates.IncrementPostLike(ctx, postUri)
			if err != nil {
				return err
			}

			if processor.AllFollowing.IsUser(event.Author) && event.Author != postAuthor {
				err := updates.SaveUserInteraction(ctx, writeSchema.SaveUserInteractionParams{
					InteractionUri: event.Uri,
					AuthorDid:      postAuthor,
					UserDid:        event.Author,
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
				InteractionUri:       event.Uri,
				InteractionAuthorDid: event.Author,
				UserDid:              postAuthor,
				PostUri:              postUri,
				Type:                 "like",
				IndexedAt:            indexedAt,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
