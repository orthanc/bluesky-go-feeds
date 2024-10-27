package processor

import (
	"context"

	"github.com/bluesky-social/indigo/repomgr"
	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/following"
	"github.com/orthanc/feedgenerator/subscription"
)

type FollowProcessor struct {
	AllFollowing *following.AllFollowing
	Database     *database.Database
}

func (processor *FollowProcessor) Process(ctx context.Context, event subscription.FirehoseEvent) error {
	switch event.EventKind {
	case repomgr.EvtKindCreateRecord:
		subject := event.Record["subject"].(string)
		if processor.AllFollowing.IsUser(event.Author) {
			err := processor.AllFollowing.RecordFollow(ctx, event.Uri, event.Author, subject)
			if err != nil {
				return err
			}
		}
		if processor.AllFollowing.IsUser(subject) {
			err := processor.AllFollowing.RecordFollower(ctx, event.Uri, subject, event.Author)
			if err != nil {
				return err
			}
		}
	case repomgr.EvtKindDeleteRecord:
		if processor.AllFollowing.IsUser(event.Author) {
			return processor.AllFollowing.RemoveFollow(ctx, event.Uri)
		}
	}
	return nil
}
