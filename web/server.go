package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/following"
	"github.com/orthanc/feedgenerator/subscription"
)

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
	feeds := make([]*bsky.FeedDescribeFeedGenerator_Feed, 0, len(algorithms))
	for key := range algorithms {
		feeds = append(feeds, &bsky.FeedDescribeFeedGenerator_Feed{Uri: fmt.Sprintf("at://%s/app.bsky.feed.generator/%s", PublisherDid, key)})
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

func StartServer(database *database.Database, following *following.AllFollowing, processingStats *subscription.ProcessingStats) {
	statusPage := StatusPage{
		processingStats: processingStats,
	}
	http.HandleFunc("GET /status", statusPage.ServeHTTP)
	http.HandleFunc("GET /.well-known/did.json", wellKnownDidHandler)
	http.HandleFunc("GET /xrpc/app.bsky.feed.describeFeedGenerator", describeFeedGenerator)
	http.Handle("GET /xrpc/app.bsky.feed.getFeedSkeleton", NewGetFeedSkeleton(database, following))

	fmt.Printf("Starting server on %s:%s\n", os.Getenv("FEEDGEN_LISTENHOST"), os.Getenv("FEEDGEN_PORT"))
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", os.Getenv("FEEDGEN_LISTENHOST"), os.Getenv("FEEDGEN_PORT")), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Closing server")
}
