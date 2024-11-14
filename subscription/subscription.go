package subscription

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/orthanc/feedgenerator/database"
	writeSchema "github.com/orthanc/feedgenerator/database/write"

	"github.com/bluesky-social/jetstream/pkg/client"
	js_sequentiial "github.com/bluesky-social/jetstream/pkg/client/schedulers/sequential"
	"github.com/bluesky-social/jetstream/pkg/models"

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
type JetstreamEventListener func(context.Context, *models.Event, string) error


func SubscribeJetstream(ctx context.Context, serverAddr string, database *database.Database, listeners map[string]JetstreamEventListener) error {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})))
	logger := slog.Default()

	config := client.DefaultClientConfig()
	config.WebsocketURL = serverAddr
	config.Compress = true
	
	for collection, _ := range listeners {
		config.WantedCollections = append(config.WantedCollections, collection)
	}
	

	eventCountSinceSync := 0
	windowStart := time.Now().UTC().UnixMilli()
	var lastEvtTime int64 = 0
	scheduler := js_sequentiial.NewScheduler("feed_generator", logger, func (ctx context.Context, event *models.Event) error {
		if event.Commit != nil {
			listener := listeners[event.Commit.Collection]
			err := listener(ctx, event, fmt.Sprintf("at://%s/%s/%s", event.Did, event.Commit.Collection, event.Commit.RKey))
			if err != nil {
				return err
			}

			eventCountSinceSync++
			if eventCountSinceSync >= 1000 {
				database.Updates.SaveCursor(ctx, writeSchema.SaveCursorParams{
					Service: "jetstream",
					Cursor:  event.TimeUS,
				})
				windowEnd := time.Now().UTC().UnixMilli()
				timeSpent := windowEnd - windowStart
				parsedTime := time.UnixMicro(event.TimeUS)
				evtTime := parsedTime.UnixMilli()
				caughtUp := evtTime - lastEvtTime
				lagTime := windowEnd - evtTime
				toCatchUp := time.Duration(0)
				if lagTime > caughtUp {
					toCatchUp = time.Duration(timeSpent*lagTime/caughtUp) * time.Millisecond
				}
				fmt.Printf(
					"Processed %d events in %s (%f evts/s), %s caughtUp %s, %s behind, %s to catch up)\n",
					eventCountSinceSync,
					time.Duration(timeSpent)*time.Millisecond,
					1000.0*float64(eventCountSinceSync)/float64(timeSpent),
					parsedTime.Format(time.RFC3339),
					time.Duration(caughtUp)*time.Millisecond,
					time.Duration(lagTime)*time.Millisecond,
					toCatchUp,
				)
				windowStart = windowEnd
				lastEvtTime = evtTime
				eventCountSinceSync = 0
			}
		}
	
		return nil
	})

	client, err := client.NewClient(config, logger, scheduler)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	cursorResult, err := database.Queries.GetCursor(ctx, "jetstream")
	cursor := time.Now().Add(5 * -time.Minute).UnixMicro()
	if err != nil {
		return fmt.Errorf("unable to load cursor: %w", err)
	}
	if len(cursorResult) > 0 {
		cursor = cursorResult[0].Cursor
		fmt.Printf("Cursor from db %d\n", cursor)
	}


	if err := client.ConnectAndRead(ctx, &cursor); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	log.Print("Started")
	return nil
}

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
				parsedTime, _ := time.Parse(time.RFC3339, evt.Time)
				database.Updates.SaveCursor(ctx, writeSchema.SaveCursorParams{
					Service: service,
					Cursor:  evt.Seq,
				})
				database.Updates.SaveCursor(ctx, writeSchema.SaveCursorParams{
					Service: "jetstream",
					Cursor:  parsedTime.UnixMicro(),
				})
				windowEnd := time.Now().UTC().UnixMilli()
				timeSpent := windowEnd - windowStart
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
