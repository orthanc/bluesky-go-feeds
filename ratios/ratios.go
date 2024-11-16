package ratios

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/orthanc/feedgenerator/database"
	writeSchema "github.com/orthanc/feedgenerator/database/write"
)

type Ratios struct {
	database   *database.Database
	batchMutex *sync.Mutex
}

func NewRatios(database *database.Database, batchMutex *sync.Mutex) *Ratios {
	ratios := &Ratios{
		database:   database,
		batchMutex: batchMutex,
	}
	ctx := context.Background()

	ticker := time.NewTicker(6 * time.Hour)
	go func() {
		for range ticker.C {
			err := ratios.UpdateAllRatios(ctx)
			if err != nil {
				fmt.Printf("Error updating ratios: %s\n", err)
			}
		}
	}()

	return ratios
}

var recalculateInteractionScoresSql string = `
with
	"ratios" as (
		select "authorDid", "userDid", max(0.1, log(
			1000.0 * max(0, log(
				1000.0 * count(*) / (select count(*) from "post" where "author" = "userInteraction"."authorDid")
			))
		)) as "score"
		from "userInteraction"
		group by "authorDid", "userDid"
	)
update "following" set "userInteractionRatio" = (
	select "score" from "ratios" where "authorDid" = "following" and "userDid" = "followedBy"
	union all
	select 0.1
)
`

var updatePostCountsSql string = `
update "author" set "postCount" = (select count(*) from "post" where "post"."author" = "author"."did")
`

// async updateAllMedians() {
// 	const authors = await this.db.selectFrom('author').select(['did']).execute();
// 	console.log(`Updating medians for ${authors.length} authors`)
// 	for (const {did} of authors) {
// 		await this.updateAllMediansForAuthor(did);
// 	}
// 	console.log(`Author Medians updated`);

// 	const users = await this.db.selectFrom('user').select(['userDid']).execute();
// 	console.log(`Updating medians for ${users.length} users`)
// 	for (const {userDid} of users) {
// 		await this.updateInteractionCountForFollows(userDid)
// 	}
// 	console.log(`User Medians updated`);
// }

func (ratios *Ratios) UpdateAllRatios(ctx context.Context) error {
	ratios.batchMutex.Lock()
	defer ratios.batchMutex.Unlock()
	fmt.Println("Starting updating all ratios")
	authors, err := ratios.database.Queries.ListAllAuthors(ctx)
	if err != nil {
		return err
	}
	for ind, authorDid := range authors {
		err = ratios.UpdateAllMediansForAuthor(ctx, authorDid)
		if err != nil {
			return err
		}
		if ind%1000 == 0 {
			fmt.Printf("Updated %d author medians\n", ind)
			time.Sleep(250 * time.Millisecond)
		}
	}
	fmt.Println("Updating post counts")
	err = ratios.UpdatePostCounts(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Updated post counts")
	err = ratios.RecalculateInteractionScores(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Done updating all ratios")
	return nil
}

func (ratios *Ratios) UpdatePostCounts(ctx context.Context) error {
	_, err := ratios.database.ExecContext(ctx, updatePostCountsSql)
	return err
}

func (ratios *Ratios) RecalculateInteractionScores(ctx context.Context) error {
	_, err := ratios.database.ExecContext(ctx, recalculateInteractionScoresSql)
	return err
}

func (ratios *Ratios) UpdateAllMediansForAuthor(ctx context.Context, authorDid string) error {
	rows, err := ratios.database.Queries.ListPostInteractionsForAuthor(ctx, authorDid)
	if err != nil {
		return err
	}

	directReplyCounts := make([]float64, 0, len(rows))
	interactionCounts := make([]float64, 0, len(rows))
	likeCounts := make([]float64, 0, len(rows))
	replyCounts := make([]float64, 0, len(rows))
	for _, row := range rows {
		directReplyCounts = append(directReplyCounts, row.DirectReplyCount)
		interactionCounts = append(interactionCounts, row.InteractionCount)
		likeCounts = append(likeCounts, row.LikeCount)
		replyCounts = append(replyCounts, row.ReplyCount)
	}

	err = ratios.database.Updates.UpdateAuthorMedians(ctx, writeSchema.UpdateAuthorMediansParams{
		Did: authorDid,
		MedianDirectReplyCount: median(directReplyCounts, 0),
		MedianInteractionCount: median(interactionCounts, 0),
		MedianLikeCount:        median(likeCounts, 0),
		MedianReplyCount:       median(replyCounts, 0),
	})
	return err
}

func median(data []float64, def float64) float64 {
	if len(data) == 0 {
		return def
	}
	slices.Sort(data)
	return data[len(data)/2]
}
