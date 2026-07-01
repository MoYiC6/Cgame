-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS game_rooms (
    id BIGSERIAL PRIMARY KEY,
    room_code VARCHAR(10) NOT NULL UNIQUE,
    host_user_id BIGINT NOT NULL,
    status VARCHAR(20) DEFAULT 'waiting',
    max_players INTEGER DEFAULT 4,
    current_position INTEGER DEFAULT 1,
    game_over BOOLEAN DEFAULT FALSE,
    current_turn_user_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS game_room_players (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES game_rooms(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    nickname VARCHAR(100) NOT NULL,
    avatar VARCHAR(500),
    is_host BOOLEAN DEFAULT FALSE,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    left_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS game_records (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES game_rooms(id) ON DELETE CASCADE,
    room_code VARCHAR(10) NOT NULL,
    host_user_id BIGINT NOT NULL,
    final_position INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'playing',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS game_moves (
    id BIGSERIAL PRIMARY KEY,
    record_id BIGINT NOT NULL REFERENCES game_records(id) ON DELETE CASCADE,
    room_id BIGINT NOT NULL REFERENCES game_rooms(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    dice INTEGER NOT NULL CHECK (dice >= 1 AND dice <= 6),
    from_position INTEGER NOT NULL,
    to_position INTEGER NOT NULL,
    cell_text VARCHAR(255),
    special_type VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS game_maps (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(500),
    sort INTEGER DEFAULT 0,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bomb_rankings (
    id BIGSERIAL PRIMARY KEY,
    rank_position INTEGER NOT NULL UNIQUE CHECK (rank_position >= 1 AND rank_position <= 5),
    teacher_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS teacher_map_permissions (
    id BIGSERIAL PRIMARY KEY,
    teacher_id BIGINT NOT NULL,
    map_id BIGINT NOT NULL REFERENCES game_maps(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(teacher_id, map_id)
);

CREATE TABLE IF NOT EXISTS goods_maps (
    id BIGSERIAL PRIMARY KEY,
    goods_id BIGINT NOT NULL,
    game_map_id BIGINT NOT NULL REFERENCES game_maps(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(goods_id, game_map_id)
);

CREATE INDEX IF NOT EXISTS idx_game_rooms_room_code ON game_rooms(room_code);
CREATE INDEX IF NOT EXISTS idx_game_room_players_room_id ON game_room_players(room_id);
CREATE INDEX IF NOT EXISTS idx_game_records_room_id ON game_records(room_id);
CREATE INDEX IF NOT EXISTS idx_game_moves_room_id ON game_moves(room_id);
CREATE INDEX IF NOT EXISTS idx_bomb_rankings_teacher_id ON bomb_rankings(teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_map_permissions_teacher_id ON teacher_map_permissions(teacher_id);
CREATE INDEX IF NOT EXISTS idx_goods_maps_goods_id ON goods_maps(goods_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS goods_maps;
DROP TABLE IF EXISTS teacher_map_permissions;
DROP TABLE IF EXISTS bomb_rankings;
DROP TABLE IF EXISTS game_maps;
DROP TABLE IF EXISTS game_moves;
DROP TABLE IF EXISTS game_records;
DROP TABLE IF EXISTS game_room_players;
DROP TABLE IF EXISTS game_rooms;
-- +goose StatementEnd
