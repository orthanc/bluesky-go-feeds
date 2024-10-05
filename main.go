package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"

	"github.com/bluesky-social/indigo/repomgr"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/following"
	processor "github.com/orthanc/feedgenerator/processors"
	"github.com/orthanc/feedgenerator/subscription"
	"github.com/orthanc/feedgenerator/web"
)

func main() {
	ctx := context.Background()

	database, err := database.NewDatabase(ctx)
	if err != nil {
		panic(err)
	}

	client := xrpc.Client{
		Host: "https://bsky.social",
	}
	allFollowing := following.New(
		ctx,
		database,
		&client,
	)
	allFollowing.Hydrate()

	go web.StartServer()

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
	subscription.Subscribe(ctx, fmt.Sprintf("%s/xrpc/com.atproto.sync.subscribeRepos", os.Getenv("FEEDGEN_SUBSCRIPTION_ENDPOINT")), firehoseListeners)
}
