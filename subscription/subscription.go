package subscription

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/atproto/data"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"
)

type FirehoseEvent struct {
	Uri        string
	Seq        int64
	Author     string
	Collection string
	Rid        string
	EventKind  repomgr.EventKind
	Record     map[string]any
}

type FirehoseEventListener func(FirehoseEvent)

func parseEvent(ctx context.Context, evt *atproto.SyncSubscribeRepos_Commit, op *atproto.SyncSubscribeRepos_RepoOp) (FirehoseEvent, error) {
	parts := strings.SplitN(op.Path, "/", 3)
	collection, rid := parts[0], parts[1]
	eventKind := repomgr.EventKind(op.Action)
	event := FirehoseEvent{
		Uri:        fmt.Sprintf("at://%s/%s", evt.Repo, op.Path),
		Seq:        evt.Seq,
		Author:     evt.Repo,
		Collection: collection,
		Rid:        rid,
		EventKind:  eventKind,
	}
	switch eventKind {
	case repomgr.EvtKindCreateRecord, repomgr.EvtKindUpdateRecord:
		rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
		if err != nil {
			return event, fmt.Errorf("error reading %s car: %s", collection, err)
		}
		_, recCBOR, err := rr.GetRecordBytes(ctx, op.Path)
		if err != nil {
			return event, fmt.Errorf("error reading %s bytes: %s", collection, err)
		}
		d, err := data.UnmarshalCBOR(*recCBOR)
		if err != nil {
			return event, fmt.Errorf("error unmarshaling %s: %s", collection, err)
		}

		event.Record = d
		return event, nil
	default:
		return event, nil
	}
}

func Subscribe(ctx context.Context, url string, listeners map[string]FirehoseEventListener) {
	dialer := websocket.DefaultDialer
	con, _, err := dialer.Dial(url, http.Header{})
	if err != nil {
		panic(fmt.Sprintf("subscribing to firehose failed (dialing): %w", err))
	}

	rsc := &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			for _, op := range evt.Ops {
				parts := strings.SplitN(op.Path, "/", 3)
				collection := parts[0]
				listener := listeners[collection]
				if listener != nil {
					event, err := parseEvent(ctx, evt, op)
					if err != nil {
						fmt.Println(err)
						continue
					}
					listener(event)
				}
			}
			return nil
		},
	}
	scheduler := sequential.NewScheduler("test", rsc.EventHandler)

	events.HandleRepoStream(ctx, con, scheduler)

	fmt.Println("hi")
}
