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
	ctx context.Context
	database *database.Database
	client *xrpc.Client

	UserDids map[string]bool
	FollowingRecords map[string]schema.Following
	FollowedByCount map[string]int
}

func New(ctx context.Context, database *database.Database, client *xrpc.Client) AllFollowing {
	return AllFollowing{
		ctx: ctx,
		database: database,
		client: client,
		UserDids: make(map[string]bool),
		FollowingRecords: make(map[string]schema.Following),
		FollowedByCount: make(map[string]int),
	}
}

func  (allFollowing *AllFollowing) addFollowData(record schema.Following) {
	if _, ok := allFollowing.FollowingRecords[record.Uri]; ok {
		return
	}
	allFollowing.FollowingRecords[record.Uri] = record
	allFollowing.FollowedByCount[record.Following]++
}

func  (allFollowing *AllFollowing) removeFollowData(uri string) (schema.Following, bool) {
	record, ok := allFollowing.FollowingRecords[uri];
	if ok {
		allFollowing.FollowedByCount[record.Following]--
		if allFollowing.FollowedByCount[record.Following] <= 0 {
			delete(allFollowing.FollowedByCount, record.Following)
		}
		delete(allFollowing.FollowingRecords, uri)
	}
	return record, ok
}

func (allFollowing *AllFollowing) Hydrate() {
	userDids, err := allFollowing.database.Queries.ListAllUsers(allFollowing.ctx);
	if err != nil {
		panic(err)
	}
	for _, userDid := range userDids {
		allFollowing.UserDids[userDid] = true
	}

	followingRecords, err := allFollowing.database.Queries.ListAllFollowing(allFollowing.ctx)
	if err != nil {
		panic(err)
	}
	for _, followingRecord :=  range followingRecords {
		allFollowing.addFollowData(followingRecord)
	}
}

func (allFollowing *AllFollowing) saveFollowingPage(records []schema.Following) {
	updates, tx, err := allFollowing.database.BeginTx(allFollowing.ctx)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()
	
	for _, record := range records {
		if _, ok := allFollowing.FollowingRecords[record.Uri]; ok {
			continue
		}
		if err := updates.SaveFollowing(allFollowing.ctx, writeSchema.SaveFollowingParams(record)); err != nil {
			panic(err)
		}

		author := writeSchema.SaveAuthorParams{
			Did: record.Following,
			MedianDirectReplyCount: 0,
			MedianInteractionCount: 0,
			MedianLikeCount: 0,
			MedianReplyCount: 0,
		}
		if err := updates.SaveAuthor(allFollowing.ctx, author); err != nil {
			panic(err)
		}

		allFollowing.addFollowData(record)
	}

	tx.Commit()
}

func (allFollowing *AllFollowing) SyncFollowers(userDid string, lastSeen string) {
	user := writeSchema.SaveUserParams{
		UserDid: userDid,
		LastSeen: lastSeen,
	}
	if err := allFollowing.database.Updates.SaveUser(allFollowing.ctx, user); err != nil {
		panic(err)
	}
	allFollowing.UserDids[userDid] = true

	follows := make([]schema.Following, 0, 100)
	for cursor := "";; {
		followResult, err := atproto.RepoListRecords(allFollowing.ctx, allFollowing.client, "app.bsky.graph.follow", cursor, 100, userDid, false, "", "")
		if err != nil {
			panic(err)
		}

		follows = follows[:len(followResult.Records)]
		for i, record := range followResult.Records {
			follows[i] = schema.Following{
				Uri: record.Uri,
				Following: record.Value.Val.(*bsky.GraphFollow).Subject,
				FollowedBy: userDid,
				UserInteractionRatio: sql.NullFloat64{Float64: 0.1, Valid: true},
			}
		}
		allFollowing.saveFollowingPage(follows)
		fmt.Println("Saved Page")
		if (followResult.Cursor == nil) {
			break;
		}
		cursor = *followResult.Cursor
	}
}

func (allFollowing *AllFollowing) RecordFollow(uri string, followedBy string, following string) {
	record := schema.Following{
		Uri: uri,
		Following: following,
		FollowedBy: followedBy,
		UserInteractionRatio: sql.NullFloat64{Float64: 0.1, Valid: true},
	}

	allFollowing.saveFollowingPage([]schema.Following{record})
}

func (allFollowing *AllFollowing) RemoveFollow(uri string) {
	_, ok := allFollowing.removeFollowData(uri)
	if ok {
		err := allFollowing.database.Updates.DeleteFollowing(allFollowing.ctx, uri)
		if err != nil {
			panic(err)
		}
	}
}