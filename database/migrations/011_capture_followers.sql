-- +goose Up
create table follower (
  followed_by varchar not null,
  following varchar not null,
  mutual numeric default 0,
  last_recorded varchar not null,
  primary key (followed_by, following),
  constraint fk_follower_user foreign key (following) references user (userDid) on delete cascade
);

create index idx_follower_followed_by on follower (followed_by);

create index idx_follower_following on follower (following);

create index idx_follower_following_followed_by
on follower(following, followed_by);

alter table following
add column mutual numeric default 0;

-- +goose StatementBegin
create trigger follower_on_insert AFTER insert on follower for each row begin
  update following set mutual = mutual + 1
  where following = new.followed_by AND followedBy = new.following;

  update follower set mutual = (
    select count(*) from following where following = new.followed_by AND followedBy = new.following
  ) where followed_by = new.followed_by AND following = new.following;
end;
-- +goose StatementEnd

-- +goose StatementBegin
create trigger follower_on_delete AFTER delete on follower for each row begin
  update following set mutual = mutual - 1
  where following = old.followed_by AND followedBy = old.following;
end;
-- +goose StatementEnd

drop trigger following_on_insert;
-- +goose StatementBegin
create trigger following_on_insert AFTER insert on following for each row begin
  update post_interacted_by_followed_author set followed = followed + 1
  where author = new.following AND user = new.followedBy;

  update follower set mutual = mutual + 1
  where following = new.followedBy AND followed_by = new.following;

  update following set mutual = (
    select count(*) from follower where following = new.followedBy AND followed_by = new.following
  ) where followedBy = new.followedBy AND following = new.following;
end;
-- +goose StatementEnd

drop trigger following_on_delete;
-- +goose StatementBegin
create trigger following_on_delete AFTER delete on following for each row begin
  update post_interacted_by_followed_author set followed = followed - 1
  where author = old.following AND user = old.followedBy;

  update follower set mutual = mutual - 1
  where following = old.followedBy AND followed_by = old.following;
end;
-- +goose StatementEnd


-- +goose Down
drop trigger following_on_insert;
-- +goose StatementBegin
create trigger following_on_insert AFTER insert on following for each row begin
  update post_interacted_by_followed_author set followed = followed + 1
  where author = new.following AND user = new.followedBy;
end;
-- +goose StatementEnd

drop trigger following_on_delete;
-- +goose StatementBegin
create trigger following_on_delete AFTER delete on following for each row begin
  update post_interacted_by_followed_author set followed = followed - 1
  where author = old.following AND user = old.followedBy;
end;
-- +goose StatementEnd

drop trigger follower_on_delete;

drop trigger follower_on_insert;

alter table following
drop column mutual;

drop index idx_follower_following_followed_by;

drop index idx_follower_following;

drop index idx_follower_followed_by;

drop table follower;