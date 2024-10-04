package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)


var Hostanme = os.Getenv("FEEDGEN_HOSTNAME")
var ServiceDid = fmt.Sprintf("did:web:%s", Hostanme)

func wellKnownDidHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(map[string]any{
		"@context": []string{"https://www.w3.org/ns/did/v1"},
		"id": ServiceDid,
		"service": []any{
			map[string]any{
				"id": "#bsky_fg",
				"type": "BskyFeedGenerator",
				"serviceEndpoint": fmt.Sprintf("https://%s", Hostanme),
			},
		},
	})
	if err != nil {
		panic(err)
	}
	w.Write(data)
}

func StartServer() {
	http.HandleFunc("GET /.well-known/did.json", wellKnownDidHandler)

	fmt.Printf("Starting server on %s:%s\n", os.Getenv("FEEDGEN_LISTENHOST"), os.Getenv("FEEDGEN_PORT"))
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", os.Getenv("FEEDGEN_LISTENHOST"), os.Getenv("FEEDGEN_PORT")), nil)
	if err != nil {
		panic(err)
	}
	
	fmt.Println("Closing server")
}