-- +goose Up
create table post (
  "uri" varchar primary key,
  "author" varchar not null,
  "replyParent" varchar,
  "replyRoot" varchar,
  "indexedAt" varchar not null,
  "likeCount" numeric not null,
  "replyCount" numeric not null,
  "directReplyCount" numeric not null,
  "interactionCount" numeric not null,
  "replyParentAuthor" varchar,
  "replyRootAuthor" varchar
);

create index "idx_post_indexedAt" on "post" ("indexedAt");

create table sub_state (
  "service" varchar primary key,
  "cursor" integer not null
);

create table author (
  "did" varchar primary key,
  "medianLikeCount" numeric not null,
  "medianReplyCount" numeric not null,
  "medianDirectReplyCount" numeric not null,
  "medianInteractionCount" numeric not null
);

create index "idx_post_author_indexedAt" on "post" ("author", "indexedAt");

create table session (
  "sessionId" integer primary key,
  "userDid" varchar not null,
  "startedAt" varchar not null,
  "postsSince" varchar not null,
  "lastSeen" varchar not null,
  "accessCount" numeric,
  "algo" varchar
);

create index "idx_session_lastSeen" on "session" ("lastSeen");

create table user (
  "userDid" varchar primary key,
  "lastSeen" varchar not null
);

create index "idx_user_lastSeen" on "user" ("lastSeen");

create table following (
  "uri" varchar primary key,
  "followedBy" varchar not null,
  "following" varchar not null,
  "userInteractionRatio" real default 0.1,
  constraint "fk_following_user" foreign key ("followedBy") references "user" ("userDid") on delete cascade
);

create index "idx_following_followedBy" on "following" ("followedBy");

create index "idx_following_following" on "following" ("following");

create table userInteraction (
  "interactionUri" varchar not null,
  "userDid" varchar not null,
  "type" varchar not null,
  "authorDid" varchar not null,
  "postUri" varchar not null,
  "indexedAt" varchar not null,
  constraint "pk_userInteraction" primary key ("interactionUri", "type")
);

create table interactionWithUser (
  "interactionUri" varchar not null,
  "userDid" varchar not null,
  "type" varchar not null,
  "interactionAuthorDid" varchar not null,
  "postUri" varchar not null,
  "indexedAt" varchar not null,
  constraint "pk_userInteraction" primary key ("interactionUri", "type")
);

create table repost (
  "uri" varchar primary key,
  "repostAuthor" varchar not null,
  "postUri" varchar not null,
  "postAuthor" varchar not null,
  "indexedAt" varchar not null
);
