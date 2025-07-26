CREATE TABLE checkouts
(
    id            SERIAL PRIMARY KEY,
    user_id       TEXT        NOT NULL, -- Firebase UID
    image_path    TEXT,                 -- Path to stored image
    checkout_date timestamptz NOT NULL,
    created_at    timestamptz DEFAULT NOW()
);
