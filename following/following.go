package following

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"

	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
	writeSchema "github.com/orthanc/feedgenerator/database/write"
)

type AllFollowing struct {
	database     *database.Database
	client       *xrpc.Client
	publicClient *xrpc.Client
}

var emptyFollowedBy []string

const purgePageSize = 10000
var cutoverTime = time.Date(2024, 11, 13, 8, 45, 0, 0, time.UTC)

func NewAllFollowing(database *database.Database, client *xrpc.Client, publicClient *xrpc.Client) *AllFollowing {
	allFollowing := &AllFollowing{
		database:     database,
		client:       client,
		publicClient: publicClient,
	}

	ctx := context.Background()
	ticker := time.NewTicker(time.Hour)
	go func() {
		for range ticker.C {
			err := allFollowing.Purge(ctx)
			if err != nil {
				fmt.Printf("Error purging data: %s\n", err)
			}
		}
	}()

	return allFollowing
}

func (allFollowing *AllFollowing) saveFollowingPage(ctx context.Context, records []schema.Following) error {
	updates, tx, err := allFollowing.database.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range records {
		author := writeSchema.SaveAuthorParams{
			Did:                    record.Following,
			MedianDirectReplyCount: 0,
			MedianInteractionCount: 0,
			MedianLikeCount:        0,
			MedianReplyCount:       0,
		}
		if err := updates.SaveAuthor(ctx, author); err != nil {
			return err
		}

		if err := updates.SaveFollowing(ctx, writeSchema.SaveFollowingParams{
			Uri:                  record.Uri,
			FollowedBy:           record.FollowedBy,
			Following:            record.Following,
			UserInteractionRatio: record.UserInteractionRatio,
			LastRecorded:         record.LastRecorded,
		}); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (allFollowing *AllFollowing) removeFollowingNotRecordedAfter(ctx context.Context, userDid string, before string) error {
	followingToRemove, err := allFollowing.database.Queries.ListFollowingLastRecordedBefore(ctx, schema.ListFollowingLastRecordedBeforeParams{
		FollowedBy:   userDid,
		LastRecorded: database.ToNullString(before),
	})
	if err != nil {
		return err
	}

	updates, tx, err := allFollowing.database.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, following := range followingToRemove {
		err := updates.DeleteFollowing(ctx, following.Uri)
		if err != nil {
			return err
		}
	}
	tx.Commit()
	return nil
}

func (allFollowing *AllFollowing) SyncFollowing(ctx context.Context, userDid string, lastSeen string) error {
	user := writeSchema.SaveUserParams{
		UserDid:    userDid,
		LastSeen:   lastSeen,
		LastSynced: database.ToNullString(time.Now().UTC().Format(time.RFC3339)),
	}
	if err := allFollowing.database.Updates.SaveUser(ctx, user); err != nil {
		return err
	}

	syncStart := time.Now().UTC().Format(time.RFC3339)
	follows := make([]schema.Following, 0, 100)
	for cursor := ""; ; {
		lastRecorded := time.Now().UTC().Format(time.RFC3339)
		followResult, err := atproto.RepoListRecords(ctx, allFollowing.client, "app.bsky.graph.follow", cursor, 100, user.UserDid, false, "", "")
		if err != nil {
			return err
		}

		follows = follows[:len(followResult.Records)]
		for i, record := range followResult.Records {
			follows[i] = schema.Following{
				Uri:                  record.Uri,
				Following:            record.Value.Val.(*bsky.GraphFollow).Subject,
				FollowedBy:           user.UserDid,
				UserInteractionRatio: sql.NullFloat64{Float64: 0.1, Valid: true},
				LastRecorded:         database.ToNullString(lastRecorded),
			}
		}
		err = allFollowing.saveFollowingPage(ctx, follows)
		if err != nil {
			return err
		}
		fmt.Printf("Saved following page for %s\n", userDid)
		if followResult.Cursor == nil {
			break
		}
		cursor = *followResult.Cursor
	}
	err := allFollowing.removeFollowingNotRecordedAfter(ctx, user.UserDid, syncStart)
	if err != nil {
		return err
	}

	followers := make([]schema.Follower, 0, 100)
	followersFromSync := make(map[string]bool)

	for cursor := ""; ; {
		lastRecorded := time.Now().UTC().Format(time.RFC3339)
		followersResult, err := bsky.GraphGetFollowers(ctx, allFollowing.publicClient, user.UserDid, cursor, 100)
		if err != nil {
			return err
		}

		followers = followers[:len(followersResult.Followers)]
		for i, record := range followersResult.Followers {
			followers[i] = schema.Follower{
				Following:    user.UserDid,
				FollowedBy:   record.Did,
				LastRecorded: lastRecorded,
			}
			followersFromSync[followers[i].FollowedBy] = true
		}
		err = allFollowing.saveFollowerPage(ctx, followers)
		if err != nil {
			return err
		}
		fmt.Printf("Saved follower page for %s\n", userDid)
		if followersResult.Cursor == nil {
			break
		}
		cursor = *followersResult.Cursor
	}
	err = allFollowing.removeFollowersNotRecordedAfter(ctx, user.UserDid, syncStart)
	if err != nil {
		return err
	}
	return nil
}

func (allFollowing *AllFollowing) RecordFollow(ctx context.Context, uri string, followedBy string, following string) error {
	record := schema.Following{
		Uri:                  uri,
		Following:            following,
		FollowedBy:           followedBy,
		UserInteractionRatio: sql.NullFloat64{Float64: 0.1, Valid: true},
		LastRecorded:         database.ToNullString(time.Now().UTC().Format(time.RFC3339)),
	}

	return allFollowing.saveFollowingPage(ctx, []schema.Following{record})
}

func (allFollowing *AllFollowing) RemoveFollow(ctx context.Context, uri string) error {
	err := allFollowing.database.Updates.DeleteFollowing(ctx, uri)
	if err != nil {
		return err
	}
	return nil
}

func (allFollowing *AllFollowing) RecordFollower(ctx context.Context, uri string, following string, followedBy string) error {
	record := schema.Follower{
		Following:    following,
		FollowedBy:   followedBy,
		LastRecorded: time.Now().UTC().Format(time.RFC3339),
	}

	return allFollowing.saveFollowerPage(ctx, []schema.Follower{record})
}

func (allFollowing *AllFollowing) saveFollowerPage(ctx context.Context, records []schema.Follower) error {
	updates, tx, err := allFollowing.database.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range records {
		author := writeSchema.SaveAuthorParams{
			Did:                    record.FollowedBy,
			MedianDirectReplyCount: 0,
			MedianInteractionCount: 0,
			MedianLikeCount:        0,
			MedianReplyCount:       0,
		}
		if err := updates.SaveAuthor(ctx, author); err != nil {
			return err
		}

		if err := updates.SaveFollower(ctx, writeSchema.SaveFollowerParams{
			FollowedBy:   record.FollowedBy,
			Following:    record.Following,
			LastRecorded: record.LastRecorded,
		}); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (allFollowing *AllFollowing) removeFollowersNotRecordedAfter(ctx context.Context, userDid string, before string) error {
	followersToRemove, err := allFollowing.database.Queries.ListFollowerLastRecordedBefore(ctx, schema.ListFollowerLastRecordedBeforeParams{
		Following:    userDid,
		LastRecorded: before,
	})
	if err != nil {
		return err
	}

	updates, tx, err := allFollowing.database.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, follower := range followersToRemove {
		err := updates.DeleteFollower(ctx, writeSchema.DeleteFollowerParams{
			Following:    follower.Following,
			FollowedBy:   follower.FollowedBy,
			LastRecorded: before,
		})
		if err != nil {
			return err
		}
	}
	tx.Commit()
	return nil
}

func (allFollowing *AllFollowing) purgeUser(ctx context.Context, userDid string, purgeBefore string) error {
	if _, err := allFollowing.database.Updates.DeleteUserWhenNotSeen(ctx, writeSchema.DeleteUserWhenNotSeenParams{
		UserDid:     userDid,
		PurgeBefore: purgeBefore,
	}); err != nil {
		return fmt.Errorf("error deleting user %s: %w", userDid, err)
	}
	fmt.Printf("Deleted user %s\n", userDid)
	return nil
}

func pagedPurge(messageTemplate string, deletePage func() (int64, error)) (int64, error) {
	var count int64 = 0
	var pages int64 = 0
	for {
		rows, err := deletePage()
		if err != nil {
			return count, err
		}
		count += rows
		pages++
		if rows < purgePageSize {
			fmt.Printf(messageTemplate+" in %d pages\n", count, pages)
			return count, nil
		}
		if pages%10 == 0 {
			fmt.Printf(messageTemplate+" page %d\n", count, pages)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func (allFollowing *AllFollowing) Purge(ctx context.Context) error {
	now := time.Now().UTC()
	purgeBefore := now.Add(-7 * 24 * time.Hour).Format(time.RFC3339)
	fmt.Printf("purgeBefore %s\n", purgeBefore)
	purgeBefore3Time := now.Add(-3 * 24 * time.Hour)
	fmt.Printf("now %s\n", now.Format(time.RFC3339))
	fmt.Printf("purgeBefore3Time %s\n", purgeBefore3Time.Format(time.RFC3339))
	if now.Before(cutoverTime) {
		timeUntilCutover := cutoverTime.Sub(now);
		fmt.Printf("timeUntilCutover %s\n", timeUntilCutover)
		purgeBefore3Time = purgeBefore3Time.Add(-1 * timeUntilCutover)
		fmt.Printf("purgeBefore3Time %s\n", purgeBefore3Time.Format(time.RFC3339))
	}
	updates := allFollowing.database.Updates
	fmt.Printf("Purging data before %s\n", purgeBefore)

	_, err := pagedPurge("Deleted %d posts", func() (int64, error) {
		return updates.DeletePostsBefore(ctx, writeSchema.DeletePostsBeforeParams{
			IndexedAt: purgeBefore,
			Limit:     purgePageSize,
		})
	})
	if err != nil {
		return fmt.Errorf("error purging posts: %w", err)
	}

	_, err = pagedPurge("Deleted %d reposts", func() (int64, error) {
		return updates.DeleteRepostsBefore(ctx, writeSchema.DeleteRepostsBeforeParams{
			IndexedAt: purgeBefore,
			Limit:     purgePageSize,
		})
	})
	if err != nil {
		return fmt.Errorf("error purging reposts: %w", err)
	}

	_, err = pagedPurge("Deleted %d sessions", func() (int64, error) {
		return updates.DeleteSessionsBefore(ctx, writeSchema.DeleteSessionsBeforeParams{
			LastSeen: purgeBefore,
			Limit:    purgePageSize,
		})
	})
	if err != nil {
		return fmt.Errorf("error purging sessions: %w", err)
	}

	_, err = pagedPurge("Deleted %d user interactions", func() (int64, error) {
		return updates.DeleteUserInteractionsBefore(ctx, writeSchema.DeleteUserInteractionsBeforeParams{
			IndexedAt: purgeBefore,
			Limit:     purgePageSize,
		})
	})
	if err != nil {
		return fmt.Errorf("error purging user interactions: %w", err)
	}

	_, err = pagedPurge("Deleted %d interactions with user", func() (int64, error) {
		return updates.DeleteInteractionWithUsersBefore(ctx, writeSchema.DeleteInteractionWithUsersBeforeParams{
			IndexedAt: purgeBefore,
			Limit:     purgePageSize,
		})
	})
	if err != nil {
		return fmt.Errorf("error purging interactions with user: %w", err)
	}

	_, err = pagedPurge("Deleted %d interactions by followed", func() (int64, error) {
		return updates.DeletePostInteractedByFollowedBefore(ctx, writeSchema.DeletePostInteractedByFollowedBeforeParams{
			IndexedAt: purgeBefore3Time.Format(time.RFC3339),
			Limit:     purgePageSize,
		})
	})
	if err != nil {
		return fmt.Errorf("error purging interactions by followed: %w", err)
	}

	usersToDelete, err := allFollowing.database.Queries.ListUsersNotSeenSince(ctx, purgeBefore)
	if err != nil {
		return fmt.Errorf("error listing users to purge: %w", err)
	}
	for _, userDid := range usersToDelete {
		err := allFollowing.purgeUser(ctx, userDid, purgeBefore)
		if err != nil {
			return fmt.Errorf("error deleting user %s: %w", userDid, err)
		}
	}

	fmt.Println("Purge complete")
	return nil
}
