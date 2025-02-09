// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package write

import (
	"database/sql"
)

type Author struct {
	Did                    string
	MedianLikeCount        float64
	MedianReplyCount       float64
	MedianDirectReplyCount float64
	MedianInteractionCount float64
	PostCount              sql.NullFloat64
	FollowedByCount        float64
	FollowingCount         float64
}

type FollowedInteraction struct {
	InteractionUri  string
	Author          string
	InteractionType string
	PostAuthorDid   string
	PostUri         string
	IndexedAt       string
}

type Follower struct {
	FollowedBy   string
	Following    string
	Mutual       sql.NullFloat64
	LastRecorded string
}

type Following struct {
	Uri                  string
	FollowedBy           string
	Following            string
	UserInteractionRatio sql.NullFloat64
	Mutual               sql.NullFloat64
	LastRecorded         sql.NullString
}

type InteractionWithUser struct {
	InteractionUri       string
	UserDid              string
	Type                 string
	InteractionAuthorDid string
	PostUri              string
	IndexedAt            string
}

type ListMembership struct {
	ListUri      string
	MemberDid    string
	LastRecorded string
}

type Post struct {
	Uri               string
	Author            string
	ReplyParent       sql.NullString
	ReplyRoot         sql.NullString
	IndexedAt         string
	LikeCount         float64
	ReplyCount        float64
	DirectReplyCount  float64
	InteractionCount  float64
	ReplyParentAuthor sql.NullString
	ReplyRootAuthor   sql.NullString
	RepostCount       float64
	CreatedAt         sql.NullString
}

type PostInteractedByFollowed struct {
	User                     string
	Uri                      string
	Author                   string
	IndexedAt                string
	FollowedReplyCount       float64
	FollowedDirectReplyCount float64
	FollowedLikeCount        float64
	FollowedInteractionCount float64
	FollowedRepostCount      float64
}

type PostInteractedByFollowedAuthor struct {
	Author                   string
	User                     string
	FollowedReplyCount       float64
	FollowedDirectReplyCount float64
	FollowedLikeCount        float64
	FollowedRepostCount      float64
	FollowedInteractionCount float64
	Followed                 sql.NullFloat64
}

type PostersMadness struct {
	PosterDid   string
	Stage       string
	LastChecked string
}

type PostersMadnessLog struct {
	EntryID    int64
	RecordedAt string
	PosterDid  string
	Stage      string
	Comment    sql.NullString
}

type PostersMadnessStat struct {
	Stage string
	Cnt   interface{}
}

type Repost struct {
	Uri          string
	RepostAuthor string
	PostUri      string
	PostAuthor   string
	IndexedAt    string
}

type Session struct {
	SessionId   int64
	UserDid     string
	StartedAt   string
	PostsSince  string
	LastSeen    string
	AccessCount sql.NullFloat64
	Algo        sql.NullString
}

type SubState struct {
	Service string
	Cursor  int64
}

type User struct {
	UserDid    string
	LastSeen   string
	LastSynced sql.NullString
}

type UserInteraction struct {
	InteractionUri string
	UserDid        string
	Type           string
	AuthorDid      string
	PostUri        string
	IndexedAt      string
}
