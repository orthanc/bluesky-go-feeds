-- +goose Up
create table followed_interaction (
  interaction_uri varchar not null,
  author varchar not null,
  interaction_type varchar not null,
  post_author_did varchar not null,
  post_uri varchar not null,
  indexed_at varchar not null,
  constraint pk_followed_interaction primary key (interaction_uri, interaction_type)
);

-- +goose Down
drop table followed_interaction;