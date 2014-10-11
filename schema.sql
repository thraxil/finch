CREATE TABLE users (id integer primary key, username varchar(32), password varchar(256));
CREATE TABLE channel (id integer primary key, user_id integer, slug varchar(64), label varchar(64));
CREATE TABLE post (id integer primary key, uuid varchar(256), user_id integer, body text, posted integer);
CREATE TABLE postchannel (id integer primary key, post_id integer, channel_id integer);

CREATE UNIQUE INDEX IF NOT EXISTS users_username on users (username);
CREATE INDEX IF NOT EXISTS channel_slug on channel (slug);
CREATE INDEX IF NOT EXISTS channel_user_id on channel (user_id);

CREATE UNIQUE INDEX IF NOT EXISTS post_uuid on post (uuid);
CREATE INDEX IF NOT EXISTS post_user_id on post (user_id);
CREATE INDEX IF NOT EXISTS post_posted on post (posted);

CREATE INDEX IF NOT EXISTS postchannel_post_id on postchannel (post_id);
CREATE INDEX IF NOT EXISTS postchannel_channel_id on postchannel (channel_id);
