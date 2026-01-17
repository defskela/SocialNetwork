CREATE TABLE IF NOT EXISTS social.followers (
    follower_id UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    followee_id UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, followee_id)
);

CREATE INDEX idx_followers_follower_id ON social.followers(follower_id);
CREATE INDEX idx_followers_followee_id ON social.followers(followee_id);
