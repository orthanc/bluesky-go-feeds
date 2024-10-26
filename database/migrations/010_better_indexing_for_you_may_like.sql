-- +goose Up
create index idx_post_interacted_by_followed_author_user_followed_author on post_interacted_by_followed_author(user, followed, author);

create index idx_post_interacted_by_followed_author_user on post_interacted_by_followed(author, user);

-- +goose Down
drop index idx_post_interacted_by_followed_author_user;

drop index idx_post_interacted_by_followed_author_user_followed_author;