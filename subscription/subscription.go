package subscription

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/orthanc/feedgenerator/database"
	writeSchema "github.com/orthanc/feedgenerator/database/write"

	"github.com/bluesky-social/jetstream/pkg/client"
	sequentiial "github.com/bluesky-social/jetstream/pkg/client/schedulers/sequential"
	"github.com/bluesky-social/jetstream/pkg/models"
)

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
	scheduler := sequentiial.NewScheduler("feed_generator", logger, func (ctx context.Context, event *models.Event) error {
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
