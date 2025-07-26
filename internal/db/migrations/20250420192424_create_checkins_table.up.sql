CREATE TABLE checkins
(
    id           SERIAL PRIMARY KEY,
    user_id      TEXT        NOT NULL,
    image_path   TEXT,                 -- Path to stored image
    checkin_date timestamptz NOT NULL,
    created_at   timestamptz DEFAULT NOW()
);
