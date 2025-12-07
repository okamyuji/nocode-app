-- 外部データソース機能のマイグレーション

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
    INDEX idx_data_sources_name (name),
    INDEX idx_data_sources_created_by (created_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- appsテーブルに外部データソース関連カラムを追加
ALTER TABLE apps
    ADD COLUMN is_external BOOLEAN DEFAULT FALSE AFTER icon,
    ADD COLUMN data_source_id BIGINT UNSIGNED NULL AFTER is_external,
    ADD COLUMN source_table_name VARCHAR(100) NULL AFTER data_source_id,
    ADD CONSTRAINT fk_apps_data_source FOREIGN KEY (data_source_id) REFERENCES data_sources(id) ON DELETE SET NULL,
    ADD INDEX idx_apps_data_source_id (data_source_id);

-- app_fieldsテーブルに外部カラム名を追加
ALTER TABLE app_fields
    ADD COLUMN source_column_name VARCHAR(100) NULL AFTER field_type;

