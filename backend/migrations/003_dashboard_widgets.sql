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
