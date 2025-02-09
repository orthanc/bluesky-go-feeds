-- +goose Up
create table list_membership (
  list_uri varchar not null,
  member_did varchar not null,
  last_recorded varchar not null,
  primary key (list_uri, member_did)
);

-- +goose Down
drop table list_membership;