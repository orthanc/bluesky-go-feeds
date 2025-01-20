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
)

type ListItemProcessor struct {
	database     *database.Database
	listUri string
}

func NewListItemProcessor(database *database.Database, listUri string) *ListItemProcessor {
	return &ListItemProcessor{
		database: database,
		listUri: listUri,
	}
}

func (processor *ListItemProcessor) Process(ctx context.Context, event *models.Event, listitemUri string) error {
	switch event.Commit.Operation {
	case models.CommitOperationCreate:
		var listitem bsky.GraphListitem
		if err := json.Unmarshal(event.Commit.Record, &listitem); err != nil {
			fmt.Printf("failed to unmarshal listitem: %s : at://%s/%s/%s\n", err, event.Did, event.Commit.Collection, event.Commit.RKey)
			return nil
		}
		if listitem.List != processor.listUri {
			return nil;
		}

		err := processor.database.Updates.SaveListMembership(ctx, writeSchema.SaveListMembershipParams{
			ListUri: listitem.List,
			MemberDid: listitem.Subject,
			LastRecorded: time.Now().UTC().Format(time.RFC3339),
		})
		if err != nil {
			return fmt.Errorf("unable to save list item %s", err)
		}
		return nil;
	}
	return nil
}
