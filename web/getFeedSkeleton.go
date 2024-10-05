package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
	writeSchema "github.com/orthanc/feedgenerator/database/write"
	"github.com/orthanc/feedgenerator/following"
)

type GetFeedSkeletonHandler struct {
	syncFollowingChan chan following.SyncFollowingParams
	database *database.Database
}

func NewGetFeedSkeleton(database *database.Database, syncFollowingChan chan following.SyncFollowingParams) http.Handler {
	return GetFeedSkeletonHandler{
		database: database,
		syncFollowingChan: syncFollowingChan,
	}
}

func (handler GetFeedSkeletonHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	feedUri := syntax.ATURI(r.URL.Query().Get("feed"))
	algKey := feedUri.RecordKey().String()
	alg := algorithms[algKey]
	if feedUri.Authority().String() != PublisherDid || feedUri.Collection() != "app.bsky.feed.generator" || alg == nil {
		w.WriteHeader(400)
		return
	}

	userDid, err := validateAuth(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(401)
		return
	}

	ctx := context.Background()
	lastSessionResult, err := handler.database.Queries.GetLastSession(ctx, schema.GetLastSessionParams{
		UserDid: userDid,
		Algo: database.ToNullString(algKey),
	})
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}
	lastSeen := time.Now().Format(time.RFC3339)

	fmt.Println(lastSeen, time.Now().Add(time.Duration(-1) * time.Hour).Format(time.RFC3339))
	var lastSession schema.Session
	if len(lastSessionResult) == 0 || lastSessionResult[0].LastSeen < time.Now().Add(time.Duration(-1) * time.Hour).Format(time.RFC3339) {
		updated, err := handler.database.Updates.UpdateUserLastSeen(ctx, writeSchema.UpdateUserLastSeenParams{
			UserDid: userDid,
			LastSeen: lastSeen,
		})
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
		if updated == 0 {
			handler.syncFollowingChan <- following.SyncFollowingParams{
				UserDid:  userDid,
				LastSeen: lastSeen,
			}
		}
		postsSince := time.Now().Add(time.Duration(-24) * time.Hour).Format(time.RFC3339)
		if len(lastSessionResult) > 0 {
			postsSince = lastSessionResult[0].StartedAt
		}
		lastSession = schema.Session{
			UserDid: userDid,
			StartedAt: lastSeen,
			LastSeen: lastSeen,
			PostsSince: postsSince,
			AccessCount: sql.NullFloat64{Float64: 0, Valid: true},
			Algo: database.ToNullString(algKey),
		}
		err = handler.database.Updates.SaveSession(ctx, writeSchema.SaveSessionParams{
			UserDid: lastSession.UserDid,
			StartedAt: lastSession.StartedAt,
			LastSeen: lastSession.LastSeen,
			PostsSince: lastSession.PostsSince,
			AccessCount: lastSession.AccessCount,
			Algo: lastSession.Algo,
		})
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
	} else {
		lastSession = lastSessionResult[0]
		err := handler.database.Updates.UpdateSessionLastSeen(ctx, writeSchema.UpdateSessionLastSeenParams{
			SessionId: lastSession.SessionId,
			LastSeen: lastSeen,
		})
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
	}
	fmt.Println(lastSession)
	// FIXME make sure posts since is correctly populated
	fmt.Println(lastSession.PostsSince)

	result, err := alg(lastSession)
	if err != nil {
		fmt.Printf("Error calling %s for %s: %s", feedUri, userDid, err)
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
