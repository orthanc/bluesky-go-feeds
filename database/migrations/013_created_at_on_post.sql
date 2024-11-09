-- +goose Up
alter table post
add column created_at varchar;

-- +goose Down
alter table post
drop created_at;