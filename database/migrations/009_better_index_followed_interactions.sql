-- +goose Up
alter table post_interacted_by_followed_author
add column followed numeric default 0;

create index idx_following_following_followedBy
on following(following, followedBy);

update post_interacted_by_followed_author
set followed = (
  select count(*) from following where following = author AND followedBy = user
);

-- +goose StatementBegin
create trigger post_interacted_by_followed_author_on_insert AFTER insert on post_interacted_by_followed_author for each row begin
  update post_interacted_by_followed_author set followed = (
    select count(*) from following where following = new.author AND followedBy = new.user
  )
  where author = new.author AND user = new.user;
end;
-- +goose StatementEnd

-- +goose StatementBegin
create trigger following_on_insert AFTER insert on following for each row begin
  update post_interacted_by_followed_author set followed = followed + 1
  where author = new.following AND user = new.followedBy;
end;
-- +goose StatementEnd

-- +goose StatementBegin
create trigger following_on_delete AFTER delete on following for each row begin
  update post_interacted_by_followed_author set followed = followed - 1
  where author = old.following AND user = old.followedBy;
end;
-- +goose StatementEnd

-- +goose Down
drop trigger following_on_delete;

drop trigger following_on_insert;

drop trigger post_interacted_by_followed_author_on_insert;

drop index idx_following_following_followedBy;

alter table post_interacted_by_followed_author
drop column followed;