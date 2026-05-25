-- Clear existing data if any
DELETE FROM users;
DELETE FROM channel;
DELETE FROM post;
DELETE FROM postchannel;

-- Insert test user (username: testuser, password: password)
INSERT INTO users (id, username, password) VALUES (1, 'testuser', '$2a$10$0ehrzCMCLhcmjEyU1egoKev1/cg/8OtqBRDR9Cf9aFEgyb96Bx1le');
INSERT INTO users (id, username, password) VALUES (2, 'alice', '$2a$10$0ehrzCMCLhcmjEyU1egoKev1/cg/8OtqBRDR9Cf9aFEgyb96Bx1le');

-- Insert test channels
INSERT INTO channel (id, user_id, slug, label) VALUES (1, 1, 'general', 'General');
INSERT INTO channel (id, user_id, slug, label) VALUES (2, 1, 'tech', 'Tech');
INSERT INTO channel (id, user_id, slug, label) VALUES (3, 2, 'thoughts', 'Thoughts');

-- Insert test posts (using hardcoded unix timestamps and uuids)
INSERT INTO post (id, user_id, uuid, body, posted) VALUES (1, 1, '123e4567-e89b-12d3-a456-426614174000', 'Welcome to Finch! This is the first post.', strftime('%s', 'now') - 86400);
INSERT INTO post (id, user_id, uuid, body, posted) VALUES (2, 1, '223e4567-e89b-12d3-a456-426614174001', 'This is a **markdown** test.
* item 1
* item 2', strftime('%s', 'now') - 3600);
INSERT INTO post (id, user_id, uuid, body, posted) VALUES (3, 2, '323e4567-e89b-12d3-a456-426614174002', 'Alice is here. Hello world!', strftime('%s', 'now') - 7200);
INSERT INTO post (id, user_id, uuid, body, posted) VALUES (4, 1, '423e4567-e89b-12d3-a456-426614174003', 'Just another thought...', strftime('%s', 'now') - 1800);

-- Link posts to channels
INSERT INTO postchannel (post_id, channel_id) VALUES (1, 1);
INSERT INTO postchannel (post_id, channel_id) VALUES (2, 2);
INSERT INTO postchannel (post_id, channel_id) VALUES (3, 3);
INSERT INTO postchannel (post_id, channel_id) VALUES (4, 1);
INSERT INTO postchannel (post_id, channel_id) VALUES (4, 2);
