package subscription

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/orthanc/feedgenerator/database"
	writeSchema "github.com/orthanc/feedgenerator/database/write"

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

type FirehoseEventListener func(context.Context, FirehoseEvent) error

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

func Subscribe(initialCtx context.Context, service string, database *database.Database, listeners map[string]FirehoseEventListener) error {
	eventCountSinceSync := 0
	windowStart := time.Now().UTC().UnixMilli()
	var lastEvtTime int64 = 0
	var lastSeq int64 = 0
	ctx, cancel := context.WithCancel(initialCtx)
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			now := time.Now().UTC()
			windowOpenFor := time.Duration((now.UnixMilli() - windowStart) * time.Hour.Milliseconds())
			if windowOpenFor > time.Duration(5*time.Minute) {
				log.Printf("No traffic for %s, killing connection\n", windowOpenFor)
				oldCancel := cancel
				// Reset the context so that its not cancelled for the retry
				ctx, cancel = context.WithCancel(initialCtx)
				oldCancel()
			} else {
				log.Printf("Connection health check fine at %s, window open for %s\n", now, windowOpenFor)
			}
		}
	}()
	rsc := &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			lastSeq = evt.Seq
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
					err = listener(ctx, event)
					if err != nil {
						fmt.Printf("Unable to process %s: %s\n", collection, err)
						continue
					}
				}
			}
			eventCountSinceSync++
			if eventCountSinceSync >= 1000 {
				database.Updates.SaveCursor(ctx, writeSchema.SaveCursorParams{
					Service: service,
					Cursor:  evt.Seq,
				})
				windowEnd := time.Now().UTC().UnixMilli()
				timeSpent := windowEnd - windowStart
				parsedTime, _ := time.Parse(time.RFC3339, evt.Time)
				evtTime := parsedTime.UnixMilli()
				caughtUp := evtTime - lastEvtTime
				lagTime := windowEnd - evtTime
				toCatchUp := time.Duration(0)
				if lagTime > caughtUp {
					toCatchUp = time.Duration(timeSpent*lagTime/caughtUp) * time.Millisecond
				}
				fmt.Printf(
					"Processed %d events in %s (%f evts/s), %s caughtUp %s, %s behind, %s to catch up) %d seq\n",
					eventCountSinceSync,
					time.Duration(timeSpent)*time.Millisecond,
					1000.0*float64(eventCountSinceSync)/float64(timeSpent),
					evt.Time,
					time.Duration(caughtUp)*time.Millisecond,
					time.Duration(lagTime)*time.Millisecond,
					toCatchUp,
					evt.Seq,
				)
				windowStart = windowEnd
				lastEvtTime = evtTime
				eventCountSinceSync = 0
			}
			return nil
		},
	}
	cursorResult, err := database.Queries.GetCursor(ctx, service)
	if err != nil {
		return fmt.Errorf("unable to load cursor: %w", err)
	}
	if len(cursorResult) > 0 {
		lastSeq = cursorResult[0].Cursor
	}
	for {
		queryString := ""
		if lastSeq != 0 {
			queryString = fmt.Sprintf("?cursor=%d", lastSeq)
		}
		dialer := websocket.DefaultDialer
		uri := fmt.Sprintf("%s/xrpc/com.atproto.sync.subscribeRepos%s", service, queryString)
		fmt.Printf("Connecting to %s\n", uri)
		con, _, err := dialer.Dial(uri, http.Header{})
		if err != nil {
			return fmt.Errorf("subscribing to firehose failed (dialing): %w", err)
		}

		scheduler := sequential.NewScheduler("test", rsc.EventHandler)

		eventCountSinceSync = 0
		windowStart = time.Now().UTC().UnixMilli()
		lastEvtTime = 0
		err = events.HandleRepoStream(ctx, con, scheduler)
		if err != nil {
			fmt.Printf("Error from repo stream: %s\n", err)
			time.Sleep(time.Duration(5) * time.Second)
		}
	}
}
