package main

import (
	"context"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/feeddb"
	"github.com/orthanc/feedgenerator/following"
	processor "github.com/orthanc/feedgenerator/processors"
	"github.com/orthanc/feedgenerator/subscription"
)

func main() {

	ctx := context.Background()

	database := database.New()
	database.Migrate()

	client := xrpc.Client{
		Host: "https://bsky.social",
	}
	following := following.Following{
		Ctx: ctx,
		Client: &client,
		Database: database,
	}
	following.SyncFollowers("did:plc:crngjmsdh3zpuhmd5gtgwx6q", time.Now().Format(time.RFC3339))


	firehoseListeners := make(map[string]subscription.FirehoseEventListener)
	// firehoseListeners["app.bsky.graph.follow"] = func(event firehoseEvent) {
	// 	// fmt.Println(event)
	// }
	postProcessor := processor.PostProcessor{
		Ctx:     ctx,
		Database: database,
	}
	firehoseListeners["app.bsky.feed.post"] = postProcessor.Process
	// firehoseListeners["app.bsky.feed.like"] = func(event firehoseEvent) {
	// 	// fmt.Println(event)
	// }
	fmt.Println("Starting")
	subscription.Subscribe(ctx, "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos", firehoseListeners)

	test := feeddb.Author{
		Did:                    "",
		MedianLikeCount:        0,
		MedianReplyCount:       0,
		MedianDirectReplyCount: 0,
		MedianInteractionCount: 0,
	}
	fmt.Println((test))
}
