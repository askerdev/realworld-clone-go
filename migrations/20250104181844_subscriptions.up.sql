CREATE TABLE IF NOT EXISTS subscriptions (
  user_id BIGSERIAL NOT NULL REFERENCES users(id),
  profile_id BIGSERIAL NOT NULL REFERENCES users(id),
  PRIMARY KEY (user_id, profile_id)
);
