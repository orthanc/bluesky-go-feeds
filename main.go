package main

import (
	"context"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/orthanc/feedgenerator/feeddb"
	processor "github.com/orthanc/feedgenerator/processors"
	"github.com/orthanc/feedgenerator/subscription"
)

func main() {

	ctx := context.Background()

	queries := createDb()

	firehoseListeners := make(map[string]subscription.FirehoseEventListener)
	// firehoseListeners["app.bsky.graph.follow"] = func(event firehoseEvent) {
	// 	// fmt.Println(event)
	// }
	postProcessor := processor.PostProcessor{
		Ctx:     ctx,
		Queries: *queries,
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
