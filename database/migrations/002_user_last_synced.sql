-- +goose Up
alter table user
add column lastSynced varchar;

-- +goose Down
alter table user
drop column lastSynced;