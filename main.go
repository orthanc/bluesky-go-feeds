package main

import (
	"context"
	"fmt"

	"github.com/orthanc/feedgenerator/feeddb"
)

func main() {
	firehoseListeners := make(map[string]firehoseEventListener)
	firehoseListeners["app.bsky.graph.follow"] = func (event firehoseEvent) {
		fmt.Println(event)
	}
	firehoseListeners["app.bsky.feed.post"] = func (event firehoseEvent) {
		fmt.Println(event)
	}
	firehoseListeners["app.bsky.feed.like"] = func (event firehoseEvent) {
		fmt.Println(event)
	}
	subscribe(context.Background(), "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos", firehoseListeners)

	test := feeddb.Author{
		Did: "",
		MedianLikeCount: 0,
		MedianReplyCount: 0,
		MedianDirectReplyCount: 0,
		MedianInteractionCount: 0,
	}
	fmt.Println((test))
}
