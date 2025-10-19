"""
Скрипт для инициализации подключений в Airflow
"""

from airflow.models import Connection
from airflow.utils.db import create_session
import logging

def init_airflow_connections():
    """
    Инициализация подключений к базам данных и API
    """
    
    connections = [
        {
            'conn_id': 'crm_api',
            'conn_type': 'http',
            'host': 'http://crm-service:8001',
            'port': 8001,
            'description': 'CRM API для получения данных клиентов'
        },
        {
            'conn_id': 'data_warehouse_conn',
            'conn_type': 'postgres',
            'host': 'app_db',
            'port': 5432,
            'login': 'bionicpro_user',
            'password': 'bionicpro_password',
            'schema': 'bionicpro_analytics',
            'description': 'Подключение к витрине данных'
        },
        {
            'conn_id': 'clickhouse_conn',
            'conn_type': 'postgres',
            'host': 'clickhouse',
            'port': 9000,
            'login': 'default',
            'password': '',
            'schema': 'bionicpro_analytics',
            'description': 'Подключение к ClickHouse для телеметрии'
        },
        {
            'conn_id': 'airflow_db_conn',
            'conn_type': 'postgres',
            'host': 'airflow_db',
            'port': 5432,
            'login': 'airflow',
            'password': 'airflow',
            'schema': 'airflow',
            'description': 'Подключение к базе данных Airflow'
        }
    ]
    
    with create_session() as session:
        for conn_config in connections:
            # Проверяем, существует ли подключение
            existing_conn = session.query(Connection).filter(
                Connection.conn_id == conn_config['conn_id']
            ).first()
            
            if existing_conn:
                logging.info(f"Подключение {conn_config['conn_id']} уже существует")
                continue
            
            # Создаем новое подключение
            conn = Connection(
                conn_id=conn_config['conn_id'],
                conn_type=conn_config['conn_type'],
                host=conn_config.get('host'),
                port=conn_config.get('port'),
                login=conn_config.get('login'),
                password=conn_config.get('password'),
                schema=conn_config.get('schema'),
                description=conn_config.get('description')
            )
            
            session.add(conn)
            logging.info(f"Создано подключение: {conn_config['conn_id']}")
        
        session.commit()
        logging.info("Все подключения инициализированы")

if __name__ == "__main__":
    init_airflow_connections()
