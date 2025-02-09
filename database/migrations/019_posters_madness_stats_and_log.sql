-- +goose Up
create table posters_madness_stats (
  stage varchar not null primary key,
  cnt number not null
);

insert into
  posters_madness_stats
select
  stage,
  count(*) as cnt
from
  posters_madness
group by
  stage;

create table posters_madness_log (
  entry_id integer primary key,
  recorded_at varchar not null,
  poster_did varchar not null,
  stage varchar not null,
  comment varchar
);

-- +goose StatementBegin
create trigger posters_madness_on_insert after insert on posters_madness for each row begin
insert into
  posters_madness_stats (stage, cnt)
values
  (new.stage, 1) on conflict do
update
set
  cnt = cnt + 1;

end;

-- +goose StatementEnd
-- +goose StatementBegin
create trigger posters_madness_on_stage_change after
update of stage on posters_madness for each row begin
insert into
  posters_madness_stats (stage, cnt)
values
  (new.stage, 1) on conflict do
update
set
  cnt = cnt + 1;

update posters_madness_stats
set
  cnt = cnt - 1
where
  stage = old.stage;

end;

-- +goose StatementEnd
-- +goose StatementBegin
create trigger posters_madness_on_delete after delete on posters_madness for each row begin
update posters_madness_stats
set
  cnt = cnt - 1
where
  stage = old.stage;

end;

-- +goose StatementEnd
-- +goose Down
drop trigger posters_madness_on_delete;

drop trigger posters_madness_on_insert;

drop trigger posters_madness_on_stage_change;

drop table posters_madness_log;

drop table posters_madness_stats;