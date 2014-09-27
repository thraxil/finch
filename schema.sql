CREATE TABLE users (id integer primary key, username varchar(32), password varchar(256));
CREATE TABLE channel (id integer primary key, user_id integer, slug varchar(64), label varchar(64));
CREATE TABLE post (id integer primary key, user_id integer, body text, posted integer);
CREATE TABLE postchannel (post_id integer primary key, channel_id integer);
