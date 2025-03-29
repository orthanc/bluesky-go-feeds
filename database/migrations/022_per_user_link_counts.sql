-- +goose Up
create table user_link (
  user_did varchar not null,
  link_uri varchar not null,
  share_count numeric not null default 1,
  first_seen varchar not null,
  last_seen varchar not null,
  first_seen_post_uri varchar not null,
  last_seen_post_uri varchar not null,
  constraint pk_user_link primary key (user_did, link_uri),
  constraint fk_user_link_user foreign key (user_did) references user (userDid) on delete cascade
);

create index idx_user_link_user_share_count_first_seen on user_link (
  user_did,
  share_count,
  first_seen,
  first_seen_post_uri
);

create index idx_user_link_last_seen on user_link (last_seen);

-- +goose Down
drop index idx_user_link_last_seen;

drop index idx_user_link_user_share_count_first_seen;

drop table user_link;