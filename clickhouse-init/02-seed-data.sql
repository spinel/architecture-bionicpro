USE reports_db;

-- Тестовые данные для пользователя prothetic1
INSERT INTO reports_summary 
(report_id, user_id, username, device_id, device_type, period_start, period_end, 
 total_usage_hours, average_daily_hours, battery_changes, maintenance_alerts, 
 movement_activities, error_count, comfort_score, efficiency_score, created_at)
VALUES
    -- Январь 2024
    ('rep_001', 'user_123', 'prothetic1', 'dev_arm_001', 'Bionic Arm Pro', 
     '2024-01-01 00:00:00', '2024-01-31 23:59:59',
     450.5, 14.5, 12, 2, 8500, 3, 8.7, 9.2, '2024-02-01 00:00:00'),
    
    -- Февраль 2024
    ('rep_002', 'user_123', 'prothetic1', 'dev_arm_001', 'Bionic Arm Pro',
     '2024-02-01 00:00:00', '2024-02-29 23:59:59',
     425.3, 14.7, 10, 1, 9200, 2, 8.9, 9.3, '2024-03-01 00:00:00'),
    
    -- Март 2024
    ('rep_003', 'user_123', 'prothetic1', 'dev_arm_001', 'Bionic Arm Pro',
     '2024-03-01 00:00:00', '2024-03-31 23:59:59',
     480.2, 15.5, 11, 3, 9800, 5, 8.5, 9.0, '2024-04-01 00:00:00'),
    
    -- Апрель 2024
    ('rep_004', 'user_123', 'prothetic1', 'dev_arm_001', 'Bionic Arm Pro',
     '2024-04-01 00:00:00', '2024-04-30 23:59:59',
     462.8, 15.4, 13, 2, 9500, 4, 8.8, 9.1, '2024-05-01 00:00:00'),
    
    -- Май 2024
    ('rep_005', 'user_123', 'prothetic1', 'dev_arm_001', 'Bionic Arm Pro',
     '2024-05-01 00:00:00', '2024-05-31 23:59:59',
     495.0, 16.0, 14, 1, 10200, 2, 9.0, 9.4, '2024-06-01 00:00:00');

-- Тестовые данные для пользователя prothetic2
INSERT INTO reports_summary 
(report_id, user_id, username, device_id, device_type, period_start, period_end, 
 total_usage_hours, average_daily_hours, battery_changes, maintenance_alerts, 
 movement_activities, error_count, comfort_score, efficiency_score, created_at)
VALUES
    -- Январь 2024
    ('rep_006', 'user_456', 'prothetic2', 'dev_leg_002', 'Bionic Leg Advanced', 
     '2024-01-01 00:00:00', '2024-01-31 23:59:59',
     520.0, 16.8, 8, 1, 12500, 1, 9.2, 9.5, '2024-02-01 00:00:00'),
    
    -- Февраль 2024
    ('rep_007', 'user_456', 'prothetic2', 'dev_leg_002', 'Bionic Leg Advanced',
     '2024-02-01 00:00:00', '2024-02-29 23:59:59',
     490.5, 16.9, 9, 0, 13000, 0, 9.3, 9.6, '2024-03-01 00:00:00'),
    
    -- Март 2024
    ('rep_008', 'user_456', 'prothetic2', 'dev_leg_002', 'Bionic Leg Advanced',
     '2024-03-01 00:00:00', '2024-03-31 23:59:59',
     515.8, 16.6, 10, 2, 12800, 2, 9.0, 9.4, '2024-04-01 00:00:00');

-- Тестовые данные для пользователя prothetic3
INSERT INTO reports_summary 
(report_id, user_id, username, device_id, device_type, period_start, period_end, 
 total_usage_hours, average_daily_hours, battery_changes, maintenance_alerts, 
 movement_activities, error_count, comfort_score, efficiency_score, created_at)
VALUES
    -- Январь 2024
    ('rep_009', 'user_789', 'prothetic3', 'dev_hand_003', 'Bionic Hand Elite', 
     '2024-01-01 00:00:00', '2024-01-31 23:59:59',
     380.0, 12.3, 15, 3, 6500, 4, 8.2, 8.8, '2024-02-01 00:00:00'),
    
    -- Февраль 2024
    ('rep_010', 'user_789', 'prothetic3', 'dev_hand_003', 'Bionic Hand Elite',
     '2024-02-01 00:00:00', '2024-02-29 23:59:59',
     405.2, 14.0, 12, 2, 7200, 3, 8.5, 9.0, '2024-03-01 00:00:00'),
    
    -- Март 2024
    ('rep_011', 'user_789', 'prothetic3', 'dev_hand_003', 'Bionic Hand Elite',
     '2024-03-01 00:00:00', '2024-03-31 23:59:59',
     420.5, 13.6, 14, 1, 7800, 2, 8.7, 9.2, '2024-04-01 00:00:00'),
    
    -- Апрель 2024
    ('rep_012', 'user_789', 'prothetic3', 'dev_hand_003', 'Bionic Hand Elite',
     '2024-04-01 00:00:00', '2024-04-30 23:59:59',
     435.0, 14.5, 13, 2, 8100, 3, 8.6, 9.1, '2024-05-01 00:00:00');

-- Тестовые данные для администратора (для теста ролей)
INSERT INTO reports_summary 
(report_id, user_id, username, device_id, device_type, period_start, period_end, 
 total_usage_hours, average_daily_hours, battery_changes, maintenance_alerts, 
 movement_activities, error_count, comfort_score, efficiency_score, created_at)
VALUES
    ('rep_013', 'admin_001', 'admin1', 'dev_test_001', 'Test Device', 
     '2024-01-01 00:00:00', '2024-01-31 23:59:59',
     100.0, 3.2, 5, 0, 1000, 0, 9.0, 9.0, '2024-02-01 00:00:00');

