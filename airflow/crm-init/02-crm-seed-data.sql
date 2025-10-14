-- Тестовые данные для CRM системы BionicPRO

-- Вставка клиентов
INSERT INTO customers (customer_id, username, email, first_name, last_name, phone, address, registration_date, status) VALUES
('cust_001', 'prothetic1', 'prothetic1@example.com', 'Алексей', 'Петров', '+7-900-123-4567', 'г. Москва, ул. Тверская, д. 1', '2023-01-15', 'active'),
('cust_002', 'prothetic2', 'prothetic2@example.com', 'Мария', 'Сидорова', '+7-900-234-5678', 'г. Санкт-Петербург, Невский пр., д. 10', '2023-02-20', 'active'),
('cust_003', 'prothetic3', 'prothetic3@example.com', 'Дмитрий', 'Козлов', '+7-900-345-6789', 'г. Екатеринбург, ул. Ленина, д. 25', '2023-03-10', 'active'),
('cust_004', 'admin1', 'admin1@example.com', 'Администратор', 'Системы', '+7-900-999-9999', 'г. Москва, офис', '2023-01-01', 'active')
ON CONFLICT (customer_id) DO NOTHING;

-- Вставка устройств
INSERT INTO devices (device_id, customer_id, device_type, device_model, serial_number, purchase_date, warranty_expiry, status) VALUES
('dev_001', 'cust_001', 'Bionic Arm Pro', 'BAP-2024', 'SN001234567', '2023-01-20', '2026-01-20', 'active'),
('dev_002', 'cust_001', 'Bionic Leg Advanced', 'BLA-2024', 'SN001234568', '2023-06-15', '2026-06-15', 'active'),
('dev_003', 'cust_002', 'Bionic Arm Pro', 'BAP-2024', 'SN001234569', '2023-02-25', '2026-02-25', 'active'),
('dev_004', 'cust_003', 'Bionic Hand Precision', 'BHP-2024', 'SN001234570', '2023-03-15', '2026-03-15', 'active'),
('dev_005', 'cust_003', 'Bionic Leg Advanced', 'BLA-2024', 'SN001234571', '2023-08-10', '2026-08-10', 'active')
ON CONFLICT (device_id) DO NOTHING;

-- Вставка сессий использования (последние 30 дней)
INSERT INTO usage_sessions (session_id, device_id, start_time, end_time, duration_minutes, battery_level_start, battery_level_end, comfort_score, efficiency_score, movement_count, error_count) VALUES
-- Сессии для dev_001 (prothetic1)
('sess_001', 'dev_001', '2024-01-01 08:00:00', '2024-01-01 12:00:00', 240, 100, 85, 8.5, 9.2, 1500, 1),
('sess_002', 'dev_001', '2024-01-01 14:00:00', '2024-01-01 18:00:00', 240, 85, 70, 9.1, 8.8, 1200, 0),
('sess_003', 'dev_001', '2024-01-02 09:00:00', '2024-01-02 13:00:00', 240, 100, 82, 8.8, 9.0, 1350, 2),
('sess_004', 'dev_001', '2024-01-02 15:00:00', '2024-01-02 19:00:00', 240, 82, 65, 9.2, 8.9, 1100, 0),
('sess_005', 'dev_001', '2024-01-03 10:00:00', '2024-01-03 14:00:00', 240, 100, 80, 8.7, 9.1, 1400, 1),

-- Сессии для dev_002 (prothetic1)
('sess_006', 'dev_002', '2024-01-01 07:00:00', '2024-01-01 11:00:00', 240, 100, 88, 8.9, 9.3, 2000, 0),
('sess_007', 'dev_002', '2024-01-01 13:00:00', '2024-01-01 17:00:00', 240, 88, 72, 9.0, 9.1, 1800, 1),
('sess_008', 'dev_002', '2024-01-02 08:00:00', '2024-01-02 12:00:00', 240, 100, 85, 8.8, 9.2, 1950, 0),

-- Сессии для dev_003 (prothetic2)
('sess_009', 'dev_003', '2024-01-01 09:00:00', '2024-01-01 13:00:00', 240, 100, 83, 8.6, 8.9, 1300, 1),
('sess_010', 'dev_003', '2024-01-01 15:00:00', '2024-01-01 19:00:00', 240, 83, 67, 8.9, 8.7, 1150, 0),
('sess_011', 'dev_003', '2024-01-02 10:00:00', '2024-01-02 14:00:00', 240, 100, 81, 8.7, 8.8, 1250, 2),

-- Сессии для dev_004 (prothetic3)
('sess_012', 'dev_004', '2024-01-01 11:00:00', '2024-01-01 15:00:00', 240, 100, 86, 8.4, 8.6, 1000, 1),
('sess_013', 'dev_004', '2024-01-01 17:00:00', '2024-01-01 21:00:00', 240, 86, 70, 8.6, 8.5, 950, 0),
('sess_014', 'dev_004', '2024-01-02 12:00:00', '2024-01-02 16:00:00', 240, 100, 84, 8.5, 8.7, 1050, 1),

-- Сессии для dev_005 (prothetic3)
('sess_015', 'dev_005', '2024-01-01 06:00:00', '2024-01-01 10:00:00', 240, 100, 90, 9.1, 9.4, 2200, 0),
('sess_016', 'dev_005', '2024-01-01 12:00:00', '2024-01-01 16:00:00', 240, 90, 74, 9.2, 9.3, 2100, 0),
('sess_017', 'dev_005', '2024-01-02 07:00:00', '2024-01-02 11:00:00', 240, 100, 87, 9.0, 9.2, 2150, 1)
ON CONFLICT (session_id) DO NOTHING;

-- Вставка событий телеметрии
INSERT INTO telemetry_events (event_id, device_id, event_type, event_timestamp, event_data) VALUES
-- События для dev_001
('evt_001', 'dev_001', 'battery_low', '2024-01-01 11:45:00', '{"battery_level": 15, "threshold": 20}'),
('evt_002', 'dev_001', 'movement_detected', '2024-01-01 08:15:00', '{"movement_type": "grasp", "force": 0.8}'),
('evt_003', 'dev_001', 'error_occurred', '2024-01-01 10:30:00', '{"error_code": "E001", "description": "Sensor calibration needed"}'),
('evt_004', 'dev_001', 'maintenance_alert', '2024-01-01 12:00:00', '{"alert_type": "routine", "next_service": "2024-02-01"}'),

-- События для dev_002
('evt_005', 'dev_002', 'movement_detected', '2024-01-01 07:30:00', '{"movement_type": "step", "force": 1.2}'),
('evt_006', 'dev_002', 'battery_charged', '2024-01-01 11:00:00', '{"battery_level": 100, "charge_time": 180}'),
('evt_007', 'dev_002', 'error_occurred', '2024-01-01 15:45:00', '{"error_code": "E002", "description": "Motor overheating"}'),

-- События для dev_003
('evt_008', 'dev_003', 'movement_detected', '2024-01-01 09:20:00', '{"movement_type": "grasp", "force": 0.6}'),
('evt_009', 'dev_003', 'battery_low', '2024-01-01 12:30:00', '{"battery_level": 18, "threshold": 20}'),
('evt_010', 'dev_003', 'maintenance_alert', '2024-01-01 18:00:00', '{"alert_type": "urgent", "next_service": "2024-01-15"}'),

-- События для dev_004
('evt_011', 'dev_004', 'movement_detected', '2024-01-01 11:45:00', '{"movement_type": "precision_grasp", "force": 0.4}'),
('evt_012', 'dev_004', 'error_occurred', '2024-01-01 13:15:00', '{"error_code": "E003", "description": "Sensor drift detected"}'),
('evt_013', 'dev_004', 'battery_charged', '2024-01-01 16:00:00', '{"battery_level": 100, "charge_time": 120}'),

-- События для dev_005
('evt_014', 'dev_005', 'movement_detected', '2024-01-01 06:30:00', '{"movement_type": "step", "force": 1.5}'),
('evt_015', 'dev_005', 'movement_detected', '2024-01-01 08:00:00', '{"movement_type": "jump", "force": 2.1}'),
('evt_016', 'dev_005', 'maintenance_alert', '2024-01-01 10:00:00', '{"alert_type": "routine", "next_service": "2024-03-01"}')
ON CONFLICT (event_id) DO NOTHING;
