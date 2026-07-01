-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS chat_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    teacher_id BIGINT,
    teacher_user_id BIGINT,
    order_id BIGINT,
    last_message_id BIGINT,
    last_message_content VARCHAR(255),
    last_message_time TIMESTAMPTZ,
    user_unread_count INTEGER DEFAULT 0,
    teacher_unread_count INTEGER DEFAULT 0,
    status SMALLINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL,
    sender_id BIGINT NOT NULL,
    sender_type VARCHAR(20) NOT NULL,
    receiver_id BIGINT,
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text',
    extra_data JSONB,
    is_read SMALLINT DEFAULT 0,
    read_time TIMESTAMPTZ,
    status SMALLINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_id ON chat_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_teacher_id ON chat_sessions(teacher_id);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_order_id ON chat_sessions(order_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_session_id ON chat_messages(session_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at ON chat_messages(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chat_sessions;
-- +goose StatementEnd
