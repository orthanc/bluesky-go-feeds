-- +goose Up
create index idx_user_interaction_user_post on "userInteraction" ("userDid", "postUri");

-- +goose Down
drop index idx_user_interaction_user_post;