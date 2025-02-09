package web

import (
	"context"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/orthanc/feedgenerator/database"
	schema "github.com/orthanc/feedgenerator/database/read"
)

type algorithm = func(ctx context.Context, database database.Database, session schema.Session, cursor string, limit int) (bsky.FeedGetFeedSkeleton_Output, error)

var algorithms = map[string]algorithm{
	catchupAlgorithmId:           catchup,
	catchupMutualsAlgorithmId:    catchupMutuals,
	catchupFollowersAlgorithmId:  catchupFollowers,
	goodStuffAlgorithmId:         goodStuff,
	loveliesAlgorithmId:          lovelies,
	quietPostersAlgorithmId:      quietPosters,
	youMightLikeAlgorithmId:      youMightLike,
	followerRouletteAlgorithmId:  followerRoulette,
	followingRouletteAlgorithmId: followingRoulette,
	postersMadnessAlgorithmId:    postersMadnessFeed,
	debugAlgorithmId:             debugFeed,
}
