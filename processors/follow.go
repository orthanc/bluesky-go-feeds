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

type FollowProcessor struct {
	AllFollowing *following.AllFollowing
	Database     *database.Database
}

func (processor *FollowProcessor) Process(ctx context.Context, event *models.Event, followUri string) error {
	switch event.Commit.Operation {
	case models.CommitOperationCreate:
		var follow bsky.GraphFollow
		if err := json.Unmarshal(event.Commit.Record, &follow); err != nil {
			return fmt.Errorf("failed to unmarshal follow: %w", err)
		}
		subject := follow.Subject
		if processor.AllFollowing.IsUser(event.Did) {
			err := processor.AllFollowing.RecordFollow(ctx, followUri, event.Did, subject)
			if err != nil {
				return err
			}
		}
		if processor.AllFollowing.IsUser(subject) {
			err := processor.AllFollowing.RecordFollower(ctx, followUri, subject, event.Did)
			if err != nil {
				return err
			}
		}
	case models.CommitOperationDelete:
		if processor.AllFollowing.IsUser(event.Did) {
			return processor.AllFollowing.RemoveFollow(ctx, followUri)
		}
	}
	return nil
}
