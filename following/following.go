package following

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"

	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
	writeSchema "github.com/orthanc/feedgenerator/database/write"
)

type AllFollowing struct {
	database *database.Database
	client   *xrpc.Client

	userDids         sync.Map
	followingRecords sync.Map
	followedByCount  sync.Map
}

func NewAllFollowing(database *database.Database, client *xrpc.Client) *AllFollowing {
	return &AllFollowing{
		database: database,
		client:   client,
	}
}

func (allFollowing *AllFollowing) IsUser(userDid string) bool {
	value, present := allFollowing.userDids.Load(userDid)
	return present && value.(bool)
}

func (allFollowing *AllFollowing) IsFollowed(authorDid string) bool {
	value, present := allFollowing.followedByCount.Load(authorDid)
	return present && value.(int) > 0
}

func (allFollowing *AllFollowing) addFollowData(record schema.Following) {
	_, loaded := allFollowing.followingRecords.Swap(record.Uri, record)
	if loaded {
		return
	}
	for {
		current, _ := allFollowing.followedByCount.LoadOrStore(record.Following, 0)
		swapped := allFollowing.followedByCount.CompareAndSwap(record.Following, current, current.(int)+1)
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
		current, present := allFollowing.followedByCount.Load(record.Following)
		if !present {
			break
		}
		newVal := current.(int) - 1
		if newVal > 0 {
			swapped := allFollowing.followedByCount.CompareAndSwap(record.Following, current, current.(int)-1)
			if swapped {
				break
			}
		} else {
			deleted := allFollowing.followedByCount.CompareAndDelete(record.Following, current)
			if deleted {
				break
			}
		}
	}
	return record, loaded
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
}

func (allFollowing *AllFollowing) saveFollowingPage(ctx context.Context, records []schema.Following) error {
	updates, tx, err := allFollowing.database.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range records {
		if err := updates.SaveFollowing(ctx, writeSchema.SaveFollowingParams(record)); err != nil {
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

func (allFollowing *AllFollowing) SyncFollowing(ctx context.Context, userDid string, lastSeen string) error {
	user := writeSchema.SaveUserParams{
		UserDid: userDid,
		LastSeen: lastSeen,
	}
	if err := allFollowing.database.Updates.SaveUser(ctx, user); err != nil {
		return err
	}
	allFollowing.userDids.Store(user.UserDid, true)

	follows := make([]schema.Following, 0, 100)
	for cursor := ""; ; {
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
			}
		}
		err = allFollowing.saveFollowingPage(ctx, follows)
		if err != nil {
			return err
		}
		fmt.Println("Saved Page")
		if followResult.Cursor == nil {
			break
		}
		cursor = *followResult.Cursor
	}
	return nil
}

func (allFollowing *AllFollowing) RecordFollow(ctx context.Context, uri string, followedBy string, following string) error {
	record := schema.Following{
		Uri:                  uri,
		Following:            following,
		FollowedBy:           followedBy,
		UserInteractionRatio: sql.NullFloat64{Float64: 0.1, Valid: true},
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
