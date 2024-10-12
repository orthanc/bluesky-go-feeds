-- +goose Up
alter table author
add column "postCount" numeric default 0;

-- +goose Down
alter table author
drop column "postCount";