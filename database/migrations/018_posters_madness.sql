-- +goose Up
create table posters_madness (
  poster_did varchar not null primary key,
  stage varchar not null,
  last_checked varchar not null
);

create index idx_posters_madness_last_checked on posters_madness (last_checked);

-- +goose Down
drop index idx_posters_madness_last_checked;

drop table posters_madness;