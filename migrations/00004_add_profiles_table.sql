-- +goose Up
-- +goose StatementBegin
CREATE TABLE profiles (
    id UUID PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    gender TEXT NOT NULL,
    birth_date DATE NOT NULL,
    city_id int NOT NULL,
    avatar_url TEXT,
    "description" TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_sports (
    user_id UUID NOT NULL,
    sport_id UUID NOT NULL,
    PRIMARY KEY (user_id, sport_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_sports;
DROP TABLE IF EXISTS profiles;
-- +goose StatementEnd