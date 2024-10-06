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

func (processor *FollowProcessor) Process(ctx context.Context, event subscription.FirehoseEvent) {
	switch event.EventKind {
	case repomgr.EvtKindCreateRecord:
		if processor.AllFollowing.UserDids[event.Author] {
			processor.AllFollowing.RecordFollow(event.Uri, event.Author, event.Record["subject"].(string))
		}
	case repomgr.EvtKindDeleteRecord:
		if processor.AllFollowing.UserDids[event.Author] {
			processor.AllFollowing.RemoveFollow(event.Uri)
		}
	}
}
