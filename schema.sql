CREATE TABLE users (id integer primary key, username varchar(32), password varchar(256));
CREATE TABLE channel (id integer primary key, user_id integer, slug varchar(64), label varchar(64));
CREATE TABLE post (id integer primary key, uuid varchar(256), user_id integer, body text, posted integer);
CREATE TABLE postchannel (id integer primary key, post_id integer, channel_id integer);
