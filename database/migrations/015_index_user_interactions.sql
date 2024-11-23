-- +goose Up
create index userInteraction_authorDid_userDid on userInteraction (authorDid, userDid);

-- +goose Down
drop index userInteraction_authorDid_userDid;