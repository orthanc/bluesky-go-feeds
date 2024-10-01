package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/orthanc/feedgenerator/feeddb"
)

//go:embed schema.sql
var ddl string

func toNullString(val string) sql.NullString {
	if (val == "") {
		return sql.NullString{Valid: false};
	}
	return sql.NullString{String: val, Valid: true}
}

func getAuthorFromPostUri(postUri string) string {
	parts := strings.SplitN(postUri, "/",4)
	repo := parts[2]
	collection := parts[3]
	if collection == "app.bsky.feed.post" {
		return ""
	}
	return repo
}

func main() {

	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	// create tables
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		panic(err)
	}
	queries := feeddb.New(db)


	firehoseListeners := make(map[string]firehoseEventListener)
	firehoseListeners["app.bsky.graph.follow"] = func (event firehoseEvent) {
		// fmt.Println(event)
	}
	firehoseListeners["app.bsky.feed.post"] = func (event firehoseEvent) {
		var replyParent, replyParentAuthor, replyRoot, replyRootAuthor string
		reply := event.record["reply"]
		if reply != nil {
			parent := (reply.(map[string]any))["parent"]
			if (parent != nil) {
				replyParent = parent.(map[string]any)["uri"].(string)
				replyParentAuthor = getAuthorFromPostUri(replyParent);
			}
			root := (reply.(map[string]any))["root"]
			if (parent != nil) {
				replyRoot = root.(map[string]any)["uri"].(string)
				replyRootAuthor = getAuthorFromPostUri(replyRoot);
			}
		}
		queries.SavePost(ctx, feeddb.SavePostParams{
			Uri: event.uri,
			Author: event.author,
			ReplyParent: toNullString(replyParent),
			ReplyParentAuthor: toNullString(replyParentAuthor),
			ReplyRoot: toNullString(replyRoot),
			ReplyRootAuthor: toNullString(replyRootAuthor),
			IndexedAt: time.Now().Format(time.RFC3339),
			DirectReplyCount: 0,
			InteractionCount: 0,
			LikeCount: 0,
			ReplyCount: 0,
		})
		// fmt.Println(event)
		
		res, _ := queries.ListPosts(ctx);
		fmt.Println(res)
		if reply != nil {
			panic("test")
		}
	}
	firehoseListeners["app.bsky.feed.like"] = func (event firehoseEvent) {
		// fmt.Println(event)
	}
		fmt.Println("Starting")
	subscribe(ctx, "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos", firehoseListeners)

	test := feeddb.Author{
		Did: "",
		MedianLikeCount: 0,
		MedianReplyCount: 0,
		MedianDirectReplyCount: 0,
		MedianInteractionCount: 0,
	}
	fmt.Println((test))
}
