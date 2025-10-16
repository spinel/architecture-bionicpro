-- Создание базы данных
CREATE DATABASE IF NOT EXISTS reports_db;

USE reports_db;

-- Таблица с агрегированными отчётами (OLAP оптимизация)
CREATE TABLE IF NOT EXISTS reports_summary
(
    report_id             String,
    user_id               String,
    username              String,
    device_id             String,
    device_type           String,
    period_start          DateTime,
    period_end            DateTime,
    total_usage_hours     Float64,
    average_daily_hours   Float64,
    battery_changes       UInt32,
    maintenance_alerts    UInt32,
    movement_activities   UInt32,
    error_count           UInt32,
    comfort_score         Float64,
    efficiency_score      Float64,
    created_at            DateTime
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(period_start)
ORDER BY (user_id, period_start)
SETTINGS index_granularity = 8192;

-- Индекс для быстрого поиска по report_id
ALTER TABLE reports_summary ADD INDEX idx_report_id report_id TYPE minmax GRANULARITY 4;

-- Комментарии к таблице
COMMENT ON TABLE reports_summary 'Агрегированные отчёты об использовании бионических протезов';

