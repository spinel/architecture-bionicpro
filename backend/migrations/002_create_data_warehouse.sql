-- Создание витрины данных для сервиса отчётов BionicPRO
-- ========================================================

-- Создание базы данных для аналитики (если не существует)
-- CREATE DATABASE bionicpro_analytics;

-- Подключение к базе данных аналитики
-- \c bionicpro_analytics;

-- Таблица для агрегированных данных по пользователям
CREATE TABLE IF NOT EXISTS user_analytics_warehouse (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    date_created DATE NOT NULL,
    total_sessions INTEGER NOT NULL DEFAULT 0,
    avg_accuracy DECIMAL(5,2),
    avg_battery_level DECIMAL(5,2),
    avg_temperature DECIMAL(5,2),
    first_activity TIMESTAMP,
    last_activity TIMESTAMP,
    total_movements INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Индексы для быстрого доступа
    CONSTRAINT unique_user_date UNIQUE (user_id, date_created)
);

-- Создание индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_user_analytics_user_id ON user_analytics_warehouse(user_id);
CREATE INDEX IF NOT EXISTS idx_user_analytics_date ON user_analytics_warehouse(date_created);
CREATE INDEX IF NOT EXISTS idx_user_analytics_user_date ON user_analytics_warehouse(user_id, date_created);

-- Таблица для детальных данных телеметрии
CREATE TABLE IF NOT EXISTS telemetry_details_warehouse (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    device_id VARCHAR(50) NOT NULL,
    session_id VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    movement_type VARCHAR(50),
    accuracy DECIMAL(5,2),
    battery_level DECIMAL(5,2),
    temperature DECIMAL(5,2),
    sensor_status VARCHAR(20),
    calibration_date DATE,
    date_created DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для детальных данных
CREATE INDEX IF NOT EXISTS idx_telemetry_user_id ON telemetry_details_warehouse(user_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_timestamp ON telemetry_details_warehouse(timestamp);
CREATE INDEX IF NOT EXISTS idx_telemetry_user_timestamp ON telemetry_details_warehouse(user_id, timestamp);

-- Таблица для метаданных ETL процессов
CREATE TABLE IF NOT EXISTS etl_process_log (
    id SERIAL PRIMARY KEY,
    process_name VARCHAR(100) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    status VARCHAR(20) NOT NULL, -- 'running', 'success', 'failed'
    records_processed INTEGER DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Функция для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_user_analytics_updated_at 
    BEFORE UPDATE ON user_analytics_warehouse 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Представление для быстрого доступа к данным пользователей
CREATE OR REPLACE VIEW user_analytics_summary AS
SELECT 
    user_id,
    date_created,
    total_sessions,
    avg_accuracy,
    avg_battery_level,
    avg_temperature,
    first_activity,
    last_activity,
    total_movements,
    -- Дополнительные вычисляемые поля
    CASE 
        WHEN avg_accuracy >= 95 THEN 'Отличная'
        WHEN avg_accuracy >= 90 THEN 'Хорошая'
        WHEN avg_accuracy >= 80 THEN 'Удовлетворительная'
        ELSE 'Требует внимания'
    END as accuracy_status,
    
    CASE 
        WHEN avg_battery_level >= 80 THEN 'Высокий'
        WHEN avg_battery_level >= 50 THEN 'Средний'
        WHEN avg_battery_level >= 20 THEN 'Низкий'
        ELSE 'Критический'
    END as battery_status,
    
    EXTRACT(EPOCH FROM (last_activity - first_activity))/3600 as session_duration_hours
FROM user_analytics_warehouse
ORDER BY user_id, date_created DESC;

-- Представление для статистики по дням
CREATE OR REPLACE VIEW daily_analytics_summary AS
SELECT 
    date_created,
    COUNT(DISTINCT user_id) as active_users,
    AVG(total_sessions) as avg_sessions_per_user,
    AVG(avg_accuracy) as avg_accuracy_overall,
    AVG(avg_battery_level) as avg_battery_overall,
    SUM(total_movements) as total_movements_all_users
FROM user_analytics_warehouse
GROUP BY date_created
ORDER BY date_created DESC;

-- Функция для получения данных пользователя за период
CREATE OR REPLACE FUNCTION get_user_analytics(
    p_user_id VARCHAR(50),
    p_start_date DATE,
    p_end_date DATE
)
RETURNS TABLE (
    user_id VARCHAR(50),
    date_created DATE,
    total_sessions INTEGER,
    avg_accuracy DECIMAL(5,2),
    avg_battery_level DECIMAL(5,2),
    avg_temperature DECIMAL(5,2),
    first_activity TIMESTAMP,
    last_activity TIMESTAMP,
    total_movements INTEGER,
    accuracy_status TEXT,
    battery_status TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ua.user_id,
        ua.date_created,
        ua.total_sessions,
        ua.avg_accuracy,
        ua.avg_battery_level,
        ua.avg_temperature,
        ua.first_activity,
        ua.last_activity,
        ua.total_movements,
        CASE 
            WHEN ua.avg_accuracy >= 95 THEN 'Отличная'
            WHEN ua.avg_accuracy >= 90 THEN 'Хорошая'
            WHEN ua.avg_accuracy >= 80 THEN 'Удовлетворительная'
            ELSE 'Требует внимания'
        END as accuracy_status,
        CASE 
            WHEN ua.avg_battery_level >= 80 THEN 'Высокий'
            WHEN ua.avg_battery_level >= 50 THEN 'Средний'
            WHEN ua.avg_battery_level >= 20 THEN 'Низкий'
            ELSE 'Критический'
        END as battery_status
    FROM user_analytics_warehouse ua
    WHERE ua.user_id = p_user_id
    AND ua.date_created BETWEEN p_start_date AND p_end_date
    ORDER BY ua.date_created DESC;
END;
$$ LANGUAGE plpgsql;

-- Вставка тестовых данных (для демонстрации)
INSERT INTO user_analytics_warehouse (
    user_id, date_created, total_sessions, avg_accuracy,
    avg_battery_level, avg_temperature, first_activity,
    last_activity, total_movements
) VALUES 
('user_001', CURRENT_DATE, 12, 94.5, 87.2, 23.1, 
 CURRENT_TIMESTAMP - INTERVAL '8 hours', CURRENT_TIMESTAMP - INTERVAL '1 hour', 1247),
('user_002', CURRENT_DATE, 8, 91.3, 65.4, 24.2,
 CURRENT_TIMESTAMP - INTERVAL '6 hours', CURRENT_TIMESTAMP - INTERVAL '30 minutes', 892),
('user_003', CURRENT_DATE, 15, 96.8, 92.1, 22.8,
 CURRENT_TIMESTAMP - INTERVAL '10 hours', CURRENT_TIMESTAMP - INTERVAL '15 minutes', 1567)
ON CONFLICT (user_id, date_created) DO NOTHING;

-- Создание пользователя для Airflow
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'airflow_user') THEN
        CREATE USER airflow_user WITH PASSWORD 'airflow_password';
    END IF;
END
$$;

GRANT ALL PRIVILEGES ON DATABASE bionicpro TO airflow_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO airflow_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO airflow_user;
