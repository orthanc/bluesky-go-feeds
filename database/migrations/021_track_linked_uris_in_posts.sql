-- +goose Up
alter table post
add column quoted_post_uri varchar;

alter table post
add column external_uri varchar;

-- +goose Down
alter table post
drop column quoted_post_uri;

alter table post
drop column external_uri;