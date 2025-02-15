package processor

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/database/read"
	"github.com/orthanc/feedgenerator/database/write"
)

const InteractionInfectionProbability float64 = 0.004;
const IncubatingToPreInfectiousProbability float64 = 0.05;
const PreInfectiousToInfectiousProbability float64 = 0.05;
const InfectiousToPostInfectiouscProbability float64 = 0.05;
const PostInfectiousToImmuneProbability float64 = 0.01;
const ImmuneToDeleteProbability float64 = 0.05;

const StageIncubating string = "incubating";
const StagePreInfectious string = "pre-infectious";
const StageInfectious string = "infectious";
const StagePostInfectious string = "post-infectious";
const StageImmune string = "immune";

type PostersMadness struct {
	database     *database.Database
}

func NewPostersMadness(database *database.Database) *PostersMadness {
	madness := &PostersMadness{
		database:            database,
	}

	ctx := context.Background()
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			err := madness.UpdateStages(ctx)
			if err != nil {
				fmt.Printf("Error updating stages: %s\n", err)
			}
			err = madness.PurgeImmune(ctx)
			if err != nil {
				fmt.Printf("Error purging immune: %s\n", err)
			}
		}
	}()

	return madness
}

func (madness *PostersMadness) PostersMadnessInteraction(ctx context.Context, interactionByDid string, interactionWithDid string) error {
	if interactionByDid == interactionWithDid {
		return nil;
	}
	statusResult, err := madness.database.Queries.GetPostersMadnessStatus(ctx, []string{interactionByDid, interactionWithDid});
	if err != nil {
		return err
	}
	// Either neither is infections or both already in stage in which case no action
	if len(statusResult) != 1 {
		return nil;
	}
	poster := statusResult[0];
	// The poster isn't infectious, so return
	if poster.Stage != StageInfectious {
		return nil;
	}
	otherPoster := interactionByDid
	if poster.PosterDid == interactionByDid {
		otherPoster = interactionWithDid;
	}
	// The poster wasn't infected
	if rand.Float64() > InteractionInfectionProbability {
		return nil;
	}
	fmt.Printf("Interaction between %s and infectious %s, INFECTION!\n", otherPoster, poster.PosterDid)
	timestamp := time.Now().UTC().Format(time.RFC3339);
	err = madness.database.Updates.SavePostersMadness(ctx, write.SavePostersMadnessParams{
		PosterDid: otherPoster,
		Stage: StageIncubating,
		LastChecked: timestamp,
	});
	if err != nil {
		return err;
	}
	err = madness.database.Updates.SavePostersMadnessLog(ctx, write.SavePostersMadnessLogParams{
		RecordedAt: timestamp,
		PosterDid: otherPoster,
		Stage: StageIncubating,
		Comment: database.ToNullString(fmt.Sprintf("Infected by %s", poster.PosterDid)),
	})
	return err;
}

func (madness *PostersMadness) UpdateStages(ctx context.Context) error {
	now := time.Now().UTC()
	updateBefore := now.Add(-1 * time.Hour).Format(time.RFC3339)
	nowRFC3339 := now.Format(time.RFC3339)
	toUpdate, err := madness.database.Queries.GetPostersMadnessNotUpdatedSince(ctx, read.GetPostersMadnessNotUpdatedSinceParams{
		LastChecked: updateBefore,
		Stage: StageImmune,
	});
	if err != nil {
		return err;
	}
	fmt.Printf("Checking stage of %d posters\n", len(toUpdate))
	var updates []write.UpdatePostersMadnessStageParams;
	var logs []write.SavePostersMadnessLogParams;
	for _, poster := range toUpdate {
		if (poster.Stage == StageIncubating && rand.Float64() <= IncubatingToPreInfectiousProbability) {
			updates = append(updates, write.UpdatePostersMadnessStageParams{
				Stage: StagePreInfectious,
				LastChecked: nowRFC3339,
				PosterDid: poster.PosterDid,
			});
			logs = append(logs, write.SavePostersMadnessLogParams{
				RecordedAt: nowRFC3339,
				PosterDid: poster.PosterDid,
				Stage: StagePreInfectious,
			})
		} else if (poster.Stage == StagePreInfectious && rand.Float64() <= PreInfectiousToInfectiousProbability) {
			updates = append(updates, write.UpdatePostersMadnessStageParams{
				Stage: StageInfectious,
				LastChecked: nowRFC3339,
				PosterDid: poster.PosterDid,
			});
			logs = append(logs, write.SavePostersMadnessLogParams{
				RecordedAt: nowRFC3339,
				PosterDid: poster.PosterDid,
				Stage: StageInfectious,
			})
		} else if (poster.Stage == StageInfectious && rand.Float64() <= InfectiousToPostInfectiouscProbability) {
			updates = append(updates, write.UpdatePostersMadnessStageParams{
				Stage: StagePostInfectious,
				LastChecked: nowRFC3339,
				PosterDid: poster.PosterDid,
			});
			logs = append(logs, write.SavePostersMadnessLogParams{
				RecordedAt: nowRFC3339,
				PosterDid: poster.PosterDid,
				Stage: StagePostInfectious,
			})
		} else if (poster.Stage == StagePostInfectious && rand.Float64() <= PostInfectiousToImmuneProbability) {
			updates = append(updates, write.UpdatePostersMadnessStageParams{
				Stage: StageImmune,
				LastChecked: nowRFC3339,
				PosterDid: poster.PosterDid,
			});
			logs = append(logs, write.SavePostersMadnessLogParams{
				RecordedAt: nowRFC3339,
				PosterDid: poster.PosterDid,
				Stage: StageImmune,
			})
		} else {
			updates = append(updates, write.UpdatePostersMadnessStageParams{
				Stage: poster.Stage,
				LastChecked: nowRFC3339,
				PosterDid: poster.PosterDid,
			});
		}
	}
	updater, tx, err := madness.database.BeginTx(ctx)
	if err != nil {
		return err;
	}
	defer tx.Rollback();
	for _, update := range updates {
		err := updater.UpdatePostersMadnessStage(ctx, update);
		if (err != nil) {
			return err;
		}
	}
	for _, log := range logs {
		err := updater.SavePostersMadnessLog(ctx, log);
		if (err != nil) {
			return err;
		}
	}
	tx.Commit();
	return nil;
}

func (madness *PostersMadness) PurgeImmune(ctx context.Context) error {
	now := time.Now().UTC()
	updateBefore := now.Add(-24 * time.Hour).Format(time.RFC3339)
	nowRFC3339 := now.Format(time.RFC3339)
	toUpdate, err := madness.database.Queries.GetPostersMadnessInStageNotUpdatedSince(ctx, read.GetPostersMadnessInStageNotUpdatedSinceParams{
		LastChecked: updateBefore,
		Stage: StageImmune,
	});
	if err != nil {
		return err;
	}
	var updates []write.UpdatePostersMadnessStageParams;
	var deletes []string;
	var logs []write.SavePostersMadnessLogParams;
	for _, poster := range toUpdate {
		if (rand.Float64() <= ImmuneToDeleteProbability) {
			deletes = append(deletes, poster.PosterDid);
			logs = append(logs, write.SavePostersMadnessLogParams{
				RecordedAt: nowRFC3339,
				PosterDid: poster.PosterDid,
				Stage: "immunity_waned",
			})
		} else {
			updates = append(updates, write.UpdatePostersMadnessStageParams{
				Stage: poster.Stage,
				LastChecked: nowRFC3339,
				PosterDid: poster.PosterDid,
			});
		}
	}
	updater, tx, err := madness.database.BeginTx(ctx)
	if err != nil {
		return err;
	}
	defer tx.Rollback();
	for _, update := range updates {
		err := updater.UpdatePostersMadnessStage(ctx, update);
		if (err != nil) {
			return err;
		}
	}
	for _, didToDelete := range deletes {
		err := updater.DeletePostersMadness(ctx, didToDelete);
		if (err != nil) {
			return err;
		}
	}
	for _, log := range logs {
		err := updater.SavePostersMadnessLog(ctx, log);
		if (err != nil) {
			return err;
		}
	}
	tx.Commit();
	return nil;
}