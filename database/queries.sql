-- name: ListPosts :many
SELECT "author",
    "directReplyCount",
    "indexedAt",
    "interactionCount",
    "likeCount",
    "replyCount",
    "uri",
    "replyParent",
    "replyParentAuthor",
    "replyRoot",
    "replyRootAuthor" FROM post;

-- name: ListAllUsers :many
select "userDid" FROM user;

-- name: ListAllFollowing :many
select * from following;