"""
BionicPRO Hourly Update DAG
===========================

Ежечасное обновление текущих данных телеметрии и быстрых метрик.
Этот DAG запускается каждый час для обновления актуальной информации.

Расписание: каждый час (0 * * * *)

Автор: BionicPRO Team
Дата: 2024
"""

from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python import PythonOperator
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
    'retries': 1,
    'retry_delay': timedelta(minutes=2),
    'catchup': False
}

# Создание DAG
dag = DAG(
    'bionicpro_hourly_update',
    default_args=default_args,
    description='Hourly update of BionicPRO telemetry data',
    schedule_interval='0 * * * *',  # Каждый час
    max_active_runs=1,
    tags=['bionicpro', 'hourly', 'telemetry', 'realtime']
)

def update_recent_sessions(**context):
    """
    Обновление данных о последних сессиях использования
    """
    logging.info("Обновляем данные о последних сессиях...")
    
    # Подключение к CRM базе данных
    crm_hook = PostgresHook(postgres_conn_id='crm_postgres')
    
    # Получение сессий за последний час
    recent_sessions_query = """
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
    WHERE us.start_time >= NOW() - INTERVAL '2 hours'
    """
    
    sessions_df = crm_hook.get_pandas_df(recent_sessions_query)
    
    if len(sessions_df) == 0:
        logging.info("Нет новых сессий за последний час")
        return {'updated_sessions': 0}
    
    # Подключение к ClickHouse
    clickhouse_hook = HttpHook(http_conn_id='clickhouse_default', method='POST')
    
    # Обновление ежедневной активности
    daily_updates = []
    for _, session in sessions_df.iterrows():
        activity_date = session['start_time'].date()
        
        # Проверяем, есть ли уже запись за этот день
        check_query = f"""
        SELECT count() FROM reports_db.daily_activity_mart 
        WHERE customer_id = '{session['customer_id']}' 
        AND device_id = '{session['device_id']}' 
        AND activity_date = '{activity_date}'
        """
        
        # Проверяем существующие записи
        response = clickhouse_hook.run(
            endpoint='/',
            data=check_query,
            headers={'Content-Type': 'text/plain'}
        )
        existing_count = int(response.text.strip()) if response.text.strip() else 0
        
        if existing_count > 0:
            # Обновляем существующую запись
            update_query = f"""
            ALTER TABLE reports_db.daily_activity_mart 
            UPDATE 
                sessions_count = sessions_count + 1,
                total_usage_minutes = total_usage_minutes + {session['duration_minutes']},
                avg_comfort_score = (avg_comfort_score * sessions_count + {session['comfort_score']}) / (sessions_count + 1),
                avg_efficiency_score = (avg_efficiency_score * sessions_count + {session['efficiency_score']}) / (sessions_count + 1),
                total_movements = total_movements + {session['movement_count']},
                errors_count = errors_count + {session['error_count']}
            WHERE customer_id = '{session['customer_id']}' 
            AND device_id = '{session['device_id']}' 
            AND activity_date = '{activity_date}'
            """
        else:
            # Создаём новую запись
            daily_updates.append((
                session['customer_id'],
                session['username'],
                session['device_id'],
                activity_date,
                1,  # sessions_count
                session['duration_minutes'],
                session['comfort_score'],
                session['efficiency_score'],
                session['movement_count'],
                session['error_count'],
                0   # battery_cycles (упрощённо)
            ))
    
    # Вставка новых записей
    if daily_updates:
        insert_sql = """
        INSERT INTO reports_db.daily_activity_mart (
            customer_id, username, device_id, activity_date,
            sessions_count, total_usage_minutes, avg_comfort_score, avg_efficiency_score,
            total_movements, errors_count, battery_cycles
        ) VALUES
        """
        # Выполняем вставку через HTTP интерфейс
        response = clickhouse_hook.run(
            endpoint='/',
            data=insert_sql,
            headers={'Content-Type': 'text/plain'}
        )
    
    logging.info(f"Обновлено {len(sessions_df)} сессий")
    return {'updated_sessions': len(sessions_df)}

def update_recent_events(**context):
    """
    Обновление данных о последних событиях телеметрии
    """
    logging.info("Обновляем данные о последних событиях...")
    
    # Подключение к CRM базе данных
    crm_hook = PostgresHook(postgres_conn_id='crm_postgres')
    
    # Получение событий за последний час
    recent_events_query = """
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
    WHERE te.event_timestamp >= NOW() - INTERVAL '2 hours'
    """
    
    events_df = crm_hook.get_pandas_df(recent_events_query)
    
    if len(events_df) == 0:
        logging.info("Нет новых событий за последний час")
        return {'updated_events': 0}
    
    # Подключение к ClickHouse
    clickhouse_hook = HttpHook(http_conn_id='clickhouse_default', method='POST')
    
    # Подготовка данных для вставки
    events_to_insert = []
    for _, event in events_df.iterrows():
        # Определение категории и серьёзности события
        event_category = 'usage'
        event_severity = 'low'
        
        if event['event_type'] in ['error_occurred', 'battery_low']:
            event_category = 'error'
            event_severity = 'medium'
        elif event['event_type'] == 'maintenance_alert':
            event_category = 'maintenance'
            event_severity = 'high'
        elif event['event_type'] == 'battery_charged':
            event_category = 'battery'
            event_severity = 'low'
        
        # Парсинг данных события
        event_description = event['event_type'].replace('_', ' ').title()
        try:
            event_data = json.loads(event['event_data']) if isinstance(event['event_data'], str) else event['event_data']
            if isinstance(event_data, dict) and 'description' in event_data:
                event_description = event_data['description']
        except:
            pass
        
        events_to_insert.append((
            event['event_id'],
            event['customer_id'],
            event['username'],
            event['device_id'],
            event['event_timestamp'].date(),
            event['event_timestamp'],
            event['event_type'],
            event_category,
            event_severity,
            event_description,
            json.dumps(event['event_data']) if event['event_data'] else '',
            False  # resolved
        ))
    
    # Вставка событий
    if events_to_insert:
        insert_sql = """
        INSERT INTO reports_db.events_mart (
            event_id, customer_id, username, device_id, event_date, event_timestamp,
            event_type, event_category, event_severity, event_description, event_data, resolved
        ) VALUES
        """
        # Выполняем вставку событий через HTTP интерфейс
        response = clickhouse_hook.run(
            endpoint='/',
            data=insert_sql,
            headers={'Content-Type': 'text/plain'}
        )
    
    logging.info(f"Обновлено {len(events_df)} событий")
    return {'updated_events': len(events_df)}

def check_system_health(**context):
    """
    Проверка здоровья системы и генерация алертов
    """
    logging.info("Проверяем здоровье системы...")
    
    # Подключение к ClickHouse
    clickhouse_hook = HttpHook(http_conn_id='clickhouse_default', method='POST')
    
    # Проверки здоровья системы
    health_checks = {
        'high_error_rate': """
        SELECT customer_id, device_id, username, error_rate 
        FROM reports_db.customer_analytics_mart 
        WHERE report_date = today() AND error_rate > 0.1
        """,
        'low_comfort_scores': """
        SELECT customer_id, device_id, username, avg_comfort_score 
        FROM reports_db.customer_analytics_mart 
        WHERE report_date = today() AND avg_comfort_score < 6.0
        """,
        'maintenance_alerts': """
        SELECT customer_id, device_id, username, maintenance_alerts 
        FROM reports_db.customer_analytics_mart 
        WHERE report_date = today() AND maintenance_alerts > 0
        """,
        'recent_critical_events': """
        SELECT customer_id, device_id, username, event_type, event_severity 
        FROM reports_db.events_mart 
        WHERE event_date = today() AND event_severity IN ('high', 'critical')
        """
    }
    
    alerts = []
    for check_name, query in health_checks.items():
        # Выполняем SQL запрос через HTTP интерфейс
        response = clickhouse_hook.run(
            endpoint='/',
            data=query,
            headers={'Content-Type': 'text/plain'}
        )
        # Парсим результат (предполагаем CSV формат)
        results = []
        if response.text.strip():
            lines = response.text.strip().split('\n')
            for line in lines:
                if line.strip():
                    results.append(line.strip().split('\t'))
        if results:
            alerts.append({
                'check': check_name,
                'count': len(results),
                'details': results[:5]  # Первые 5 записей
            })
    
    # Логирование алертов
    if alerts:
        logging.warning(f"Обнаружены проблемы в системе: {alerts}")
    else:
        logging.info("Система работает нормально")
    
    return {'alerts': alerts, 'system_healthy': len(alerts) == 0}

# Определение задач DAG

# Задача 1: Обновление сессий
update_sessions_task = PythonOperator(
    task_id='update_recent_sessions',
    python_callable=update_recent_sessions,
    dag=dag,
    doc_md="""
    ## Обновление последних сессий
    
    Обновляет данные о сессиях использования за последний час
    """
)

# Задача 2: Обновление событий
update_events_task = PythonOperator(
    task_id='update_recent_events',
    python_callable=update_recent_events,
    dag=dag,
    doc_md="""
    ## Обновление последних событий
    
    Обновляет данные о событиях телеметрии за последний час
    """
)

# Задача 3: Проверка здоровья системы
health_check_task = PythonOperator(
    task_id='check_system_health',
    python_callable=check_system_health,
    dag=dag,
    doc_md="""
    ## Проверка здоровья системы
    
    Проверяет критические метрики и генерирует алерты
    """
)

# Определение зависимостей между задачами
[update_sessions_task, update_events_task] >> health_check_task
