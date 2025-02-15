package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/database/read"
	writeSchema "github.com/orthanc/feedgenerator/database/write"
)

type PostProcessor struct {
	Database     *database.Database
	PostersMadness *PostersMadness
}

func (processor *PostProcessor) Process(ctx context.Context, event *models.Event, postUri string) error {
	switch event.Commit.Operation {
	case models.CommitOperationCreate:
		var post bsky.FeedPost
		if err := json.Unmarshal(event.Commit.Record, &post); err != nil {
			fmt.Printf("failed to unmarshal post: %s : at://%s/%s/%s\n", err, event.Did, event.Commit.Collection, event.Commit.RKey)
			return nil
		}

		var replyParent, replyParentAuthor, replyRoot, replyRootAuthor, externalUri, quotedPostUri string
		if post.Reply != nil {
			if post.Reply.Parent != nil {
				replyParent = post.Reply.Parent.Uri
				replyParentAuthor = getAuthorFromPostUri(replyParent)
			}
			if post.Reply.Root != nil {
				replyRoot = post.Reply.Root.Uri
				replyRootAuthor = getAuthorFromPostUri(replyRoot)
			}
		}

		if post.Embed != nil {
			if post.Embed.EmbedExternal != nil && post.Embed.EmbedExternal.External != nil {
				externalUri = post.Embed.EmbedExternal.External.Uri;
			}
			if post.Embed.EmbedRecord != nil && post.Embed.EmbedRecord.Record != nil {
				quotedPostUri = post.Embed.EmbedRecord.Record.Uri;
			}
			if post.Embed.EmbedRecordWithMedia != nil && post.Embed.EmbedRecordWithMedia.Record != nil && post.Embed.EmbedRecordWithMedia.Record.Record != nil {
				quotedPostUri = post.Embed.EmbedRecordWithMedia.Record.Record.Uri
			}
		}

		if replyParentAuthor != "" {
			err := processor.PostersMadness.PostersMadnessInteraction(ctx, event.Did, replyParentAuthor);
			if err != nil {
				return err
			}
		}

		interest, err := processor.Database.Queries.GetPostFollowData(ctx, read.GetPostFollowDataParams{
			PostAuthor: event.Did,
			ReplyParentAuthor: replyParentAuthor,
			ReplyRootAuthor: replyRootAuthor,
		})
		if err != nil {
			return fmt.Errorf("unable to load post follow data %s", err)
		}

		// Quick return for posts that we have no interest in so that we can avoid starting transactions for them
		// authorFollowedBy := processor.AllFollowing.FollowedBy(event.Did)
		if !(interest.PostByAuthor > 0 ||
			interest.ReplyToAuthor > 0 ||
			interest.ReplyToUser > 0 ||
			interest.ReplyToThreadAuthor > 0 ||
			interest.ReplyToThreadUser > 0 ||
			interest.PostersMadnessSymptomatic > 0) {
			return nil
		}
		now := time.Now().UTC().Add(time.Minute)
		indexAsDate := now
		rawCreatedAt := post.CreatedAt
		createdAt, err := time.Parse(time.RFC3339, rawCreatedAt)
		if err != nil {
			fmt.Printf("Ignoring unparsable create date %s on post by %s, using %s instead\n", rawCreatedAt, event.Did, now)
			indexAsDate = now
		}
		indexAsDate = createdAt.UTC();
		if now.Before(indexAsDate) {
			fmt.Printf("Ignoring future create date %s on post by %s, using %s instead\n", indexAsDate, event.Did, now)
			indexAsDate = now
		}
		if indexAsDate.Before(now.Add(-7 * 24 * time.Hour)) {
			fmt.Printf("Dropping post by %s with create date %s, more than 7 days ago\n", event.Did, indexAsDate)
			return nil
		}
		
		updates, tx, err := processor.Database.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		indexedAt := indexAsDate.Format(time.RFC3339)
		if interest.PostByAuthor > 0 || interest.PostersMadnessSymptomatic > 0 {
			postIndexedAt := indexAsDate
			// if event.Did == replyParentAuthor && event.Did == replyRootAuthor {
			// 	parentPostDates, _ := processor.Database.Queries.GetPostDates(ctx, replyParent)
			// 	if parentPostDates.IndexedAt != "" {
			// 		parentIndexedAt, err := time.Parse(time.RFC3339, parentPostDates.IndexedAt)
			// 		if err == nil {
			// 			minIndexedAt := parentIndexedAt.Add(30 * time.Second)
			// 			if postIndexedAt.Before(minIndexedAt) {
			// 				fmt.Printf("Delaying thread post by %s from %s to %s\n", event.Did, indexAsDate, postIndexedAt)
			// 				postIndexedAt = minIndexedAt
			// 			}
			// 		}
			// 	}
			// }
			var posters_madness int64 = 0;
			if interest.PostersMadnessSymptomatic > 0 && (replyParentAuthor == "" || interest.PostersMadnessReplyToSymptomatic > 0) {
				posters_madness = 1;
			}
			err := updates.SavePost(ctx, writeSchema.SavePostParams{
				Uri:               postUri,
				Author:            event.Did,
				ReplyParent:       database.ToNullString(replyParent),
				ReplyParentAuthor: database.ToNullString(replyParentAuthor),
				ReplyRoot:         database.ToNullString(replyRoot),
				ReplyRootAuthor:   database.ToNullString(replyRootAuthor),
				IndexedAt:         postIndexedAt.Format(time.RFC3339),
				CreatedAt:         database.ToNullString(rawCreatedAt),
				DirectReplyCount:  0,
				InteractionCount:  0,
				LikeCount:         0,
				ReplyCount:        0,
				PostersMadness:    sql.NullInt64{Int64: posters_madness, Valid: true},
				ExternalUri:       database.ToNullString(externalUri),
				QuotedPostUri:     database.ToNullString(quotedPostUri),
			})
			if err != nil {
				return err
			}
			// if replyParent != "" {
			// 	for _, followedBy := range authorFollowedBy {
			// 		err := updates.SavePostDirectRepliedToByFollowing(ctx, writeSchema.SavePostDirectRepliedToByFollowingParams{
			// 			User:      followedBy,
			// 			Uri:       replyParent,
			// 			Author:    replyParentAuthor,
			// 			IndexedAt: indexedAt,
			// 		})
			// 		if err != nil {
			// 			return err
			// 		}
			// 	}
			// }
			// if replyRoot != replyParent {
			// 	for _, followedBy := range authorFollowedBy {
			// 		err := updates.SavePostRepliedToByFollowing(ctx, writeSchema.SavePostRepliedToByFollowingParams{
			// 			User:      followedBy,
			// 			Uri:       replyRoot,
			// 			Author:    replyRootAuthor,
			// 			IndexedAt: indexedAt,
			// 		})
			// 		if err != nil {
			// 			return err
			// 		}
			// 	}
			// }
		}
		if interest.ReplyToAuthor > 0 {
			err := updates.IncrementPostDirectReply(ctx, postUri)
			if err != nil {
				return err
			}

			if interest.PostByUser > 0 && event.Did != replyParentAuthor {
				err := updates.SaveUserInteraction(ctx, writeSchema.SaveUserInteractionParams{
					InteractionUri: postUri,
					AuthorDid:      replyParentAuthor,
					UserDid:        event.Did,
					PostUri:        replyParent,
					Type:           "reply",
					IndexedAt:      indexedAt,
				})
				if err != nil {
					return err
				}
			}
		}
		if interest.ReplyToUser > 0 && event.Did != replyParentAuthor {
			// Someone replying a post by one of the users
			err := updates.SaveInteractionWithUser(ctx, writeSchema.SaveInteractionWithUserParams{
				InteractionUri:       postUri,
				InteractionAuthorDid: event.Did,
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
			if interest.ReplyToThreadAuthor > 0 {
				err := updates.IncrementPostIndirectReply(ctx, postUri)
				if err != nil {
					return err
				}

				if interest.PostByUser > 0 && event.Did != replyRootAuthor {
					err := updates.SaveUserInteraction(ctx, writeSchema.SaveUserInteractionParams{
						InteractionUri: postUri,
						AuthorDid:      replyRootAuthor,
						UserDid:        event.Did,
						PostUri:        replyRoot,
						Type:           "threadReply",
						IndexedAt:      indexedAt,
					})
					if err != nil {
						return err
					}
				}
			}
			if interest.ReplyToThreadUser > 0 && event.Did != replyRootAuthor {
				// Someone replying a post by one of the users
				err := updates.SaveInteractionWithUser(ctx, writeSchema.SaveInteractionWithUserParams{
					InteractionUri:       postUri,
					InteractionAuthorDid: event.Did,
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
