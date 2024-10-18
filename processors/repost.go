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

type RepostProcessor struct {
	AllFollowing *following.AllFollowing
	Database     *database.Database
}

func (processor *RepostProcessor) Process(ctx context.Context, event subscription.FirehoseEvent) error {
	switch event.EventKind {
	case repomgr.EvtKindCreateRecord:
		postUri := event.Record["subject"].(map[string]any)["uri"].(string)
		postAuthor := getAuthorFromPostUri(postUri)

		// Quick return for likes that we have no interest in so that we can avoid starting transactions for them
		authorFollowedBy := processor.AllFollowing.FollowedBy(event.Author)
		authorIsFollowed := len(authorFollowedBy) > 0
		if !(processor.AllFollowing.IsFollowed(postAuthor) ||
			authorIsFollowed) {
			return nil
		}

		updates, tx, err := processor.Database.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		indexedAt := time.Now().UTC().Format(time.RFC3339)
		if processor.AllFollowing.IsFollowed(postAuthor) {
			err := updates.IncrementPostRepost(ctx, postUri)
			if err != nil {
				return err
			}
		}

		for _, followedBy := range authorFollowedBy {
			err := updates.SavePostRepostedByFollowing(ctx, writeSchema.SavePostRepostedByFollowingParams{
				User: followedBy,
				Uri: postUri,
				Author: postAuthor,
				IndexedAt: indexedAt,
			})
			if err != nil {
				return err
			}
		}
		tx.Commit()
	}
	return nil
}
