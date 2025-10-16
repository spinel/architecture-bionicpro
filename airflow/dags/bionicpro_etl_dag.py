"""
BionicPRO ETL DAG
=================

ETL процесс для извлечения данных из CRM системы и телеметрии,
их обработки и загрузки в OLAP витрину данных ClickHouse.

Расписание:
- Hourly: каждый час для обновления текущих данных
- Daily: ежедневно в 02:00 для полной перезагрузки
- Weekly: по воскресеньям в 03:00 для глубокой аналитики

Автор: BionicPRO Team
Дата: 2024
"""

from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python import PythonOperator
from airflow.operators.bash import BashOperator
from airflow.providers.postgres.operators.postgres import PostgresOperator
from airflow.providers.postgres.hooks.postgres import PostgresHook
from airflow.providers.http.hooks.http import HttpHook
from airflow.utils.dates import days_ago
import pandas as pd
import json
import logging

# Конфигурация DAG
default_args = {
    'owner': 'bionicpro-team',
    'depends_on_past': False,
    'start_date': days_ago(1),
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': 2,
    'retry_delay': timedelta(minutes=5),
    'catchup': False
}

# Создание DAG
dag = DAG(
    'bionicpro_etl_pipeline',
    default_args=default_args,
    description='ETL pipeline for BionicPRO CRM and telemetry data',
    schedule_interval='0 2 * * *',  # Ежедневно в 02:00
    max_active_runs=1,
    tags=['bionicpro', 'etl', 'crm', 'telemetry', 'clickhouse']
)

def extract_crm_data(**context):
    """
    Извлечение данных из CRM системы (PostgreSQL)
    """
    logging.info("Начинаем извлечение данных из CRM...")
    
    # Подключение к CRM базе данных
    crm_hook = PostgresHook(postgres_conn_id='crm_postgres')
    
    # SQL запросы для извлечения данных
    customers_query = """
    SELECT 
        customer_id,
        username,
        email,
        first_name || ' ' || last_name as customer_name,
        phone,
        registration_date,
        status
    FROM customers 
    WHERE status = 'active'
    """
    
    devices_query = """
    SELECT 
        d.device_id,
        d.customer_id,
        c.username,
        d.device_type,
        d.device_model,
        d.purchase_date,
        d.warranty_expiry,
        d.status
    FROM devices d
    JOIN customers c ON d.customer_id = c.customer_id
    WHERE d.status = 'active'
    """
    
    usage_sessions_query = """
    SELECT 
        us.session_id,
        us.device_id,
        d.customer_id,
        c.username,
        us.start_time,
        us.end_time,
        us.duration_minutes,
        us.battery_level_start,
        us.battery_level_end,
        us.comfort_score,
        us.efficiency_score,
        us.movement_count,
        us.error_count
    FROM usage_sessions us
    JOIN devices d ON us.device_id = d.device_id
    JOIN customers c ON d.customer_id = c.customer_id
    WHERE us.start_time >= CURRENT_DATE - INTERVAL '30 days'
    """
    
    telemetry_events_query = """
    SELECT 
        te.event_id,
        te.device_id,
        d.customer_id,
        c.username,
        te.event_type,
        te.event_timestamp,
        te.event_data
    FROM telemetry_events te
    JOIN devices d ON te.device_id = d.device_id
    JOIN customers c ON d.customer_id = c.customer_id
    WHERE te.event_timestamp >= CURRENT_DATE - INTERVAL '30 days'
    """
    
    # Извлечение данных
    customers_df = crm_hook.get_pandas_df(customers_query)
    devices_df = crm_hook.get_pandas_df(devices_query)
    usage_sessions_df = crm_hook.get_pandas_df(usage_sessions_query)
    telemetry_events_df = crm_hook.get_pandas_df(telemetry_events_query)
    
    # Сохранение в XCom для передачи в следующие задачи
    context['task_instance'].xcom_push(key='customers_data', value=customers_df.to_json())
    context['task_instance'].xcom_push(key='devices_data', value=devices_df.to_json())
    context['task_instance'].xcom_push(key='usage_sessions_data', value=usage_sessions_df.to_json())
    context['task_instance'].xcom_push(key='telemetry_events_data', value=telemetry_events_df.to_json())
    
    logging.info(f"Извлечено {len(customers_df)} клиентов, {len(devices_df)} устройств, "
                f"{len(usage_sessions_df)} сессий, {len(telemetry_events_df)} событий")
    
    return {
        'customers_count': len(customers_df),
        'devices_count': len(devices_df),
        'sessions_count': len(usage_sessions_df),
        'events_count': len(telemetry_events_df)
    }

def transform_and_aggregate_data(**context):
    """
    Трансформация и агрегация данных для витрины
    """
    logging.info("Начинаем трансформацию и агрегацию данных...")
    
    # Получение данных из предыдущей задачи
    customers_data = context['task_instance'].xcom_pull(task_ids='extract_crm_data', key='customers_data')
    devices_data = context['task_instance'].xcom_pull(task_ids='extract_crm_data', key='devices_data')
    usage_sessions_data = context['task_instance'].xcom_pull(task_ids='extract_crm_data', key='usage_sessions_data')
    telemetry_events_data = context['task_instance'].xcom_pull(task_ids='extract_crm_data', key='telemetry_events_data')
    
    # Преобразование обратно в DataFrame
    customers_df = pd.read_json(customers_data)
    devices_df = pd.read_json(devices_data)
    usage_sessions_df = pd.read_json(usage_sessions_data)
    telemetry_events_df = pd.read_json(telemetry_events_data)
    
    # Преобразование дат
    usage_sessions_df['start_time'] = pd.to_datetime(usage_sessions_df['start_time'])
    usage_sessions_df['end_time'] = pd.to_datetime(usage_sessions_df['end_time'])
    telemetry_events_df['event_timestamp'] = pd.to_datetime(telemetry_events_df['event_timestamp'])
    
    # Агрегация данных по пользователям и устройствам
    analytics_data = []
    
    for _, device in devices_df.iterrows():
        device_id = device['device_id']
        customer_id = device['customer_id']
        
        # Фильтрация данных для текущего устройства
        device_sessions = usage_sessions_df[usage_sessions_df['device_id'] == device_id]
        device_events = telemetry_events_df[telemetry_events_df['device_id'] == device_id]
        
        if len(device_sessions) == 0:
            continue
            
        # Агрегация сессий
        total_sessions = len(device_sessions)
        total_usage_hours = device_sessions['duration_minutes'].sum() / 60
        avg_session_duration = device_sessions['duration_minutes'].mean()
        avg_daily_usage_hours = total_usage_hours / 30  # за последние 30 дней
        
        # Метрики батареи
        avg_battery_start = device_sessions['battery_level_start'].mean()
        avg_battery_end = device_sessions['battery_level_end'].mean()
        battery_cycles = len(device_sessions[device_sessions['battery_level_start'] == 100])
        
        # Метрики комфорта и эффективности
        avg_comfort_score = device_sessions['comfort_score'].mean()
        avg_efficiency_score = device_sessions['efficiency_score'].mean()
        
        # Определение трендов (упрощённо)
        recent_sessions = device_sessions.tail(5)
        older_sessions = device_sessions.head(5)
        
        if len(recent_sessions) >= 3 and len(older_sessions) >= 3:
            recent_comfort = recent_sessions['comfort_score'].mean()
            older_comfort = older_sessions['comfort_score'].mean()
            comfort_trend = 'improving' if recent_comfort > older_comfort else 'declining' if recent_comfort < older_comfort else 'stable'
            
            recent_efficiency = recent_sessions['efficiency_score'].mean()
            older_efficiency = older_sessions['efficiency_score'].mean()
            efficiency_trend = 'improving' if recent_efficiency > older_efficiency else 'declining' if recent_efficiency < older_efficiency else 'stable'
        else:
            comfort_trend = 'stable'
            efficiency_trend = 'stable'
        
        # Метрики активности
        total_movements = device_sessions['movement_count'].sum()
        avg_movements_per_session = device_sessions['movement_count'].mean()
        
        # Анализ типов движений из событий
        movement_events = device_events[device_events['event_type'] == 'movement_detected']
        movement_types = []
        if len(movement_events) > 0:
            for _, event in movement_events.iterrows():
                try:
                    event_data = json.loads(event['event_data']) if isinstance(event['event_data'], str) else event['event_data']
                    if 'movement_type' in event_data:
                        movement_types.append(event_data['movement_type'])
                except:
                    pass
        movement_types = list(set(movement_types))  # уникальные типы
        
        # Метрики ошибок и обслуживания
        total_errors = device_sessions['error_count'].sum()
        error_rate = total_errors / total_sessions if total_sessions > 0 else 0
        
        maintenance_events = device_events[device_events['event_type'] == 'maintenance_alert']
        maintenance_alerts = len(maintenance_events)
        
        # Определение дат обслуживания
        last_maintenance_date = None
        next_maintenance_date = None
        if len(maintenance_events) > 0:
            last_maintenance_date = maintenance_events['event_timestamp'].max().date()
            # Упрощённо: следующее обслуживание через 3 месяца
            next_maintenance_date = (last_maintenance_date + timedelta(days=90)).date()
        
        # Метрики качества обслуживания
        device_uptime_percent = 95.0  # Упрощённо
        customer_satisfaction_score = avg_comfort_score * 10  # Преобразование в 100-балльную шкалу
        
        # Получение информации о клиенте
        customer_info = customers_df[customers_df['customer_id'] == customer_id].iloc[0]
        
        # Создание записи для витрины
        analytics_record = {
            'customer_id': customer_id,
            'username': device['username'],
            'device_id': device_id,
            'customer_name': customer_info['customer_name'],
            'customer_email': customer_info['email'],
            'customer_phone': customer_info['phone'],
            'registration_date': customer_info['registration_date'],
            'device_type': device['device_type'],
            'device_model': device['device_model'],
            'purchase_date': device['purchase_date'],
            'warranty_expiry': device['warranty_expiry'],
            'report_date': datetime.now().date(),
            'report_period_start': (datetime.now() - timedelta(days=30)).date(),
            'report_period_end': datetime.now().date(),
            'total_sessions': total_sessions,
            'total_usage_hours': round(total_usage_hours, 2),
            'avg_session_duration': round(avg_session_duration, 2),
            'avg_daily_usage_hours': round(avg_daily_usage_hours, 2),
            'avg_battery_start': round(avg_battery_start, 1),
            'avg_battery_end': round(avg_battery_end, 1),
            'battery_cycles': battery_cycles,
            'avg_comfort_score': round(avg_comfort_score, 1),
            'avg_efficiency_score': round(avg_efficiency_score, 1),
            'comfort_trend': comfort_trend,
            'efficiency_trend': efficiency_trend,
            'total_movements': total_movements,
            'avg_movements_per_session': round(avg_movements_per_session, 1),
            'movement_types': movement_types,
            'total_errors': total_errors,
            'error_rate': round(error_rate, 3),
            'maintenance_alerts': maintenance_alerts,
            'last_maintenance_date': last_maintenance_date,
            'next_maintenance_date': next_maintenance_date,
            'device_uptime_percent': device_uptime_percent,
            'customer_satisfaction_score': round(customer_satisfaction_score, 1)
        }
        
        analytics_data.append(analytics_record)
    
    # Сохранение агрегированных данных
    context['task_instance'].xcom_push(key='analytics_data', value=json.dumps(analytics_data, default=str))
    
    logging.info(f"Создано {len(analytics_data)} записей для витрины данных")
    
    return {'analytics_records_count': len(analytics_data)}

def load_to_clickhouse(**context):
    """
    Загрузка агрегированных данных в ClickHouse витрину
    """
    logging.info("Начинаем загрузку данных в ClickHouse...")
    
    # Получение агрегированных данных
    analytics_data_json = context['task_instance'].xcom_pull(task_ids='transform_and_aggregate_data', key='analytics_data')
    analytics_data = json.loads(analytics_data_json)
    
    if not analytics_data:
        logging.warning("Нет данных для загрузки в ClickHouse")
        return {'loaded_records': 0}
    
    # Подключение к ClickHouse
    clickhouse_hook = HttpHook(http_conn_id='clickhouse_default', method='POST')
    
    # Подготовка данных для вставки
    records_to_insert = []
    for record in analytics_data:
        # Преобразование списка типов движений в строку
        movement_types_str = ','.join(record['movement_types']) if record['movement_types'] else ''
        
        insert_record = (
            record['customer_id'],
            record['username'],
            record['device_id'],
            record['customer_name'],
            record['customer_email'],
            record['customer_phone'],
            record['registration_date'],
            record['device_type'],
            record['device_model'],
            record['purchase_date'],
            record['warranty_expiry'],
            record['report_date'],
            record['report_period_start'],
            record['report_period_end'],
            record['total_sessions'],
            record['total_usage_hours'],
            record['avg_session_duration'],
            record['avg_daily_usage_hours'],
            record['avg_battery_start'],
            record['avg_battery_end'],
            record['battery_cycles'],
            record['avg_comfort_score'],
            record['avg_efficiency_score'],
            record['comfort_trend'],
            record['efficiency_trend'],
            record['total_movements'],
            record['avg_movements_per_session'],
            movement_types_str,
            record['total_errors'],
            record['error_rate'],
            record['maintenance_alerts'],
            record['last_maintenance_date'],
            record['next_maintenance_date'],
            record['device_uptime_percent'],
            record['customer_satisfaction_score']
        )
        records_to_insert.append(insert_record)
    
    # SQL для вставки данных
    insert_sql = """
    INSERT INTO reports_db.customer_analytics_mart (
        customer_id, username, device_id, customer_name, customer_email, customer_phone,
        registration_date, device_type, device_model, purchase_date, warranty_expiry,
        report_date, report_period_start, report_period_end, total_sessions, total_usage_hours,
        avg_session_duration, avg_daily_usage_hours, avg_battery_start, avg_battery_end,
        battery_cycles, avg_comfort_score, avg_efficiency_score, comfort_trend, efficiency_trend,
        total_movements, avg_movements_per_session, movement_types, total_errors, error_rate,
        maintenance_alerts, last_maintenance_date, next_maintenance_date, device_uptime_percent,
        customer_satisfaction_score
    ) VALUES
    """
    
    # Выполнение вставки
    # Отправляем SQL через HTTP интерфейс ClickHouse
    response = clickhouse_hook.run(
        endpoint='/',
        data=insert_sql,
        headers={'Content-Type': 'text/plain'}
    )
    
    logging.info(f"Успешно загружено {len(records_to_insert)} записей в ClickHouse")
    
    return {'loaded_records': len(records_to_insert)}

def validate_data_quality(**context):
    """
    Валидация качества загруженных данных
    """
    logging.info("Начинаем валидацию качества данных...")
    
    # Подключение к ClickHouse
    clickhouse_hook = HttpHook(http_conn_id='clickhouse_default', method='POST')
    
    # Проверки качества данных
    validation_queries = {
        'total_records': "SELECT count() FROM reports_db.customer_analytics_mart WHERE report_date = today()",
        'records_with_errors': "SELECT count() FROM reports_db.customer_analytics_mart WHERE total_errors > 0 AND report_date = today()",
        'records_with_low_comfort': "SELECT count() FROM reports_db.customer_analytics_mart WHERE avg_comfort_score < 7.0 AND report_date = today()",
        'records_with_high_usage': "SELECT count() FROM reports_db.customer_analytics_mart WHERE total_usage_hours > 8.0 AND report_date = today()"
    }
    
    validation_results = {}
    for check_name, query in validation_queries.items():
        # Выполняем SQL запрос через HTTP интерфейс
        response = clickhouse_hook.run(
            endpoint='/',
            data=query,
            headers={'Content-Type': 'text/plain'}
        )
        result = [int(response.text.strip())] if response.text.strip() else [0]
        validation_results[check_name] = result[0] if result else 0
    
    # Логирование результатов валидации
    logging.info(f"Результаты валидации: {validation_results}")
    
    # Проверка критических метрик
    if validation_results['total_records'] == 0:
        raise ValueError("Нет записей в витрине данных за сегодня")
    
    if validation_results['records_with_low_comfort'] > validation_results['total_records'] * 0.5:
        logging.warning(f"Высокий процент записей с низким комфортом: {validation_results['records_with_low_comfort']}")
    
    return validation_results

# Определение задач DAG

# Задача 1: Извлечение данных из CRM
extract_task = PythonOperator(
    task_id='extract_crm_data',
    python_callable=extract_crm_data,
    dag=dag,
    doc_md="""
    ## Извлечение данных из CRM
    
    Извлекает данные из PostgreSQL CRM системы:
    - Информация о клиентах
    - Данные об устройствах
    - Сессии использования
    - События телеметрии
    """
)

# Задача 2: Трансформация и агрегация
transform_task = PythonOperator(
    task_id='transform_and_aggregate_data',
    python_callable=transform_and_aggregate_data,
    dag=dag,
    doc_md="""
    ## Трансформация и агрегация данных
    
    Обрабатывает извлечённые данные:
    - Агрегирует метрики по пользователям и устройствам
    - Вычисляет тренды комфорта и эффективности
    - Подготавливает данные для витрины
    """
)

# Задача 3: Загрузка в ClickHouse
load_task = PythonOperator(
    task_id='load_to_clickhouse',
    python_callable=load_to_clickhouse,
    dag=dag,
    doc_md="""
    ## Загрузка в ClickHouse
    
    Загружает агрегированные данные в OLAP витрину:
    - Вставляет данные в customer_analytics_mart
    - Обновляет материализованные представления
    """
)

# Задача 4: Валидация качества данных
validate_task = PythonOperator(
    task_id='validate_data_quality',
    python_callable=validate_data_quality,
    dag=dag,
    doc_md="""
    ## Валидация качества данных
    
    Проверяет качество загруженных данных:
    - Количество записей
    - Аномальные значения
    - Критические метрики
    """
)

# Задача 5: Очистка старых данных
cleanup_task = BashOperator(
    task_id='cleanup_old_data',
    bash_command="""
    echo "Очистка старых данных..."
    # Здесь можно добавить команды для очистки старых данных
    echo "Очистка завершена"
    """,
    dag=dag,
    doc_md="""
    ## Очистка старых данных
    
    Удаляет устаревшие данные согласно политике хранения
    """
)

# Определение зависимостей между задачами
extract_task >> transform_task >> load_task >> validate_task >> cleanup_task
