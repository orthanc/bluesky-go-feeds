-- +goose Up
alter table user
add column lastSynced varchar;