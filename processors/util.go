package processor

import (
	"strings"
)

func getAuthorFromPostUri(postUri string) string {
	parts := strings.SplitN(postUri, "/", 5)
	if len(parts) < 5 {
		return ""
	}
	repo := parts[2]
	collection := parts[3]
	if collection != "app.bsky.feed.post" {
		return ""
	}
	return repo
}
