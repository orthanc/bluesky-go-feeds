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

type Following struct {
	Ctx context.Context
	Database *database.Database
	Client *xrpc.Client
}

func (following *Following) saveFollowingPage(records []feeddb.Following) {
	tx, err := following.Database.DB.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	qtx := following.Database.Queries.WithTx(tx)
	for _, record := range records {
		if err := qtx.SaveFollowing(following.Ctx, feeddb.SaveFollowingParams(record)); err != nil {
			panic(err)
		}

		author := feeddb.SaveAuthorParams{
			Did: record.Following,
			MedianDirectReplyCount: 0,
			MedianInteractionCount: 0,
			MedianLikeCount: 0,
			MedianReplyCount: 0,
		}
		if err := qtx.SaveAuthor(following.Ctx, author); err != nil {
			panic(err)
		}
		// TODO local state
	}

	tx.Commit()
}

func (following *Following) SyncFollowers(userDid string, lastSeen string) {
	user := feeddb.SaveUserParams{
		UserDid: userDid,
		LastSeen: lastSeen,
	}
	if err := following.Database.Queries.SaveUser(following.Ctx, user); err != nil {
		panic(err)
	}

	follows := make([]feeddb.Following, 0, 100)
	for cursor := "";; {
		followResult, err := atproto.RepoListRecords(following.Ctx, following.Client, "app.bsky.graph.follow", cursor, 100, userDid, false, "", "")
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
		following.saveFollowingPage(follows)
		fmt.Println("Saved Page")
		if (followResult.Cursor == nil) {
			break;
		}
		cursor = *followResult.Cursor
	}
}