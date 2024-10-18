-- +goose Up
alter table post
add column repost_count numeric not null default 0;

alter table post_interacted_by_followed
add column followed_repost_count numeric not null default 0;

-- +goose Down
alter table post_interacted_by_followed
drop column followed_repost_count;

alter table post
drop column repost_count;