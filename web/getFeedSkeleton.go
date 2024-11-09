package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
	writeSchema "github.com/orthanc/feedgenerator/database/write"
	"github.com/orthanc/feedgenerator/following"
)

type GetFeedSkeletonHandler struct {
	following *following.AllFollowing
	database  *database.Database
}

func NewGetFeedSkeleton(database *database.Database, following *following.AllFollowing) http.Handler {
	return GetFeedSkeletonHandler{
		database:  database,
		following: following,
	}
}

func (handler GetFeedSkeletonHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	feedUri := syntax.ATURI(r.URL.Query().Get("feed"))
	algKey := feedUri.RecordKey().String()
	if algKey == testAlgorithmId {
		algKey = catchupAlgorithmId
	}
	alg := algorithms[algKey]
	if feedUri.Authority().String() != PublisherDid || feedUri.Collection() != "app.bsky.feed.generator" || alg == nil {
		w.WriteHeader(400)
		return
	}
	
	cursor := r.URL.Query().Get("cursor")
	limit := 30
	limitParam := r.URL.Query().Get("limit")
	if limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err != nil {
			fmt.Printf("invalid limit %s: %s\n", limitParam, err)
			w.WriteHeader(400)
		}
		limit = parsedLimit
	}

	userDid, err := validateAuth(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(401)
		return
	}
	fmt.Printf("FEED SKELETON %s %s %s(%d)\n", algKey, userDid, cursor, limit)

	ctx := context.Background()
	lastSessionResult, err := handler.database.Queries.GetLastSession(ctx, schema.GetLastSessionParams{
		UserDid: userDid,
		Algo:    database.ToNullString(algKey),
	})
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	lastSeen := time.Now().UTC().Format(time.RFC3339)

	var lastSession schema.Session
	if len(lastSessionResult) == 0 || lastSessionResult[0].LastSeen < time.Now().UTC().Add(time.Duration(-1)*time.Hour).Format(time.RFC3339) {
		updated, err := handler.database.Updates.UpdateUserLastSeen(ctx, writeSchema.UpdateUserLastSeenParams{
			UserDid:  userDid,
			LastSeen: lastSeen,
		})
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
		if len(updated) == 0 || !updated[0].LastSynced.Valid || updated[0].LastSynced.String < time.Now().UTC().Add(time.Duration(-24)*time.Hour).Format(time.RFC3339) {
			go func() {
				err := handler.following.SyncFollowing(context.Background(), userDid, lastSeen)
				if err != nil {
					fmt.Printf("Error syncing follow for %s: %s\n", userDid, err)
				}
			}()
		}
		postsSince := time.Now().UTC().Add(time.Duration(-24) * time.Hour).Format(time.RFC3339)
		if len(lastSessionResult) > 0 {
			postsSince = lastSessionResult[0].StartedAt
		}
		lastSession = schema.Session{
			UserDid:     userDid,
			StartedAt:   lastSeen,
			LastSeen:    lastSeen,
			PostsSince:  postsSince,
			AccessCount: sql.NullFloat64{Float64: 0, Valid: true},
			Algo:        database.ToNullString(algKey),
		}
		// We only save the session if there's a hint that the user is actually looking.
		// This prevents an accidental flick to the time based feeds resetting the session
		if
			cursor != "" || // if there's a cursor we're loading the second page so that's a good sign it's a real session
			limit == 1 { // Generally limit is 30, but bluesky polls with limit one while the page is open to check for more
			fmt.Printf("New session for %s/%s postsSince %s\n", userDid, algKey, lastSession.PostsSince)
			err = handler.database.Updates.SaveSession(ctx, writeSchema.SaveSessionParams{
				UserDid:     lastSession.UserDid,
				StartedAt:   lastSession.StartedAt,
				LastSeen:    lastSession.LastSeen,
				PostsSince:  lastSession.PostsSince,
				AccessCount: lastSession.AccessCount,
				Algo:        lastSession.Algo,
			})
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(500)
				return
			}
		}
	} else {
		lastSession = lastSessionResult[0]
		err := handler.database.Updates.UpdateSessionLastSeen(ctx, writeSchema.UpdateSessionLastSeenParams{
			SessionId: lastSession.SessionId,
			LastSeen:  lastSeen,
		})
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
	}

	result, err := alg(ctx, *handler.database, lastSession, cursor, limit)
	if err != nil {
		fmt.Printf("Error calling %s for %s: %s\n", feedUri, userDid, err)
		w.WriteHeader(500)
		return
	}
	if cursor == "" && len(result.Feed) == 0 && result.Cursor == nil {
		result.Feed = []*bsky.FeedDefs_SkeletonFeedPost{
			{
				Post: "at://did:plc:crngjmsdh3zpuhmd5gtgwx6q/app.bsky.feed.post/3l7anu7yoj524",
				Reason: &bsky.FeedDefs_SkeletonFeedPost_Reason{
					FeedDefs_SkeletonReasonPin: &bsky.FeedDefs_SkeletonReasonPin{},
				},
			},
		}
	}

	data, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
