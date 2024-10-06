package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
	"github.com/orthanc/feedgenerator/following"
	processor "github.com/orthanc/feedgenerator/processors"
	"github.com/orthanc/feedgenerator/subscription"
	"github.com/orthanc/feedgenerator/web"
)

func backgroundJobs(following *following.AllFollowing, syncFollowingChan chan following.SyncFollowingParams) {
	for {
		select {
		case syncFollowingParams := <-syncFollowingChan:
			err := following.SyncFollowing(syncFollowingParams)
			if err != nil {
				fmt.Printf("Error syncing follow for %s: %s\n", syncFollowingParams.UserDid, err)
			}
		}
	}
}

func main() {
	ctx := context.Background()

	database, err := database.NewDatabase(ctx)
	if err != nil {
		panic(err)
	}

	lastSession, err := database.Queries.GetLastSession(ctx, schema.GetLastSessionParams{
		UserDid: "aaaa",
		Algo:    sql.NullString{String: "1234", Valid: true},
	})
	fmt.Println(lastSession, err)

	client := xrpc.Client{
		Host: "https://bsky.social",
	}
	allFollowing := following.NewAllFollowing(
		ctx,
		database,
		&client,
	)
	allFollowing.Hydrate()

	syncFollowingChan := make(chan following.SyncFollowingParams)
	go backgroundJobs(allFollowing, syncFollowingChan)

	go web.StartServer(database, syncFollowingChan)

	firehoseListeners := make(map[string]subscription.FirehoseEventListener)
	firehoseListeners["app.bsky.graph.follow"] = (&processor.FollowProcessor{
		Database:     database,
		AllFollowing: allFollowing,
	}).Process
	firehoseListeners["app.bsky.feed.post"] = (&processor.PostProcessor{
		Database:     database,
		AllFollowing: allFollowing,
	}).Process
	firehoseListeners["app.bsky.feed.like"] = (&processor.LikeProcessor{
		Database:     database,
		AllFollowing: allFollowing,
	}).Process
	fmt.Println("Starting")
	err = subscription.Subscribe(ctx, os.Getenv("FEEDGEN_SUBSCRIPTION_ENDPOINT"), database, firehoseListeners)
	if err != nil {
		panic(fmt.Sprintf("subscribing to firehose failed (dialing): %s", err))
	}
}
