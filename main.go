package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"

	"github.com/bluesky-social/indigo/repomgr"
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
			following.SyncFollowing(syncFollowingParams)
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
		Algo: sql.NullString{String: "1234", Valid: true},
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
	firehoseListeners["app.bsky.graph.follow"] = func(event subscription.FirehoseEvent) {
		switch event.EventKind {
		case repomgr.EvtKindCreateRecord:
			if allFollowing.UserDids[event.Author] {
				fmt.Println(allFollowing.FollowedByCount[event.Record["subject"].(string)])
				fmt.Println(allFollowing.FollowingRecords[event.Uri])
				allFollowing.RecordFollow(event.Uri, event.Author, event.Record["subject"].(string))
				fmt.Printf("Record Following %s %s => %s\n", event.Uri, event.Author, event.Record["subject"].(string))
				fmt.Println(allFollowing.FollowedByCount[event.Record["subject"].(string)])
				fmt.Println(allFollowing.FollowingRecords[event.Uri])
			}
		case repomgr.EvtKindDeleteRecord:
			if allFollowing.UserDids[event.Author] {
				fmt.Println(allFollowing.FollowedByCount["did:plc:k626emd4xi4h3wxpd44s4wpk"])
				fmt.Println(allFollowing.FollowingRecords[event.Uri])
				allFollowing.RemoveFollow(event.Uri)
				fmt.Printf("Remove Following %s %s\n", event.Uri, event.Author)
				fmt.Println(allFollowing.FollowedByCount["did:plc:k626emd4xi4h3wxpd44s4wpk"])
				fmt.Println(allFollowing.FollowingRecords[event.Uri])
			}
		}
	}
	postProcessor := processor.PostProcessor{
		Ctx:      ctx,
		Database: database,
	}
	firehoseListeners["app.bsky.feed.post"] = postProcessor.Process
	// firehoseListeners["app.bsky.feed.like"] = func(event firehoseEvent) {
	// 	// fmt.Println(event)
	// }
	fmt.Println("Starting")
	err = subscription.Subscribe(ctx, os.Getenv("FEEDGEN_SUBSCRIPTION_ENDPOINT"), database, firehoseListeners)
	if err != nil {
		panic(fmt.Sprintf("subscribing to firehose failed (dialing): %s", err))
	}
}
