package processor

import (
	"strings"

	"github.com/orthanc/feedgenerator/subscription"
)

type Processor interface {
	Process(subscription.FirehoseEvent)
}

func getAuthorFromPostUri(postUri string) string {
	parts := strings.SplitN(postUri, "/", 4)
	repo := parts[2]
	collection := parts[3]
	if collection == "app.bsky.feed.post" {
		return ""
	}
	return repo
}
