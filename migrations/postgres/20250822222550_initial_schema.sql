-- +goose Up
-- +goose StatementBegin
-- Create extension for ULID support
CREATE EXTENSION IF NOT EXISTS "ulid";

-- Users table
CREATE TABLE users (
    id ULID PRIMARY KEY DEFAULT gen_ulid() NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    image VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- Todos table
CREATE TABLE todos (
    id ULID PRIMARY KEY DEFAULT gen_ulid() NOT NULL,
    user_id ULID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed')),
    priority VARCHAR(10) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    due_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- Indexes for performance
CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

CREATE INDEX idx_todos_user_id ON todos(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_status ON todos(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_priority ON todos(priority) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_due_date ON todos(due_date) WHERE due_date IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_todos_created_at ON todos(created_at);
CREATE INDEX idx_todos_user_status ON todos(user_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_user_priority ON todos(user_id, priority) WHERE deleted_at IS NULL;
CREATE INDEX idx_todos_deleted_at ON todos(deleted_at);

-- Full-text search index for todos (only non-deleted)
CREATE INDEX idx_todos_search ON todos USING gin(to_tsvector('english', title || ' ' || COALESCE(description, ''))) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS todos;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "ulid";
-- +goose StatementEnd
