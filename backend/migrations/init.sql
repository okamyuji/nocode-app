-- ノーコードアプリ 初期スキーマ (PostgreSQL)

-- updated_at 自動更新用の共通トリガ関数
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ユーザーテーブル
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- データソーステーブル
CREATE TABLE IF NOT EXISTS data_sources (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    db_type VARCHAR(20) NOT NULL CHECK (db_type IN ('postgresql')),
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    database_name VARCHAR(100) NOT NULL,
    username VARCHAR(100) NOT NULL,
    encrypted_password TEXT NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_data_sources_created_by ON data_sources(created_by);

DROP TRIGGER IF EXISTS trg_data_sources_updated_at ON data_sources;
CREATE TRIGGER trg_data_sources_updated_at
    BEFORE UPDATE ON data_sources
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- アプリテーブル
CREATE TABLE IF NOT EXISTS apps (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    table_name VARCHAR(64) NOT NULL UNIQUE,
    icon VARCHAR(50) DEFAULT 'default',
    is_external BOOLEAN NOT NULL DEFAULT FALSE,
    data_source_id BIGINT NULL REFERENCES data_sources(id) ON DELETE SET NULL,
    source_table_name VARCHAR(255) NULL,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_apps_created_by ON apps(created_by);
CREATE INDEX IF NOT EXISTS idx_apps_data_source_id ON apps(data_source_id);

DROP TRIGGER IF EXISTS trg_apps_updated_at ON apps;
CREATE TRIGGER trg_apps_updated_at
    BEFORE UPDATE ON apps
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- アプリフィールドテーブル
CREATE TABLE IF NOT EXISTS app_fields (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    field_code VARCHAR(64) NOT NULL,
    field_name VARCHAR(100) NOT NULL,
    field_type VARCHAR(20) NOT NULL,
    source_column_name VARCHAR(255) NULL,
    options JSONB,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_app_field_code UNIQUE (app_id, field_code)
);

CREATE INDEX IF NOT EXISTS idx_app_fields_app_id ON app_fields(app_id);

DROP TRIGGER IF EXISTS trg_app_fields_updated_at ON app_fields;
CREATE TRIGGER trg_app_fields_updated_at
    BEFORE UPDATE ON app_fields
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- アプリビューテーブル
CREATE TABLE IF NOT EXISTS app_views (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    view_type VARCHAR(20) NOT NULL DEFAULT 'table'
        CHECK (view_type IN ('table', 'list', 'calendar', 'chart')),
    config JSONB,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_app_views_app_id ON app_views(app_id);

DROP TRIGGER IF EXISTS trg_app_views_updated_at ON app_views;
CREATE TRIGGER trg_app_views_updated_at
    BEFORE UPDATE ON app_views
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- チャート設定テーブル
CREATE TABLE IF NOT EXISTS chart_configs (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    chart_type VARCHAR(20) NOT NULL,
    config JSONB NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_chart_configs_app_id ON chart_configs(app_id);

DROP TRIGGER IF EXISTS trg_chart_configs_updated_at ON chart_configs;
CREATE TRIGGER trg_chart_configs_updated_at
    BEFORE UPDATE ON chart_configs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ダッシュボードウィジェットテーブル
CREATE TABLE IF NOT EXISTS dashboard_widgets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    app_id BIGINT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    display_order INT NOT NULL DEFAULT 0,
    view_type VARCHAR(20) NOT NULL DEFAULT 'table'
        CHECK (view_type IN ('table', 'list', 'chart')),
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    widget_size VARCHAR(20) NOT NULL DEFAULT 'medium'
        CHECK (widget_size IN ('small', 'medium', 'large')),
    config JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_dashboard_user_app UNIQUE (user_id, app_id)
);

CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_user_id ON dashboard_widgets(user_id);
CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_app_id ON dashboard_widgets(app_id);
CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_user_order ON dashboard_widgets(user_id, display_order);

DROP TRIGGER IF EXISTS trg_dashboard_widgets_updated_at ON dashboard_widgets;
CREATE TRIGGER trg_dashboard_widgets_updated_at
    BEFORE UPDATE ON dashboard_widgets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- デフォルト管理者ユーザーを挿入（パスワード: admin123）
INSERT INTO users (email, password_hash, name, role) VALUES
('admin@example.com', '$2a$10$e8i3egbnenpqzZlow/3Q0.5L6uN8vNyktEYkgRdWwP13xSkCtR1re', 'Admin', 'admin')
ON CONFLICT (email) DO NOTHING;
