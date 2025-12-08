-- ノーコードアプリ 初期スキーマ

-- ユーザーテーブル
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    role ENUM('admin', 'user') DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- データソーステーブル
CREATE TABLE IF NOT EXISTS data_sources (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    db_type ENUM('postgresql', 'mysql', 'oracle', 'sqlserver') NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    database_name VARCHAR(100) NOT NULL,
    username VARCHAR(100) NOT NULL,
    encrypted_password TEXT NOT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_data_sources_created_by (created_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- アプリテーブル
CREATE TABLE IF NOT EXISTS apps (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    table_name VARCHAR(64) NOT NULL UNIQUE,
    icon VARCHAR(50) DEFAULT 'default',
    is_external BOOLEAN DEFAULT FALSE,
    data_source_id BIGINT UNSIGNED NULL,
    source_table_name VARCHAR(255) NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (data_source_id) REFERENCES data_sources(id) ON DELETE SET NULL,
    INDEX idx_apps_created_by (created_by),
    INDEX idx_apps_data_source_id (data_source_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- アプリフィールドテーブル
CREATE TABLE IF NOT EXISTS app_fields (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    app_id BIGINT UNSIGNED NOT NULL,
    field_code VARCHAR(64) NOT NULL,
    field_name VARCHAR(100) NOT NULL,
    field_type VARCHAR(20) NOT NULL,
    source_column_name VARCHAR(255) NULL,
    options JSON,
    required BOOLEAN DEFAULT FALSE,
    display_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
    UNIQUE KEY uk_app_field_code (app_id, field_code),
    INDEX idx_app_fields_app_id (app_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- アプリビューテーブル
CREATE TABLE IF NOT EXISTS app_views (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    app_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    view_type ENUM('table', 'list', 'calendar', 'chart') NOT NULL DEFAULT 'table',
    config JSON,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
    INDEX idx_app_views_app_id (app_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- チャート設定テーブル
CREATE TABLE IF NOT EXISTS chart_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    app_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    chart_type VARCHAR(20) NOT NULL,
    config JSON NOT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_chart_configs_app_id (app_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ダッシュボードウィジェットテーブル
-- ユーザーごとのダッシュボード設定を管理
CREATE TABLE IF NOT EXISTS dashboard_widgets (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    app_id BIGINT UNSIGNED NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    view_type ENUM('table', 'list', 'chart') NOT NULL DEFAULT 'table',
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    widget_size ENUM('small', 'medium', 'large') NOT NULL DEFAULT 'medium',
    config JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
    UNIQUE KEY uk_user_app (user_id, app_id),
    INDEX idx_dashboard_widgets_user_id (user_id),
    INDEX idx_dashboard_widgets_app_id (app_id),
    INDEX idx_dashboard_widgets_display_order (user_id, display_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- デフォルト管理者ユーザーを挿入（パスワード: admin123）
INSERT INTO users (email, password_hash, name, role) VALUES 
('admin@example.com', '$2a$10$e8i3egbnenpqzZlow/3Q0.5L6uN8vNyktEYkgRdWwP13xSkCtR1re', 'Admin', 'admin');
