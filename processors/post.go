package processor

import (
	"context"
	"time"

	"github.com/orthanc/feedgenerator/database"
	writeSchema "github.com/orthanc/feedgenerator/database/write"
	"github.com/orthanc/feedgenerator/subscription"
)

type PostProcessor struct {
	Ctx context.Context
	Database *database.Database
}

func (processor *PostProcessor) Process(event subscription.FirehoseEvent) {
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
	updates, tx, error := processor.Database.BeginTx(processor.Ctx)
	if error != nil {
		panic(error)
	}
	defer tx.Rollback()
	updates.SavePost(processor.Ctx, writeSchema.SavePostParams{
		Uri:               event.Uri,
		Author:            event.Author,
		ReplyParent:       toNullString(replyParent),
		ReplyParentAuthor: toNullString(replyParentAuthor),
		ReplyRoot:         toNullString(replyRoot),
		ReplyRootAuthor:   toNullString(replyRootAuthor),
		IndexedAt:         time.Now().Format(time.RFC3339),
		DirectReplyCount:  0,
		InteractionCount:  0,
		LikeCount:         0,
		ReplyCount:        0,
	})
	tx.Commit()

}