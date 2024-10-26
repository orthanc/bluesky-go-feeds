package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/following"
	processor "github.com/orthanc/feedgenerator/processors"
	"github.com/orthanc/feedgenerator/ratios"
	"github.com/orthanc/feedgenerator/subscription"
	"github.com/orthanc/feedgenerator/web"
)

func main() {
	ctx := context.Background()

	database, err := database.NewDatabase(ctx)
	if err != nil {
		panic(err)
	}

	batchMutex := &sync.Mutex{}
	client := xrpc.Client{
		Host: "https://bsky.social",
	}
	allFollowing := following.NewAllFollowing(
		database,
		&client,
		batchMutex,
	)
	allFollowing.Hydrate(ctx)
	ratios.NewRatios(database, batchMutex)

	go web.StartServer(database, allFollowing)
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
	firehoseListeners["app.bsky.feed.repost"] = (&processor.RepostProcessor{
		Database:     database,
		AllFollowing: allFollowing,
	}).Process
	fmt.Println("Starting")
	err = subscription.Subscribe(ctx, os.Getenv("FEEDGEN_SUBSCRIPTION_ENDPOINT"), database, firehoseListeners)
	if err != nil {
		panic(fmt.Sprintf("subscribing to firehose failed (dialing): %s", err))
	}
}
