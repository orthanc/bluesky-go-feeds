package web

import (
	"github.com/bluesky-social/indigo/api/bsky"
	schema "github.com/orthanc/feedgenerator/database/read"
)

type algorithm = func(session schema.Session) (bsky.FeedGetFeedSkeleton_Output, error)

var algorithms = map[string]algorithm{
	"replies-foll": func(_ schema.Session) (bsky.FeedGetFeedSkeleton_Output, error) {
		return bsky.FeedGetFeedSkeleton_Output{
			Feed: []*bsky.FeedDefs_SkeletonFeedPost{
				{
					Post: "at://did:plc:difjsauz26vnv7c5woktj4ju/app.bsky.feed.post/3l5pu5bnmrd2c",
				},
				{
					Post: "at://did:plc:l5ykap4c5bmtdodwpikl24u3/app.bsky.feed.post/3l5pu5ervzl2y",
				},
			},
		}, nil
	},
}
