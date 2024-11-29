package subscription

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/orthanc/feedgenerator/database"
	writeSchema "github.com/orthanc/feedgenerator/database/write"
	"github.com/orthanc/feedgenerator/pauser"

	"github.com/bluesky-social/jetstream/pkg/client"
	sequentiial "github.com/bluesky-social/jetstream/pkg/client/schedulers/sequential"
	"github.com/bluesky-social/jetstream/pkg/models"
)

type JetstreamEventListener func(context.Context, *models.Event, string) error

type ProcessingRun struct {
	Timestamp time.Time
	Events int
	ProcessingTime time.Duration
	EventsPerSecond float64
	LastEventTime time.Time
	CaughtUp time.Duration
	LagTime time.Duration
	ToCatchUp time.Duration
}

type ProcessingStats struct {
	runs [3000]ProcessingRun
	end int
	start int
	lock *sync.RWMutex
}

func NewProcessingStats() *ProcessingStats {
	var lock sync.RWMutex
	return &ProcessingStats{
		lock: &lock,
	}
}

func (status *ProcessingStats) Push(run *ProcessingRun) {
	status.lock.Lock()
	defer status.lock.Unlock()
	status.runs[status.end] = *run
	status.end = (status.end + 1) % len(status.runs)
	if status.start == status.end {
		status.start = (status.end + 1) % len(status.runs)
	}
}

func (status *ProcessingStats) Iterate(cb func (run *ProcessingRun)) {
	status.lock.RLock()
	defer status.lock.RUnlock()
	for i := status.start; i != status.end; i = (i + 1) % len(status.runs) {
		cb(&status.runs[i])
	}
}


func SubscribeJetstream(ctx context.Context, serverAddr string, database *database.Database, listeners map[string]JetstreamEventListener, pauser *pauser.Pauser, stats *ProcessingStats) error {
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


	cursorResult, err := database.Queries.GetCursor(ctx, "jetstream")
	cursor := time.Now().Add(5 * -time.Minute).UnixMicro()
	if err != nil {
		return fmt.Errorf("unable to load cursor: %w", err)
	}
	if len(cursorResult) > 0 {
		cursor = cursorResult[0].Cursor
		fmt.Printf("Cursor from db %d\n", cursor)
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
				now := time.Now().UTC()
				windowEnd := now.UnixMilli()
				timeSpent := windowEnd - windowStart
				parsedTime := time.UnixMicro(event.TimeUS)
				evtTime := parsedTime.UnixMilli()
				caughtUp := evtTime - lastEvtTime
				lagTime := windowEnd - evtTime
				toCatchUp := time.Duration(0)
				if lagTime > caughtUp {
					toCatchUp = time.Duration(timeSpent*lagTime/caughtUp) * time.Millisecond
				}
				eventStats := ProcessingRun{
					Timestamp: now,
					Events: eventCountSinceSync,
					ProcessingTime:time.Duration(timeSpent)*time.Millisecond,
					EventsPerSecond: 1000.0*float64(eventCountSinceSync)/float64(timeSpent),
					LastEventTime: parsedTime,
					CaughtUp: time.Duration(caughtUp)*time.Millisecond,
					LagTime: time.Duration(lagTime)*time.Millisecond,
					ToCatchUp: toCatchUp,
				}
				fmt.Printf(
					"Processed %d events in %s (%f evts/s), %s caughtUp %s, %s behind, %s to catch up)\n",
					eventStats.Events,
					eventStats.ProcessingTime,
					eventStats.EventsPerSecond,
					eventStats.LastEventTime.Format(time.RFC3339),
					eventStats.CaughtUp,
					eventStats.LagTime,
					toCatchUp,
				)
				stats.Push(&eventStats)
				if lagTime > 60000 {
					pauser.Pause()
				} else if (lagTime < 15000) {
					pauser.Unpause()
				}	
				windowStart = windowEnd
				lastEvtTime = evtTime
				eventCountSinceSync = 0
			}
		}
		cursor = event.TimeUS
	
		return nil
	})

	for {
		fmt.Printf("Connecting at cursor %d\n", cursor)
		client, err := client.NewClient(config, logger, scheduler)
		if err == nil {
			if err := client.ConnectAndRead(ctx, &cursor); err != nil {
				fmt.Printf("failed to connect: %v\n", err)
			}
		} else {
			fmt.Printf("failed to create client: %v\n", err)
		}
		time.Sleep(time.Duration(5) * time.Second)
	}
}
