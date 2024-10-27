-- +goose Up
alter table following
  add column last_recorded varchar;

-- +goose Down
alter table following
  drop column;