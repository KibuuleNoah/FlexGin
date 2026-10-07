-- +goose Up
INSERT INTO users (id,username, phone_no, avatar_url, is_banned, verified)
VALUES
  ('Moxie1','MoxiePlus', '+256700111222', 'default1.png', FALSE, TRUE),
  ('Moxie2','TriMoxie', '+256700333444', 'default2.png', FALSE, TRUE);

-- +goose Down
