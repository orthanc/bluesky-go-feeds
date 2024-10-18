-- +goose Up
create table post_interacted_by_followed_author (
  author varchar not null,
  user varchar not null,
  followed_reply_count numeric not null default 0,
  followed_direct_reply_count numeric not null default 0,
  followed_like_count numeric not null default 0,
  followed_repost_count numeric not null default 0,
  followed_interaction_count numeric not null default 0,
  constraint pk_post_interacted_by_followed_author primary key (user, author)
);

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
select
  author,
  user,
  sum(followed_reply_count),
  sum(followed_direct_reply_count),
  sum(followed_like_count),
  sum(followed_repost_count),
  sum(followed_interaction_count)
from
  post_interacted_by_followed
group by
  author,
  user;

-- +goose Down
drop trigger post_interacted_by_followed_on_insert;

drop trigger post_interacted_by_followed_on_update;

drop trigger post_interacted_by_followed_on_delete;

drop table post_interacted_by_followed_author