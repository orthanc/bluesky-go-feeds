package web

import (
	"fmt"
	"time"

	schema "github.com/orthanc/feedgenerator/database/read"
)

func logFeedAccess(name string, session schema.Session) func() {
	fmt.Printf("[FEED] %s for %s since %s\n", name, session.UserDid, session.PostsSince)
	start := time.Now()
	return func() {
		fmt.Printf("[FEED-END] %s for %s since %s took %s\n", name, session.UserDid, session.PostsSince, time.Since(start))
	}
}
