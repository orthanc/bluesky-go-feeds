package main

import (
	"context"
	"flag"
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
	dbDown := flag.Bool("db-down", false, "migrate the database down one revision then exit");
	dbUp := flag.Bool("db-up", false, "migrate the database up to the latest revision then exit");
	flag.Parse()

	database, err := database.NewDatabase(ctx)
	if err != nil {
		panic(err)
	}
	if *dbDown {
		err := database.MigrateDown(ctx)
		if err != nil {
			panic(err)
		}
		return
	}
	err = database.MigrateUp(ctx)
	if err != nil {
		panic(err)
	}
	if *dbUp {
		return
	}

	batchMutex := &sync.Mutex{}
	client := xrpc.Client{
		Host: "https://bsky.social",
	}
	publicClient := xrpc.Client{
		Host: "https://public.api.bsky.app",
	}
	allFollowing := following.NewAllFollowing(
		database,
		&client,
		&publicClient,
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
