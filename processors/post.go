package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/bluesky-social/jetstream/pkg/models"
	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/database/read"
	writeSchema "github.com/orthanc/feedgenerator/database/write"
)

type ReferencedPostType = int64

const (
	RootPost ReferencedPostType = iota
	QuotedPost
	RepostedPost
)

type ReferencedPost struct {
	PostUri           string
	SourcePostUri     string
	SourceEventAuthor string
	SourceIndexedAt   string
	ReferenceType     ReferencedPostType
}

type PostProcessor struct {
	Database     *database.Database
	PublicClient *xrpc.Client
	PostUrisChan chan ReferencedPost
}

var commandRegexp = regexp.MustCompile(`!(catchup)\s+(rewind|forward|new)`)

func NewPostProcessor(Database *database.Database,
	PublicClient *xrpc.Client) *PostProcessor {
	PostUrisChan := make(chan ReferencedPost, 100)
	processor := PostProcessor{
		Database:     Database,
		PublicClient: PublicClient,
		PostUrisChan: PostUrisChan,
	}

	go processor.batchEnsurePostsSaved(context.Background())
	return &processor
}

func safeIndexedAt(rawCreatedAt string, authorDid string) (bool, time.Time) {
	now := time.Now().UTC().Add(time.Minute)
	indexedAtDate := now
	createdAt, err := time.Parse(time.RFC3339, rawCreatedAt)
	if err != nil {
		fmt.Printf("Ignoring unparsable create date %s on post by %s, using %s instead\n", rawCreatedAt, authorDid, now)
		indexedAtDate = now
	}
	indexedAtDate = createdAt.UTC()
	if now.Before(indexedAtDate) {
		fmt.Printf("Ignoring future create date %s on post by %s, using %s instead\n", indexedAtDate, authorDid, now)
		indexedAtDate = now
	}
	if indexedAtDate.Before(now.Add(-7 * 24 * time.Hour)) {
		fmt.Printf("Dropping post by %s with create date %s, more than 7 days ago\n", authorDid, indexedAtDate)
		return true, indexedAtDate
	}
	return false, indexedAtDate
}

func (processor *PostProcessor) ensurePostsSaved(ctx context.Context, referencedPosts []ReferencedPost) ([]ReferencedPost, error) {
	var additionalPostsToResolve []ReferencedPost
	postUris := make([]string, 0, len(referencedPosts))
	referencedPostsByUri := make(map[string]ReferencedPost)
	for _, referencedPost := range referencedPosts {
		postUris = append(postUris, referencedPost.PostUri)
		referencedPostsByUri[referencedPost.PostUri] = referencedPost
	}
	existingPosts, err := processor.Database.Queries.GetPostsByUri(ctx, postUris)
	if err != nil {
		return additionalPostsToResolve, err
	}
	var urisToFetch []string
	for _, postUri := range postUris {
		index := slices.IndexFunc(existingPosts, func(post read.GetPostsByUriRow) bool {
			return post.Uri == postUri
		})
		if index == -1 {
			urisToFetch = append(urisToFetch, postUri)
		}
	}
	var fetchedPosts *bsky.FeedGetPosts_Output
	if len(urisToFetch) > 0 {
		fetchedPosts, err = bsky.FeedGetPosts(ctx, processor.PublicClient, urisToFetch)
		if err != nil {
			return additionalPostsToResolve, err
		}
	}
	updates, tx, err := processor.Database.BeginTx(ctx)
	if err != nil {
		return additionalPostsToResolve, err
	}
	defer tx.Rollback()
	if fetchedPosts != nil {
		for _, post := range fetchedPosts.Posts {
			postRecord := post.Record.Val.(*bsky.FeedPost)
			var replyParent, replyParentAuthor, replyRoot, replyRootAuthor, externalUri, quotedPostUri string
			if postRecord.Reply != nil {
				if postRecord.Reply.Parent != nil {
					replyParent = postRecord.Reply.Parent.Uri
					replyParentAuthor = getAuthorFromPostUri(replyParent)
				}
				if postRecord.Reply.Root != nil {
					replyRoot = postRecord.Reply.Root.Uri
					replyRootAuthor = getAuthorFromPostUri(replyRoot)
				}
			}

			if post.Embed != nil {
				if post.Embed.EmbedExternal_View != nil && post.Embed.EmbedExternal_View.External != nil {
					externalUri = post.Embed.EmbedExternal_View.External.Uri
				}
				if post.Embed.EmbedRecord_View != nil && post.Embed.EmbedRecord_View.Record != nil && post.Embed.EmbedRecord_View.Record.EmbedRecord_ViewRecord != nil && getAuthorFromPostUri(post.Embed.EmbedRecord_View.Record.EmbedRecord_ViewRecord.Uri) != "" {
					quotedPostUri = post.Embed.EmbedRecord_View.Record.EmbedRecord_ViewRecord.Uri
				}
				if post.Embed.EmbedRecordWithMedia_View != nil && post.Embed.EmbedRecordWithMedia_View.Record != nil && post.Embed.EmbedRecordWithMedia_View.Record.Record != nil && post.Embed.EmbedRecordWithMedia_View.Record.Record.EmbedRecord_ViewRecord != nil && getAuthorFromPostUri(post.Embed.EmbedRecordWithMedia_View.Record.Record.EmbedRecord_ViewRecord.Uri) != "" {
					quotedPostUri = post.Embed.EmbedRecordWithMedia_View.Record.Record.EmbedRecord_ViewRecord.Uri
				}
			}

			referenceType := referencedPostsByUri[post.Uri].ReferenceType
			if quotedPostUri != "" && (referenceType == RepostedPost || referenceType == RootPost) {
				additionalPostsToResolve = append(additionalPostsToResolve, ReferencedPost{
					PostUri:           quotedPostUri,
					SourceEventAuthor: referencedPostsByUri[post.Uri].SourceEventAuthor,
					SourceIndexedAt:   referencedPostsByUri[post.Uri].SourceIndexedAt,
					SourcePostUri:     referencedPostsByUri[post.Uri].SourcePostUri,
					ReferenceType:     QuotedPost,
				})
			}

			rawCreatedAt := postRecord.CreatedAt
			skip, indexedAtDate := safeIndexedAt(rawCreatedAt, post.Author.Did)
			if skip {
				continue
			}
			indexedAt := indexedAtDate.Format(time.RFC3339)
			err := updates.SavePost(ctx, writeSchema.SavePostParams{
				Uri:               post.Uri,
				Author:            post.Author.Did,
				ReplyParent:       database.ToNullString(replyParent),
				ReplyParentAuthor: database.ToNullString(replyParentAuthor),
				ReplyRoot:         database.ToNullString(replyRoot),
				ReplyRootAuthor:   database.ToNullString(replyRootAuthor),
				IndexedAt:         indexedAt,
				CreatedAt:         database.ToNullString(rawCreatedAt),
				DirectReplyCount:  0,
				InteractionCount:  0,
				LikeCount:         0,
				ReplyCount:        0,
				ExternalUri:       database.ToNullString(externalUri),
				QuotedPostUri:     database.ToNullString(quotedPostUri),
			})
			if err != nil {
				return additionalPostsToResolve, err
			}

			if externalUri != "" {
				err := updates.SaveUserLink(ctx, writeSchema.SaveUserLinkParams{
					LinkUri:    externalUri,
					SeenAt:     referencedPostsByUri[post.Uri].SourceIndexedAt,
					PostUri:    referencedPostsByUri[post.Uri].SourcePostUri,
					PostAuthor: referencedPostsByUri[post.Uri].SourceEventAuthor,
				})
				if err != nil {
					return additionalPostsToResolve, err
				}
			}
		}
	}
	for _, post := range existingPosts {
		if post.ExternalUri.Valid {
			err := updates.SaveUserLink(ctx, writeSchema.SaveUserLinkParams{
				LinkUri:    post.ExternalUri.String,
				SeenAt:     referencedPostsByUri[post.Uri].SourceIndexedAt,
				PostUri:    referencedPostsByUri[post.Uri].SourcePostUri,
				PostAuthor: referencedPostsByUri[post.Uri].SourceEventAuthor,
			})
			if err != nil {
				return additionalPostsToResolve, err
			}
		}

		referenceType := referencedPostsByUri[post.Uri].ReferenceType
		if post.QuotedPostUri.Valid && getAuthorFromPostUri(post.QuotedPostUri.String) != "" && (referenceType == RepostedPost || referenceType == RootPost) {
			additionalPostsToResolve = append(additionalPostsToResolve, ReferencedPost{
				PostUri:           post.QuotedPostUri.String,
				SourceEventAuthor: referencedPostsByUri[post.Uri].SourceEventAuthor,
				SourceIndexedAt:   referencedPostsByUri[post.Uri].SourceIndexedAt,
				SourcePostUri:     referencedPostsByUri[post.Uri].SourcePostUri,
				ReferenceType:     QuotedPost,
			})
		}
	}
	tx.Commit()

	return additionalPostsToResolve, nil
}

func (processor *PostProcessor) batchEnsurePostsSaved(ctx context.Context) {
	var batch []ReferencedPost
	for {
		if len(batch) == 0 {
			batch = append(batch, <-processor.PostUrisChan)
		}
	LOOP:
		for len(batch) < 25 {
			select {
			case referencedPost := <-processor.PostUrisChan:
				batch = append(batch, referencedPost)
			default:
				break LOOP
			}
		}
		additionalPostsToResolve, err := processor.ensurePostsSaved(ctx, batch)
		if err != nil {
			fmt.Printf("Error saving post batch %e\n", err)
		}
		batch = append(batch[:0], additionalPostsToResolve...)
	}
}

func (processor *PostProcessor) processCommandPosts(ctx context.Context, event *models.Event, post bsky.FeedPost) {
	commandMatches := commandRegexp.FindStringSubmatch(strings.ToLower(post.Text))
	if commandMatches != nil {
		if commandMatches[1] == "catchup" && commandMatches[2] == "rewind" {
			lastSessions, err := processor.Database.Queries.GetLastSession(ctx, read.GetLastSessionParams{
				UserDid: event.Did,
				Algo:    database.ToNullString("catchup"),
			})
			if err != nil {
				fmt.Printf("CATCHUP REWIND ERROR: unable to load last session %e\n", err)
				return
			}
			if len(lastSessions) == 0 {
				return
			}
			lastSession := lastSessions[0]
			currentPostsSince, err := time.Parse(time.RFC3339, lastSession.PostsSince)
			if err != nil {
				fmt.Printf("CATCHUP REWIND ERROR: unable to parse current posts since %e\n", err)
				return
			}

			rewindToSessions, err := processor.Database.Queries.GetRewindToSession(ctx, read.GetRewindToSessionParams{
				UserDid:       event.Did,
				Algo:          database.ToNullString("catchup"),
				StartedBefore: lastSession.PostsSince,
			})
			if err != nil {
				fmt.Printf("CATCHUP REWIND ERROR: unable to load rewind session %e\n", err)
				return
			}
			newPostsSince := currentPostsSince.Add(time.Duration(-24) * time.Hour).Format(time.RFC3339)
			if len(rewindToSessions) > 0 {
				newPostsSince = rewindToSessions[0].StartedAt
			}

			processor.Database.Updates.UpdateSessionPostsSince(ctx, writeSchema.UpdateSessionPostsSinceParams{
				SessionId:  lastSession.SessionId,
				PostsSince: newPostsSince,
			})
		}
		if commandMatches[1] == "catchup" && commandMatches[2] == "forward" {
			lastSessions, err := processor.Database.Queries.GetLastSession(ctx, read.GetLastSessionParams{
				UserDid: event.Did,
				Algo:    database.ToNullString("catchup"),
			})
			if err != nil {
				fmt.Printf("CATCHUP FORWARD ERROR: unable to load last session %e\n", err)
				return
			}
			if len(lastSessions) == 0 {
				return
			}
			lastSession := lastSessions[0]

			forwardToSessions, err := processor.Database.Queries.GetForwardSession(ctx, read.GetForwardSessionParams{
				UserDid:      event.Did,
				Algo:         database.ToNullString("catchup"),
				StartedAfter: lastSession.PostsSince,
			})
			if err != nil {
				fmt.Printf("CATCHUP FORWARD ERROR: unable to load forward session %e\n", err)
				return
			}
			if len(forwardToSessions) == 0 {
				return
			}
			forwardToSession := forwardToSessions[0]

			processor.Database.Updates.UpdateSessionPostsSince(ctx, writeSchema.UpdateSessionPostsSinceParams{
				SessionId:  lastSession.SessionId,
				PostsSince: forwardToSession.StartedAt,
			})
		}
		if commandMatches[1] == "catchup" && commandMatches[2] == "new" {
			lastSessions, err := processor.Database.Queries.GetLastSession(ctx, read.GetLastSessionParams{
				UserDid: event.Did,
				Algo:    database.ToNullString("catchup"),
			})
			if err != nil {
				fmt.Printf("CATCHUP NEW ERROR: unable to load last session %e\n", err)
				return
			}
			if len(lastSessions) == 0 {
				return
			}
			lastSession := lastSessions[0]

			lastSeen := time.Now().UTC().Format(time.RFC3339)
			err = processor.Database.Updates.SaveSession(ctx, writeSchema.SaveSessionParams{
				UserDid:     lastSession.UserDid,
				StartedAt:   lastSeen,
				LastSeen:    lastSeen,
				PostsSince:  lastSession.StartedAt,
				AccessCount: sql.NullFloat64{Float64: 0, Valid: true},
				Algo:        lastSession.Algo,
			})
			if err != nil {
				fmt.Printf("CATCHUP NEW ERROR: unable to save new session %e\n", err)
				return
			}
		}
	}
}

func (processor *PostProcessor) Process(ctx context.Context, event *models.Event, postUri string) error {
	switch event.Commit.Operation {
	case models.CommitOperationCreate:
		var post bsky.FeedPost
		if err := json.Unmarshal(event.Commit.Record, &post); err != nil {
			fmt.Printf("failed to unmarshal post: %s : at://%s/%s/%s\n", err, event.Did, event.Commit.Collection, event.Commit.RKey)
			return nil
		}

		processor.processCommandPosts(ctx, event, post)

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
				externalUri = post.Embed.EmbedExternal.External.Uri
			}
			if post.Embed.EmbedRecord != nil && post.Embed.EmbedRecord.Record != nil && getAuthorFromPostUri(post.Embed.EmbedRecord.Record.Uri) != "" {
				quotedPostUri = post.Embed.EmbedRecord.Record.Uri
			}
			if post.Embed.EmbedRecordWithMedia != nil && post.Embed.EmbedRecordWithMedia.Record != nil && post.Embed.EmbedRecordWithMedia.Record.Record != nil && getAuthorFromPostUri(post.Embed.EmbedRecordWithMedia.Record.Record.Uri) != "" {
				quotedPostUri = post.Embed.EmbedRecordWithMedia.Record.Record.Uri
			}
		}

		interest, err := processor.Database.Queries.GetPostFollowData(ctx, read.GetPostFollowDataParams{
			PostAuthor:        event.Did,
			ReplyParentAuthor: replyParentAuthor,
			ReplyRootAuthor:   replyRootAuthor,
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
			interest.ReplyToThreadUser > 0) {
			return nil
		}

		rawCreatedAt := post.CreatedAt
		skip, indexedAtDate := safeIndexedAt(rawCreatedAt, event.Did)
		if skip {
			return nil
		}
		indexedAt := indexedAtDate.Format(time.RFC3339)

		if replyRoot != "" {
			processor.PostUrisChan <- ReferencedPost{
				PostUri:           replyRoot,
				SourceEventAuthor: event.Did,
				SourceIndexedAt:   indexedAt,
				SourcePostUri:     postUri,
				ReferenceType:     RootPost,
			}
		}
		if quotedPostUri != "" {
			processor.PostUrisChan <- ReferencedPost{
				PostUri:           quotedPostUri,
				SourceEventAuthor: event.Did,
				SourceIndexedAt:   indexedAt,
				SourcePostUri:     postUri,
				ReferenceType:     QuotedPost,
			}
		}

		updates, tx, err := processor.Database.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if interest.PostByAuthor > 0 {
			// if event.Did == replyParentAuthor && event.Did == replyRootAuthor {
			// 	parentPostDates, _ := processor.Database.Queries.GetPostDates(ctx, replyParent)
			// 	if parentPostDates.IndexedAt != "" {
			// 		parentIndexedAt, err := time.Parse(time.RFC3339, parentPostDates.IndexedAt)
			// 		if err == nil {
			// 			minIndexedAt := parentIndexedAt.Add(30 * time.Second)
			// 			if postIndexedAt.Before(minIndexedAt) {
			// 				fmt.Printf("Delaying thread post by %s from %s to %s\n", event.Did, indexedAtDate, postIndexedAt)
			// 				postIndexedAt = minIndexedAt
			// 			}
			// 		}
			// 	}
			// }
			err := updates.SavePost(ctx, writeSchema.SavePostParams{
				Uri:               postUri,
				Author:            event.Did,
				ReplyParent:       database.ToNullString(replyParent),
				ReplyParentAuthor: database.ToNullString(replyParentAuthor),
				ReplyRoot:         database.ToNullString(replyRoot),
				ReplyRootAuthor:   database.ToNullString(replyRootAuthor),
				IndexedAt:         indexedAt,
				CreatedAt:         database.ToNullString(rawCreatedAt),
				DirectReplyCount:  0,
				InteractionCount:  0,
				LikeCount:         0,
				ReplyCount:        0,
				ExternalUri:       database.ToNullString(externalUri),
				QuotedPostUri:     database.ToNullString(quotedPostUri),
			})
			if err != nil {
				return err
			}

			if externalUri != "" {
				err := updates.SaveUserLink(ctx, writeSchema.SaveUserLinkParams{
					LinkUri:    externalUri,
					SeenAt:     indexedAt,
					PostUri:    postUri,
					PostAuthor: event.Did,
				})
				if err != nil {
					return err
				}
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
