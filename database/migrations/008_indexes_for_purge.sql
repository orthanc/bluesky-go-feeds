-- +goose Up
create index idx_repost_indexedAt on repost(indexedAt);
create index idx_userInteraction_indexedAt on userInteraction(indexedAt);
create index idx_interactionWithUser_indexedAt on interactionWithUser(indexedAt);
create index idx_post_interacted_by_followed_indexed_at on post_interacted_by_followed(indexed_at);

-- +goose Down
drop index idx_repost_indexedAt;
drop index idx_userInteraction_indexedAt;
drop index idx_interactionWithUser_indexedAt;
drop index idx_post_interacted_by_followed_indexed_ad;
