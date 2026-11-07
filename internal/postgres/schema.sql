CREATE TABLE IF NOT EXISTS rooms (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    floor TEXT NOT NULL,
    capacity INTEGER NOT NULL CHECK (capacity > 0),
    equipment TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bookings (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES rooms(id),
    title TEXT NOT NULL,
    organizer TEXT NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (ends_at > starts_at)
);

CREATE INDEX IF NOT EXISTS bookings_room_time_idx
    ON bookings (room_id, starts_at, ends_at);

INSERT INTO rooms (name, floor, capacity, equipment)
SELECT * FROM (VALUES
    ('Альфа', '2 этаж', 8, ARRAY['Экран', 'Видеосвязь']),
    ('Бета', '2 этаж', 6, ARRAY['Экран']),
    ('Гамма', '3 этаж', 4, ARRAY['Маркерная доска']),
    ('Малая', '3 этаж', 2, ARRAY[]::TEXT[])
) AS seed(name, floor, capacity, equipment)
WHERE NOT EXISTS (SELECT 1 FROM rooms);
