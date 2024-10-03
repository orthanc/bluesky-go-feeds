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
	"github.com/orthanc/feedgenerator/feeddb"
)

type AllFollowing struct {
	ctx context.Context
	database *database.Database
	client *xrpc.Client

	FollowingRecords map[string]feeddb.Following
	FollowedByCount map[string]int
}

func New(ctx context.Context, database *database.Database, client *xrpc.Client) AllFollowing {
	return AllFollowing{
		ctx: ctx,
		database: database,
		client: client,
		FollowingRecords: make(map[string]feeddb.Following),
		FollowedByCount: make(map[string]int),
	}
}

func  (allFollowing *AllFollowing) addFollowData(record feeddb.Following) {
	if _, ok := allFollowing.FollowingRecords[record.Uri]; ok {
		return
	}
	allFollowing.FollowingRecords[record.Uri] = record
	allFollowing.FollowedByCount[record.Following]++
}

func (allFollowing *AllFollowing) saveFollowingPage(records []feeddb.Following) {
	tx, err := allFollowing.database.DB.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()
	
	qtx := allFollowing.database.Queries.WithTx(tx)
	for _, record := range records {
		if _, ok := allFollowing.FollowingRecords[record.Uri]; ok {
			continue
		}
		if err := qtx.SaveFollowing(allFollowing.ctx, feeddb.SaveFollowingParams(record)); err != nil {
			panic(err)
		}

		author := feeddb.SaveAuthorParams{
			Did: record.Following,
			MedianDirectReplyCount: 0,
			MedianInteractionCount: 0,
			MedianLikeCount: 0,
			MedianReplyCount: 0,
		}
		if err := qtx.SaveAuthor(allFollowing.ctx, author); err != nil {
			panic(err)
		}

		allFollowing.addFollowData(record)
	}

	tx.Commit()
}

func (allFollowing *AllFollowing) SyncFollowers(userDid string, lastSeen string) {
	user := feeddb.SaveUserParams{
		UserDid: userDid,
		LastSeen: lastSeen,
	}
	if err := allFollowing.database.Queries.SaveUser(allFollowing.ctx, user); err != nil {
		panic(err)
	}

	follows := make([]feeddb.Following, 0, 100)
	for cursor := "";; {
		followResult, err := atproto.RepoListRecords(allFollowing.ctx, allFollowing.client, "app.bsky.graph.follow", cursor, 100, userDid, false, "", "")
		if err != nil {
			panic(err)
		}

		follows = follows[:len(followResult.Records)]
		for i, record := range followResult.Records {
			follows[i] = feeddb.Following{
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