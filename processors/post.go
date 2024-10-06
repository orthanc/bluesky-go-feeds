package processor

import (
	"context"
	"time"

	"github.com/orthanc/feedgenerator/database"
	writeSchema "github.com/orthanc/feedgenerator/database/write"
	"github.com/orthanc/feedgenerator/subscription"
)

type PostProcessor struct {
	Database *database.Database
}

func (processor *PostProcessor) Process(ctx context.Context, event subscription.FirehoseEvent) {
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
	updates, tx, error := processor.Database.BeginTx(ctx)
	if error != nil {
		panic(error)
	}
	defer tx.Rollback()
	updates.SavePost(ctx, writeSchema.SavePostParams{
		Uri:               event.Uri,
		Author:            event.Author,
		ReplyParent:       database.ToNullString(replyParent),
		ReplyParentAuthor: database.ToNullString(replyParentAuthor),
		ReplyRoot:         database.ToNullString(replyRoot),
		ReplyRootAuthor:   database.ToNullString(replyRootAuthor),
		IndexedAt:         time.Now().Format(time.RFC3339),
		DirectReplyCount:  0,
		InteractionCount:  0,
		LikeCount:         0,
		ReplyCount:        0,
	})
	tx.Commit()

}
