-- +goose Up
alter table author
add column followed_by_count NUMERIC not null default 0;

alter table author
add column following_count NUMERIC not null default 0;

update author
set
  followed_by_count = (
    select
      count(*)
    from
      following
    where
      following.following = author.did
  ),
  following_count = (
    select
      count(*)
    from
      follower
    where
      follower.followed_by = author.did
  );

delete from author
where
  followed_by_count <= 0
  and following_count <= 0;

drop trigger follower_on_insert;

-- +goose StatementBegin
create trigger follower_on_insert AFTER insert on follower for each row begin
update following
set
  mutual = mutual + 1
where
  following = new.followed_by
  AND followedBy = new.following;

update follower
set
  mutual = (
    select
      count(*)
    from
      following
    where
      following = new.followed_by
      AND followedBy = new.following
  )
where
  followed_by = new.followed_by
  AND following = new.following;

update author
set
  following_count = following_count + 1
where
  did = new.followed_by;

end;

-- +goose StatementEnd
drop trigger follower_on_delete;

-- +goose StatementBegin
create trigger follower_on_delete AFTER delete on follower for each row begin
update following
set
  mutual = mutual - 1
where
  following = old.followed_by
  AND followedBy = old.following;

update author
set
  following_count = following_count - 1
where
  did = old.followed_by;

end;

-- +goose StatementEnd
drop trigger following_on_insert;

-- +goose StatementBegin
create trigger following_on_insert AFTER insert on following for each row begin
update post_interacted_by_followed_author
set
  followed = followed + 1
where
  author = new.following
  AND user = new.followedBy;

update follower
set
  mutual = mutual + 1
where
  following = new.followedBy
  AND followed_by = new.following;

update following
set
  mutual = (
    select
      count(*)
    from
      follower
    where
      following = new.followedBy
      AND followed_by = new.following
  )
where
  followedBy = new.followedBy
  AND following = new.following;

update author
set
  followed_by_count = followed_by_count + 1
where
  did = new.following;

end;

-- +goose StatementEnd
drop trigger following_on_delete;

-- +goose StatementBegin
create trigger following_on_delete AFTER delete on following for each row begin
update post_interacted_by_followed_author
set
  followed = followed - 1
where
  author = old.following
  AND user = old.followedBy;

update follower
set
  mutual = mutual - 1
where
  following = old.followedBy
  AND followed_by = old.following;

update author
set
  followed_by_count = followed_by_count - 1
where
  did = old.following;

end;

-- +goose StatementEnd
-- +goose StatementBegin
create trigger author_on_update AFTER
update on author for each row begin
delete from author
where
  did = new.did
  and followed_by_count <= 0
  and following_count <= 0;

end;

-- +goose StatementEnd
-- +goose Down
drop trigger author_on_update;

drop trigger follower_on_insert;

-- +goose StatementBegin
create trigger follower_on_insert AFTER insert on follower for each row begin
update following
set
  mutual = mutual + 1
where
  following = new.followed_by
  AND followedBy = new.following;

update follower
set
  mutual = (
    select
      count(*)
    from
      following
    where
      following = new.followed_by
      AND followedBy = new.following
  )
where
  followed_by = new.followed_by
  AND following = new.following;

end;

-- +goose StatementEnd
drop trigger follower_on_delete;

-- +goose StatementBegin
create trigger follower_on_delete AFTER delete on follower for each row begin
update following
set
  mutual = mutual - 1
where
  following = old.followed_by
  AND followedBy = old.following;

end;

-- +goose StatementEnd
drop trigger following_on_insert;

-- +goose StatementBegin
create trigger following_on_insert AFTER insert on following for each row begin
update post_interacted_by_followed_author
set
  followed = followed + 1
where
  author = new.following
  AND user = new.followedBy;

update follower
set
  mutual = mutual + 1
where
  following = new.followedBy
  AND followed_by = new.following;

update following
set
  mutual = (
    select
      count(*)
    from
      follower
    where
      following = new.followedBy
      AND followed_by = new.following
  )
where
  followedBy = new.followedBy
  AND following = new.following;

end;

-- +goose StatementEnd
drop trigger following_on_delete;

-- +goose StatementBegin
create trigger following_on_delete AFTER delete on following for each row begin
update post_interacted_by_followed_author
set
  followed = followed - 1
where
  author = old.following
  AND user = old.followedBy;

update follower
set
  mutual = mutual - 1
where
  following = old.followedBy
  AND followed_by = old.following;

end;

-- +goose StatementEnd
alter table author
drop column followed_by_count;

alter table author
drop column following_count;
