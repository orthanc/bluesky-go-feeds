package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bluesky-social/indigo/atproto/syntax"
)

func getFeedSkeleton(w http.ResponseWriter, r *http.Request) {
	feedUri := syntax.ATURI(r.URL.Query().Get("feed"))
	alg := algorithms[feedUri.RecordKey().String()]
	if feedUri.Authority().String() != PublisherDid || feedUri.Collection() != "app.bsky.feed.generator" || alg == nil {
		w.WriteHeader(400)
		return
	}

	subjectDid, err := validateAuth(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(401)
		return
	}

	result, err := alg(subjectDid)
	if err != nil {
		fmt.Printf("Error calling %s for %s: %s", feedUri, subjectDid, err)
		w.WriteHeader(500)
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
