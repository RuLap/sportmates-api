-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL,
    "description" TEXT NULL,
    "start_date" TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NULL,
    city_id INT NOT NULL,
    place TEXT NOT NULL,
    photo_url TEXT NULL,
    sport_id UUID NOT NULL,
    creator_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE event_participants (
    event_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (event_id, user_id)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE event_participants;
DROP TABLE events;
-- +goose StatementEnd
