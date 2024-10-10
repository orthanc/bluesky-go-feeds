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

type PostProcessor struct {
	AllFollowing *following.AllFollowing
	Database     *database.Database
}

func (processor *PostProcessor) Process(ctx context.Context, event subscription.FirehoseEvent) error {
	switch event.EventKind {
	case repomgr.EvtKindCreateRecord:
		var replyParent, replyParentAuthor, replyRoot, replyRootAuthor string
		reply := event.Record["reply"]
		if reply != nil {
			parent := (reply.(map[string]any))["parent"]
			if parent != nil {
				replyParent = parent.(map[string]any)["uri"].(string)
				replyParentAuthor = getAuthorFromPostUri(replyParent)
			}
			root := (reply.(map[string]any))["root"]
			if parent != nil {
				replyRoot = root.(map[string]any)["uri"].(string)
				replyRootAuthor = getAuthorFromPostUri(replyRoot)
			}
		}

		// Quick return for posts that we have no interest in so that we can avoid starting transactions for them
		if !(processor.AllFollowing.IsFollowed(event.Author) ||
			processor.AllFollowing.IsFollowed(replyParentAuthor) ||
			processor.AllFollowing.IsUser(replyParentAuthor) ||
			processor.AllFollowing.IsFollowed(replyRootAuthor) ||
			processor.AllFollowing.IsUser(replyRootAuthor)) {
			return nil
		}

		updates, tx, err := processor.Database.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		indexedAt := time.Now().UTC().Format(time.RFC3339)
		if processor.AllFollowing.IsFollowed(event.Author) {
			err := updates.SavePost(ctx, writeSchema.SavePostParams{
				Uri:               event.Uri,
				Author:            event.Author,
				ReplyParent:       database.ToNullString(replyParent),
				ReplyParentAuthor: database.ToNullString(replyParentAuthor),
				ReplyRoot:         database.ToNullString(replyRoot),
				ReplyRootAuthor:   database.ToNullString(replyRootAuthor),
				IndexedAt:         indexedAt,
				DirectReplyCount:  0,
				InteractionCount:  0,
				LikeCount:         0,
				ReplyCount:        0,
			})
			if err != nil {
				return err
			}
		}
		if processor.AllFollowing.IsFollowed(replyParentAuthor) {
			err := updates.IncrementPostDirectReply(ctx, event.Uri)
			if err != nil {
				return err
			}

			if processor.AllFollowing.IsUser(event.Author) && event.Author != replyParentAuthor {
				err := updates.SaveUserInteraction(ctx, writeSchema.SaveUserInteractionParams{
					InteractionUri: event.Uri,
					AuthorDid:      replyParentAuthor,
					UserDid:        event.Author,
					PostUri:        replyParent,
					Type:           "reply",
					IndexedAt:      indexedAt,
				})
				if err != nil {
					return err
				}
			}
		}
		if processor.AllFollowing.IsUser(replyParentAuthor) && event.Author != replyParentAuthor {
			// Someone replying a post by one of the users
			err := updates.SaveInteractionWithUser(ctx, writeSchema.SaveInteractionWithUserParams{
				InteractionUri:       event.Uri,
				InteractionAuthorDid: event.Author,
				UserDid:              replyParentAuthor,
				PostUri:              replyParent,
				Type:                 "reply",
				IndexedAt:            indexedAt,
			})
			if err != nil {
				return err
			}
		}

		// We don't want to double process direct replies so everything after this only applies if
		// the reply parent and reply root are different
		if replyParent != replyRoot {
			if processor.AllFollowing.IsFollowed(replyRootAuthor) {
				err := updates.IncrementPostIndirectReply(ctx, event.Uri)
				if err != nil {
					return err
				}

				if processor.AllFollowing.IsUser(event.Author) && event.Author != replyRootAuthor {
					err := updates.SaveUserInteraction(ctx, writeSchema.SaveUserInteractionParams{
						InteractionUri: event.Uri,
						AuthorDid:      replyRootAuthor,
						UserDid:        event.Author,
						PostUri:        replyRoot,
						Type:           "threadReply",
						IndexedAt:      indexedAt,
					})
					if err != nil {
						return err
					}
				}
			}
			if processor.AllFollowing.IsUser(replyRootAuthor) && event.Author != replyRootAuthor {
				// Someone replying a post by one of the users
				err := updates.SaveInteractionWithUser(ctx, writeSchema.SaveInteractionWithUserParams{
					InteractionUri:       event.Uri,
					InteractionAuthorDid: event.Author,
					UserDid:              replyRootAuthor,
					PostUri:              replyRoot,
					Type:                 "threadReply",
					IndexedAt:            indexedAt,
				})
				if err != nil {
					return err
				}
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
	}
	return nil
}
