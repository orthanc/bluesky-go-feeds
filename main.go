package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/following"
	processor "github.com/orthanc/feedgenerator/processors"
	"github.com/orthanc/feedgenerator/ratios"
	"github.com/orthanc/feedgenerator/subscription"
	"github.com/orthanc/feedgenerator/web"
)

func startProfile() (func(), error) {
	cpuProfName := filepath.Join(os.Getenv("PROFILES_DIR"), fmt.Sprintf("feedgen-cpu-%s.prof", time.Now().UTC().Format(time.RFC3339)));
	cpuF, err := os.Create(cpuProfName);
	if err != nil {
		return func () {}, err;
	}
	err = pprof.StartCPUProfile(cpuF);
	if err != nil {
		return func () {
			cpuF.Close();
		}, err
	}
	log.Printf("Started CPU profile %s\n", cpuProfName);
	stopCpuProf := func () {
		log.Printf("Finishing CPU profile %s\n", cpuProfName);
		pprof.StopCPUProfile();
		cpuF.Close()
	}

	return stopCpuProf, nil
}

func main() {
	go func() {
		stopProfile, err := startProfile();
		defer func () {stopProfile();}()
		if err != nil {
			log.Printf("Unable to start profiler %e\n", err);
		}
		ticker := time.NewTicker(1 * time.Hour);
		for range ticker.C {
			stopProfile();
			stopProfile, err = startProfile();
			if err != nil {
				log.Printf("Unable to start profiler %e\n", err);
			}
		}

	}()
	ctx := context.Background()
	dbDown := flag.Bool("db-down", false, "migrate the database down one revision then exit")
	dbUp := flag.Bool("db-up", false, "migrate the database up to the latest revision then exit")
	flag.Parse()

	database, err := database.NewDatabase(ctx)
	if err != nil {
		panic(err)
	}
	if *dbDown {
		err := database.MigrateDown(ctx)
		if err != nil {
			panic(err)
		}
		return
	}
	err = database.MigrateUp(ctx)
	if err != nil {
		panic(err)
	}
	if *dbUp {
		return
	}

	client := xrpc.Client{
		Host: "https://bsky.social",
	}
	publicClient := xrpc.Client{
		Host: "https://public.api.bsky.app",
	}
	followFarmersList := os.Getenv("FOLLOW_FARMERS_LIST")
	allFollowing := following.NewAllFollowing(
		database,
		&client,
		&publicClient,
		followFarmersList,
	)
	ratioCalc := ratios.NewRatios(database)
	go func() {
		err := allFollowing.SyncList(ctx, followFarmersList)
		if err != nil {
			fmt.Println(err)
		}
		ticker := time.NewTicker(6 * time.Hour)
		for range ticker.C {
			err := allFollowing.SyncList(ctx, followFarmersList)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	postersMadness := processor.NewPostersMadness(database);
	processingStats := subscription.NewProcessingStats()
	go web.StartServer(database, allFollowing, processingStats)
	firehoseListeners := make(map[string]subscription.JetstreamEventListener)
	firehoseListeners["app.bsky.graph.follow"] = (&processor.FollowProcessor{
		Database:     database,
		AllFollowing: allFollowing,
	}).Process
	firehoseListeners["app.bsky.feed.post"] = (&processor.PostProcessor{
		Database:       database,
		PostersMadness: postersMadness,
		PublicClient:   &publicClient,
	}).Process
	firehoseListeners["app.bsky.feed.like"] = (&processor.LikeProcessor{
		Database:     database,
		PostersMadness: postersMadness,
	}).Process
	firehoseListeners["app.bsky.feed.repost"] = (&processor.RepostProcessor{
		Database:     database,
	}).Process
	firehoseListeners["app.bsky.graph.listitem"] = processor.NewListItemProcessor(database, followFarmersList ).Process
	fmt.Println("Starting")
	err = subscription.SubscribeJetstream(ctx, os.Getenv("JETSTREAM_SUBSCRIPTION_ENDPOINT"), database, firehoseListeners, ratioCalc.Pauser, processingStats)
	if err != nil {
		panic(fmt.Sprintf("subscribing to firehose failed (dialing): %s", err))
	}
}
