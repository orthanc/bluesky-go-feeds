package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/bluesky-social/indigo/api/bsky"
)

var Hostanme = os.Getenv("FEEDGEN_HOSTNAME")
var ServiceDid = fmt.Sprintf("did:web:%s", Hostanme)
var PublisherDid = os.Getenv("FEEDGEN_PUBLISHER_DID")

func wellKnownDidHandler(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(map[string]any{
		"@context": []string{"https://www.w3.org/ns/did/v1"},
		"id":       ServiceDid,
		"service": []any{
			map[string]any{
				"id":              "#bsky_fg",
				"type":            "BskyFeedGenerator",
				"serviceEndpoint": fmt.Sprintf("https://%s", Hostanme),
			},
		},
	})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func describeFeedGenerator(w http.ResponseWriter, r *http.Request) {
	// TODO Build feeds dynamically based on algorithms
	feeds := []*bsky.FeedDescribeFeedGenerator_Feed{
		{Uri: fmt.Sprintf("at://%s/app.bsky.feed.generator/%s", PublisherDid, "replies-foll")},
	}
	data, err := json.Marshal(bsky.FeedDescribeFeedGenerator_Output{
		Did:   ServiceDid,
		Feeds: feeds,
	})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func getFeedSkeleton(w http.ResponseWriter, r *http.Request) {
	// TODO actually load feed data from database
	data, err := json.Marshal(bsky.FeedGetFeedSkeleton_Output{
		Feed: []*bsky.FeedDefs_SkeletonFeedPost{
			{
				Post: "at://did:plc:difjsauz26vnv7c5woktj4ju/app.bsky.feed.post/3l5pu5bnmrd2c",
			},
			{
				Post: "at://did:plc:l5ykap4c5bmtdodwpikl24u3/app.bsky.feed.post/3l5pu5ervzl2y",
			},
		},
	})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func StartServer() {
	http.HandleFunc("GET /.well-known/did.json", wellKnownDidHandler)
	http.HandleFunc("GET /xrpc/app.bsky.feed.describeFeedGenerator", describeFeedGenerator)
	http.HandleFunc("GET /xrpc/app.bsky.feed.getFeedSkeleton", getFeedSkeleton)

	fmt.Printf("Starting server on %s:%s\n", os.Getenv("FEEDGEN_LISTENHOST"), os.Getenv("FEEDGEN_PORT"))
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", os.Getenv("FEEDGEN_LISTENHOST"), os.Getenv("FEEDGEN_PORT")), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Closing server")
}
