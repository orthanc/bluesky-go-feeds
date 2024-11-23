package ratios

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/orthanc/feedgenerator/database"
	"github.com/orthanc/feedgenerator/pauser"
)

type Ratios struct {
	database *database.Database
	Pauser   *pauser.Pauser
}

func NewRatios(database *database.Database) *Ratios {
	ratios := &Ratios{
		database: database,
		Pauser:   pauser.NewPauser(),
	}
	ctx := context.Background()

	ticker := time.NewTicker(6 * time.Hour)
	go func() {
		err := ratios.UpdateAllRatios(ctx)
		if err != nil {
			fmt.Printf("Error updating ratios: %s\n", err)
		}
		for range ticker.C {
			err := ratios.UpdateAllRatios(ctx)
			if err != nil {
				fmt.Printf("Error updating ratios: %s\n", err)
			}
		}
	}()

	return ratios
}

var selectInteractionRationsToUpdate string = `
select
	following.following,
	following.followedBy,
	IFNULL(max(0.1, log(
        1000.0 * max(0, log(
            1000.0 * count(*) / (select count(*) from "post" where "author" = "userInteraction"."authorDid")
        ))
    )), 0.1) as "score"
FROM following
LEFT JOIN userInteraction ON userInteraction.authorDid = following.following and userInteraction.userDid = following.followedBy
WHERE following.following IN (%s)
group by following.following, following.followedBy
`

const updateAuthorMedians = `update author
set
  "postCount" = ?,
  "medianInteractionCount" = ?
where
  "did" = ?
`
const updateInteractionRatio = `update following
set
  "userInteractionRatio" = ?
where
	following = ?
	AND followedBy = ?
`

type updateAuthorMediansParams struct {
	postCount              float64
	medianInteractionCount float64
	did                    string
}

func (ratios *Ratios) getAuthorStats(ctx context.Context, batch []string) ([]updateAuthorMediansParams, error) {
	query := fmt.Sprintf(
		"select author, count(*), median(interactionCount) from post where author IN (%s) group by author",
		"'"+strings.Join(batch, "','")+"'")
	rows, err := ratios.database.QueryContext(ctx, query)
	toSave := make([]updateAuthorMediansParams, len(batch))
	if err != nil {
		return toSave, fmt.Errorf("error calculating author stats: %s", err)
	}
	defer rows.Close()
	ind := 0
	for rows.Next() {
		row := toSave[ind]
		err := rows.Scan(
			&row.did,
			&row.postCount,
			&row.medianInteractionCount,
		)
		ind++
		if err != nil {
			return toSave, err
		}
	}
	return toSave, nil
}

type updateInteractionRationParam struct {
	authorDid string
	userDid   string
	score     float64
}

func (ratios *Ratios) getInteractionRatiosToUpdate(ctx context.Context, authorDids []string) ([]updateInteractionRationParam, error) {
	query := fmt.Sprintf(selectInteractionRationsToUpdate, "'"+strings.Join(authorDids, "','")+"'")
	rows, err := ratios.database.QueryContext(ctx, query)
	toSave := make([]updateInteractionRationParam, 0, len(authorDids))
	if err != nil {
		return toSave, fmt.Errorf("error calculating interaction ratios: %s", err)
	}
	defer rows.Close()
	ind := 0
	for rows.Next() {
		var row updateInteractionRationParam
		toSave = append(toSave, row)
		err := rows.Scan(
			&row.authorDid,
			&row.userDid,
			&row.score,
		)
		ind++
		if err != nil {
			return toSave, err
		}
	}
	return toSave, nil
}

func (ratios *Ratios) updateAuthorBatch(ctx context.Context, batch []string) error {
	authorsToUpdate, err := ratios.getAuthorStats(ctx, batch)
	if err != nil {
		return err
	}
	interactionRatiosToUpdate, err := ratios.getInteractionRatiosToUpdate(ctx, batch)
	if err != nil {
		return err
	}
	_, tx, err := ratios.database.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, row := range authorsToUpdate {
		_, err := tx.ExecContext(ctx, updateAuthorMedians, row.postCount, row.medianInteractionCount, row.did)
		if err != nil {
			return err
		}
	}
	for _, row := range interactionRatiosToUpdate {
		_, err := tx.ExecContext(ctx, updateInteractionRatio, row.score, row.authorDid, row.userDid)
		if err != nil {
			return err
		}
	}
	tx.Commit()
	return nil
}

func (ratios *Ratios) UpdateAllRatios(ctx context.Context) error {
	fmt.Println("Starting updating all ratios")
	authors, err := ratios.database.Queries.ListAllAuthors(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Updating author medians counts")
	for i := 0; i < len(authors); i += 1000 {
		ratios.Pauser.Wait()
		start := time.Now()
		batch := authors[i:]
		if len(batch) > 1000 {
			batch = batch[:1000]
		}
		err := ratios.updateAuthorBatch(ctx, batch)
		if err != nil {
			return fmt.Errorf("error calculating author stats: %s", err)
		}
		fmt.Printf("Updated %d author medians in %s\n", i+len(batch), time.Since(start))
		// time.Sleep(200 * time.Millisecond)
	}
	fmt.Println("Done updating all ratios")
	return nil
}
