package following

import (
	"context"
	"database/sql"
	"fmt"

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

	UserDids         map[string]bool
	FollowingRecords map[string]schema.Following
	FollowedByCount  map[string]int
}

type SyncFollowingParams struct {
	UserDid  string
	LastSeen string
}

func NewAllFollowing(database *database.Database, client *xrpc.Client) *AllFollowing {
	return &AllFollowing{
		database:         database,
		client:           client,
		UserDids:         make(map[string]bool),
		FollowingRecords: make(map[string]schema.Following),
		FollowedByCount:  make(map[string]int),
	}
}

func (allFollowing *AllFollowing) addFollowData(record schema.Following) {
	if _, ok := allFollowing.FollowingRecords[record.Uri]; ok {
		return
	}
	allFollowing.FollowingRecords[record.Uri] = record
	allFollowing.FollowedByCount[record.Following]++
}

func (allFollowing *AllFollowing) removeFollowData(uri string) (schema.Following, bool) {
	record, ok := allFollowing.FollowingRecords[uri]
	if ok {
		allFollowing.FollowedByCount[record.Following]--
		if allFollowing.FollowedByCount[record.Following] <= 0 {
			delete(allFollowing.FollowedByCount, record.Following)
		}
		delete(allFollowing.FollowingRecords, uri)
	}
	return record, ok
}

func (allFollowing *AllFollowing) Hydrate(ctx context.Context) {
	userDids, err := allFollowing.database.Queries.ListAllUsers(ctx)
	if err != nil {
		panic(err)
	}
	for _, userDid := range userDids {
		allFollowing.UserDids[userDid] = true
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
		if _, ok := allFollowing.FollowingRecords[record.Uri]; ok {
			continue
		}
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

func (allFollowing *AllFollowing) SyncFollowing(ctx context.Context, params SyncFollowingParams) error {
	user := writeSchema.SaveUserParams(params)
	if err := allFollowing.database.Updates.SaveUser(ctx, user); err != nil {
		return err
	}
	allFollowing.UserDids[user.UserDid] = true

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
