-- WCA profile picture URL for a user. Stores the avatar URL returned by the WCA
-- /me endpoint (the image itself stays on WCA — we only keep the link). Nullable:
-- not every WCA account has an avatar, and older rows predate this column.
ALTER TABLE users ADD COLUMN avatar_url TEXT;
