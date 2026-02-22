-- +goose Up
-- +goose StatementBegin
CREATE TABLE sports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "name" TEXT NOT NULL UNIQUE,
    icon_url TEXT NOT NULL
);

INSERT INTO sports ("name", icon_url) VALUES
('Велосипед', ''),
('Лыжи', ''),
('Йога', ''),
('Фитнес', ''),
('Спортзал', ''),
('Единоборства', ''),
('Коньки', ''),
('Хоккей', ''),
('Футбол', ''),
('Теннис', ''),
('Настольный теннис', ''),
('Падел-теннис', ''),
('Баскетбол', '');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sports;
-- +goose StatementEnd