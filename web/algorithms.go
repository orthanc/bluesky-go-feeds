package web

import (
	"context"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
)

type algorithm = func(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error)

var algorithms = map[string]algorithm{
	"catchup": catchup,
	"oi8ydnb44i8y": goodStuff,
	"o1s6niihick9": lovelies,
	// Test feed
	"replies-foll": lovelies,
}
