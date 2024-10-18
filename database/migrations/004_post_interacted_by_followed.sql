-- +goose Up
create table post_interacted_by_followed (
  user varchar not null,
  uri varchar not null,
  author varchar not null,
  indexed_at varchar not null,
  followed_reply_count numeric not null default 0,
  followed_direct_reply_count numeric not null default 0,
  followed_like_count numeric not null default 0,
  followed_interaction_count numeric not null default 0,
  constraint post_interacted_by_followed primary key (user, uri)
);

-- +goose Down
drop table post_interacted_by_followed;