-- +goose Up
drop trigger post_interacted_by_followed_author_on_insert;

drop trigger post_interacted_by_followed_on_insert;

drop trigger post_interacted_by_followed_on_delete;

drop trigger post_interacted_by_followed_on_update;

delete from post_interacted_by_followed;

delete from post_interacted_by_followed_author;

-- +goose StatementBegin
create trigger post_interacted_by_followed_on_insert after insert on post_interacted_by_followed for each row begin
insert into
  post_interacted_by_followed_author (
    author,
    user,
    followed_reply_count,
    followed_direct_reply_count,
    followed_like_count,
    followed_repost_count,
    followed_interaction_count
  )
values
  (
    new.author,
    new.user,
    new.followed_reply_count,
    new.followed_direct_reply_count,
    new.followed_like_count,
    new.followed_repost_count,
    new.followed_interaction_count
  ) on conflict do
update
set
  followed_reply_count = followed_reply_count + new.followed_reply_count,
  followed_direct_reply_count = followed_direct_reply_count + new.followed_direct_reply_count,
  followed_like_count = followed_like_count + new.followed_like_count,
  followed_repost_count = followed_repost_count + new.followed_repost_count,
  followed_interaction_count = followed_interaction_count + new.followed_interaction_count;

end;

-- +goose StatementEnd
-- +goose StatementBegin
create trigger post_interacted_by_followed_on_delete after delete on post_interacted_by_followed for each row begin
update post_interacted_by_followed_author
set
  followed_reply_count = followed_reply_count - old.followed_reply_count,
  followed_direct_reply_count = followed_direct_reply_count - old.followed_direct_reply_count,
  followed_like_count = followed_like_count - old.followed_like_count,
  followed_repost_count = followed_repost_count - old.followed_repost_count,
  followed_interaction_count = followed_interaction_count - old.followed_interaction_count
where
  author = old.author
  and user = old.user;

end;

-- +goose StatementEnd
-- +goose StatementBegin
create trigger post_interacted_by_followed_on_update after
update on post_interacted_by_followed for each row begin
insert into
  post_interacted_by_followed_author (
    author,
    user,
    followed_reply_count,
    followed_direct_reply_count,
    followed_like_count,
    followed_repost_count,
    followed_interaction_count
  )
values
  (
    new.author,
    new.user,
    new.followed_reply_count,
    new.followed_direct_reply_count,
    new.followed_like_count,
    new.followed_repost_count,
    new.followed_interaction_count
  ) on conflict do
update
set
  followed_reply_count = followed_reply_count - old.followed_reply_count + new.followed_reply_count,
  followed_direct_reply_count = followed_direct_reply_count - old.followed_direct_reply_count + new.followed_direct_reply_count,
  followed_like_count = followed_like_count - old.followed_like_count + new.followed_like_count,
  followed_repost_count = followed_repost_count - old.followed_repost_count + new.followed_repost_count,
  followed_interaction_count = followed_interaction_count - old.followed_interaction_count + new.followed_interaction_count;

end;

-- +goose StatementEnd
-- +goose StatementBegin
create trigger post_interacted_by_followed_author_on_insert AFTER insert on post_interacted_by_followed_author for each row begin
update post_interacted_by_followed_author
set
  followed = (
    select
      count(*)
    from
      following
    where
      following = new.author
      AND followedBy = new.user
  )
where
  author = new.author
  AND user = new.user;

end;

-- +goose StatementEnd
-- +goose Down