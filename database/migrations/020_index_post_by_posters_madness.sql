-- +goose Up
alter table post
add column posters_madness int2;

create index idx_post_posters_madness_indexedAt on post (posters_madness, indexedAt);

-- +goose Down
drop index idx_post_posters_madness_indexedAt;

alter table post
drop column posters_madness;