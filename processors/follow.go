package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/database/read"
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
			fmt.Printf("failed to unmarshal follow: %s : at://%s/%s/%s\n", err, event.Did, event.Commit.Collection, event.Commit.RKey)
			return nil
		}
		subject := follow.Subject
		interest, err := processor.Database.Queries.GetFollowingFollowData(ctx, read.GetFollowingFollowDataParams{
			FollowAuthor:  event.Did,
			FollowSubject: subject,
		})
		if err != nil {
			return fmt.Errorf("unable to load follow follow data %s", err)
		}
		if interest.FollowByUser > 0 {
			err := processor.AllFollowing.RecordFollow(ctx, followUri, event.Did, subject)
			if err != nil {
				return err
			}
		}
		if interest.FollowingUser > 0 {
			err := processor.AllFollowing.RecordFollower(ctx, followUri, subject, event.Did)
			if err != nil {
				return err
			}
		}
	case models.CommitOperationDelete:
		interest, err := processor.Database.Queries.GetFollowingFollowData(ctx, read.GetFollowingFollowDataParams{
			FollowAuthor:  event.Did,
			FollowSubject: event.Did,
		})
		if err != nil {
			return fmt.Errorf("unable to load follow follow data %s", err)
		}
		if interest.FollowByUser > 0 {
			return processor.AllFollowing.RemoveFollow(ctx, followUri)
		}
	}
	return nil
}
