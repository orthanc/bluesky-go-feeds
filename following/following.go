package following

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"sync"
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

	userDids         sync.Map
	followingRecords sync.Map
	followedBy       sync.Map
	followers        sync.Map

	batchMutex *sync.Mutex
}

var emptyFollowedBy []string

const purgePageSize = 10000
var cutoverTime = time.Date(2024, 11, 13, 8, 45, 0, 0, time.UTC)

func NewAllFollowing(database *database.Database, client *xrpc.Client, publicClient *xrpc.Client, batchMutex *sync.Mutex) *AllFollowing {
	allFollowing := &AllFollowing{
		database:     database,
		client:       client,
		publicClient: publicClient,
		batchMutex:   batchMutex,
	}

	ctx := context.Background()
	ticker := time.NewTicker(time.Hour)
	go func() {
		err := allFollowing.Purge(ctx)
		if err != nil {
			fmt.Printf("Error purging data: %s\n", err)
		}
		for range ticker.C {
			err := allFollowing.Purge(ctx)
			if err != nil {
				fmt.Printf("Error purging data: %s\n", err)
			}
		}
	}()

	return allFollowing
}

func (allFollowing *AllFollowing) IsUser(userDid string) bool {
	value, present := allFollowing.userDids.Load(userDid)
	return present && value.(bool)
}

func (allFollowing *AllFollowing) IsFollowed(authorDid string) bool {
	value, _ := allFollowing.followedBy.Load(authorDid)
	return value != nil
}

func (allFollowing *AllFollowing) IsFollower(authorDid string) bool {
	value, _ := allFollowing.followers.Load(authorDid)
	return value != nil
}

func (allFollowing *AllFollowing) IsAuthor(authorDid string) bool {
	return allFollowing.IsFollowed(authorDid) || allFollowing.IsFollower(authorDid)
}

func (allFollowing *AllFollowing) FollowedBy(authorDid string) []string {
	value, _ := allFollowing.followedBy.Load(authorDid)
	if value == nil {
		return emptyFollowedBy
	}
	return *value.(*[]string)
}

func (allFollowing *AllFollowing) addFollowData(record schema.Following) {
	_, loaded := allFollowing.followingRecords.Swap(record.Uri, record)
	if loaded {
		return
	}
	for {
		current, _ := allFollowing.followedBy.LoadOrStore(record.Following, &emptyFollowedBy)
		updated := append(slices.Clone(*current.(*[]string)), record.FollowedBy)
		swapped := allFollowing.followedBy.CompareAndSwap(record.Following, current, &updated)
		if swapped {
			break
		}
	}
}

func (allFollowing *AllFollowing) removeFollowData(uri string) (schema.Following, bool) {
	value, loaded := allFollowing.followingRecords.LoadAndDelete(uri)
	if !loaded {
		return schema.Following{}, false
	}
	record := value.(schema.Following)
	for {
		current, _ := allFollowing.followedBy.Load(record.Following)
		if current == nil {
			break
		}
		newVal := slices.Clone(*current.(*[]string))
		newVal = slices.DeleteFunc(newVal, func(e string) bool { return e == record.FollowedBy })
		if len(newVal) > 0 {
			swapped := allFollowing.followedBy.CompareAndSwap(record.Following, current, &newVal)
			if swapped {
				break
			}
		} else {
			deleted := allFollowing.followedBy.CompareAndDelete(record.Following, current)
			if deleted {
				break
			}
		}
	}
	return record, loaded
}

func (allFollowing *AllFollowing) addFollowerData(record schema.Follower) {
	for {
		current, _ := allFollowing.followers.LoadOrStore(record.FollowedBy, &emptyFollowedBy)
		updated := append(slices.Clone(*current.(*[]string)), record.Following)
		swapped := allFollowing.followers.CompareAndSwap(record.FollowedBy, current, &updated)
		if swapped {
			break
		}
	}
}

func (allFollowing *AllFollowing) removeFollowerData(following string, followedBy string) bool {
	for {
		current, _ := allFollowing.followers.Load(followedBy)
		if current == nil || slices.Index(*current.(*[]string), following) == -1 {
			return false
		}
		newVal := slices.Clone(*current.(*[]string))
		newVal = slices.DeleteFunc(newVal, func(e string) bool { return e == following })
		if len(newVal) > 0 {
			swapped := allFollowing.followers.CompareAndSwap(followedBy, current, &newVal)
			if swapped {
				return true
			}
		} else {
			deleted := allFollowing.followers.CompareAndDelete(followedBy, current)
			if deleted {
				return true
			}
		}
	}
}

func (allFollowing *AllFollowing) Hydrate(ctx context.Context) {
	userDids, err := allFollowing.database.Queries.ListAllUsers(ctx)
	if err != nil {
		panic(err)
	}
	for _, userDid := range userDids {
		allFollowing.userDids.Store(userDid, true)
	}

	followingRecords, err := allFollowing.database.Queries.ListAllFollowing(ctx)
	if err != nil {
		panic(err)
	}
	for _, followingRecord := range followingRecords {
		allFollowing.addFollowData(followingRecord)
	}

	followerRecords, err := allFollowing.database.Queries.ListAllFollowers(ctx)
	if err != nil {
		panic(err)
	}
	for _, followerRecord := range followerRecords {
		allFollowing.addFollowerData(followerRecord)
	}
}

func (allFollowing *AllFollowing) saveFollowingPage(ctx context.Context, records []schema.Following) error {
	updates, tx, err := allFollowing.database.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range records {
		if err := updates.SaveFollowing(ctx, writeSchema.SaveFollowingParams{
			Uri:                  record.Uri,
			FollowedBy:           record.FollowedBy,
			Following:            record.Following,
			UserInteractionRatio: record.UserInteractionRatio,
			LastRecorded:         record.LastRecorded,
		}); err != nil {
			return err
		}

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

		allFollowing.addFollowData(record)
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
		allFollowing.removeFollowData(following.Uri)
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
	allFollowing.userDids.Store(user.UserDid, true)

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
	_, ok := allFollowing.removeFollowData(uri)
	if ok {
		err := allFollowing.database.Updates.DeleteFollowing(ctx, uri)
		if err != nil {
			return err
		}
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
		if err := updates.SaveFollower(ctx, writeSchema.SaveFollowerParams{
			FollowedBy:   record.FollowedBy,
			Following:    record.Following,
			LastRecorded: record.LastRecorded,
		}); err != nil {
			return err
		}

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

		allFollowing.addFollowerData(record)
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
		allFollowing.removeFollowerData(follower.Following, follower.FollowedBy)
	}
	tx.Commit()
	return nil
}

func (allFollowing *AllFollowing) purgeUser(ctx context.Context, userDid string, purgeBefore string) error {
	updates, tx, err := allFollowing.database.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if rows, err := updates.DeleteUserWhenNotSeen(ctx, writeSchema.DeleteUserWhenNotSeenParams{
		UserDid:     userDid,
		PurgeBefore: purgeBefore,
	}); err != nil {
		return fmt.Errorf("error deleting user %s: %w", userDid, err)
	} else if rows == 0 {
		// If user wasn't deleted they were seen since we listed them
		// so just return with no action
		return nil
	}

	allFollowing.userDids.Delete(userDid)
	var authorsToDelete []string
	allFollowing.followingRecords.Range(func(key any, value any) bool {
		following := value.(schema.Following)
		if following.FollowedBy == userDid {
			allFollowing.removeFollowData(following.Uri)
			if !allFollowing.IsAuthor(following.Following) {
				authorsToDelete = append(authorsToDelete, following.Following)
			}
		}
		return true
	})
	allFollowing.followers.Range(func(key any, value any) bool {
		followedBy := key.(string)
		removed := allFollowing.removeFollowerData(userDid, followedBy)
		if removed && !allFollowing.IsAuthor(followedBy) {
			authorsToDelete = append(authorsToDelete, followedBy)
		}
		return true
	})
	for start := 0; start < len(authorsToDelete); {
		end := min(start+1000, len(authorsToDelete))
		batch := authorsToDelete[start:end]
		if _, err := updates.DeleteAuthorsByDid(ctx, batch); err != nil {
			return fmt.Errorf("error deleting authors now unused by user %s: %w", userDid, err)
		}
		start = end
	}

	tx.Commit()
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
	allFollowing.batchMutex.Lock()
	defer allFollowing.batchMutex.Unlock()
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
