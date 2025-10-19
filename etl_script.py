#!/usr/bin/env python3
"""
ETL скрипт для BionicPRO
========================

Простой ETL-процесс для извлечения данных из CRM и подготовки витрины данных
для сервиса отчётов о работе протезов.

Автор: BionicPRO Team
Дата: 2025
"""

import psycopg2
import requests
import pandas as pd
import logging
from datetime import datetime, date
import json

# Настройка логирования
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class BionicProETL:
    def __init__(self):
        # Конфигурация подключений
        self.crm_api_url = "http://localhost:8001"
        self.db_config = {
            'host': 'localhost',
            'port': 5434,
            'database': 'bionicpro',
            'user': 'bionicpro_user',
            'password': 'bionicpro_password'
        }
        
    def extract_crm_data(self):
        """
        Извлечение данных из CRM-системы
        """
        logger.info("Начинаем извлечение данных из CRM...")
        
        try:
            # Получаем данные клиентов
            customers_response = requests.get(f"{self.crm_api_url}/api/customers")
            customers_data = customers_response.json() if customers_response.status_code == 200 else []
            
            # Получаем данные о протезах
            prosthetics_response = requests.get(f"{self.crm_api_url}/api/prosthetics")
            prosthetics_data = prosthetics_response.json() if prosthetics_response.status_code == 200 else []
            
            # Получаем данные о сессиях
            sessions_response = requests.get(f"{self.crm_api_url}/api/sessions")
            sessions_data = sessions_response.json() if sessions_response.status_code == 200 else []
            
            logger.info(f"Извлечено {len(customers_data)} клиентов, {len(prosthetics_data)} протезов, {len(sessions_data)} сессий")
            
            return {
                'customers': customers_data,
                'prosthetics': prosthetics_data,
                'sessions': sessions_data
            }
            
        except Exception as e:
            logger.error(f"Ошибка при извлечении данных из CRM: {str(e)}")
            return {'customers': [], 'prosthetics': [], 'sessions': []}
    
    def extract_telemetry_data(self):
        """
        Извлечение данных телеметрии (имитация)
        """
        logger.info("Начинаем извлечение данных телеметрии...")
        
        # Имитируем данные телеметрии
        telemetry_data = [
            {
                'user_id': 'user_001',
                'device_id': 'device_001',
                'session_id': 'S001',
                'timestamp': datetime.now().isoformat(),
                'movement_type': 'grasp',
                'accuracy': 94.5,
                'battery_level': 87.2,
                'temperature': 23.1,
                'sensor_status': 'normal',
                'calibration_date': '2024-12-10'
            },
            {
                'user_id': 'user_002',
                'device_id': 'device_002',
                'session_id': 'S002',
                'timestamp': datetime.now().isoformat(),
                'movement_type': 'release',
                'accuracy': 91.3,
                'battery_level': 65.4,
                'temperature': 24.2,
                'sensor_status': 'normal',
                'calibration_date': '2024-12-08'
            }
        ]
        
        logger.info(f"Извлечено {len(telemetry_data)} записей телеметрии")
        return telemetry_data
    
    def transform_data(self, crm_data, telemetry_data):
        """
        Трансформация и объединение данных
        """
        logger.info("Начинаем трансформацию данных...")
        
        try:
            # Создаем DataFrame для обработки
            customers_df = pd.DataFrame(crm_data['customers'])
            prosthetics_df = pd.DataFrame(crm_data['prosthetics'])
            sessions_df = pd.DataFrame(crm_data['sessions'])
            telemetry_df = pd.DataFrame(telemetry_data)
            
            # Объединяем данные
            # 1. Объединяем клиентов с протезами
            if not customers_df.empty and not prosthetics_df.empty:
                customer_prosthetics = pd.merge(
                    customers_df, 
                    prosthetics_df, 
                    on='customer_id', 
                    how='inner'
                )
            else:
                customer_prosthetics = pd.DataFrame()
            
            # 2. Объединяем с сессиями
            if not customer_prosthetics.empty and not sessions_df.empty:
                customer_sessions = pd.merge(
                    customer_prosthetics,
                    sessions_df,
                    on='prosthetic_id',
                    how='inner'
                )
            else:
                customer_sessions = pd.DataFrame()
            
            # 3. Объединяем с данными телеметрии
            if not customer_sessions.empty and not telemetry_df.empty:
                final_data = pd.merge(
                    customer_sessions,
                    telemetry_df,
                    on=['user_id', 'session_id'],
                    how='inner'
                )
            else:
                final_data = pd.DataFrame()
            
            # Агрегируем данные по пользователям
            if not final_data.empty:
                aggregated_data = final_data.groupby('user_id').agg({
                    'session_id': 'count',
                    'accuracy': 'mean',
                    'battery_level': 'mean',
                    'temperature': 'mean',
                    'timestamp': ['min', 'max'],
                    'movement_type': 'count'
                }).reset_index()
                
                # Переименовываем колонки
                aggregated_data.columns = [
                    'user_id', 'total_sessions', 'avg_accuracy', 
                    'avg_battery_level', 'avg_temperature',
                    'first_activity', 'last_activity', 'total_movements'
                ]
            else:
                # Создаем пустой DataFrame с правильными колонками
                aggregated_data = pd.DataFrame(columns=[
                    'user_id', 'total_sessions', 'avg_accuracy', 
                    'avg_battery_level', 'avg_temperature',
                    'first_activity', 'last_activity', 'total_movements'
                ])
            
            logger.info(f"Трансформировано {len(aggregated_data)} записей")
            return aggregated_data
            
        except Exception as e:
            logger.error(f"Ошибка при трансформации данных: {str(e)}")
            return pd.DataFrame()
    
    def load_data_warehouse(self, transformed_data):
        """
        Загрузка данных в витрину данных
        """
        logger.info("Начинаем загрузку данных в витрину...")
        
        try:
            # Подключение к PostgreSQL
            conn = psycopg2.connect(**self.db_config)
            cursor = conn.cursor()
            
            # Очищаем старые данные за текущую дату
            delete_query = """
            DELETE FROM user_analytics_warehouse 
            WHERE date_created = %s
            """
            cursor.execute(delete_query, (date.today(),))
            
            # Загружаем новые данные
            if not transformed_data.empty:
                for _, record in transformed_data.iterrows():
                    insert_query = """
                    INSERT INTO user_analytics_warehouse (
                        user_id, date_created, total_sessions, avg_accuracy,
                        avg_battery_level, avg_temperature, first_activity,
                        last_activity, total_movements, created_at
                    ) VALUES (
                        %s, %s, %s, %s, %s, %s, %s, %s, %s, NOW()
                    )
                    """
                    
                    cursor.execute(insert_query, (
                        record['user_id'],
                        date.today(),
                        record['total_sessions'],
                        record['avg_accuracy'],
                        record['avg_battery_level'],
                        record['avg_temperature'],
                        record['first_activity'],
                        record['last_activity'],
                        record['total_movements']
                    ))
                
                logger.info(f"Загружено {len(transformed_data)} записей в витрину")
            else:
                logger.warning("Нет данных для загрузки")
            
            conn.commit()
            cursor.close()
            conn.close()
            
            return True
            
        except Exception as e:
            logger.error(f"Ошибка при загрузке данных в витрину: {str(e)}")
            return False
    
    def validate_data_quality(self):
        """
        Валидация качества данных
        """
        logger.info("Начинаем валидацию качества данных...")
        
        try:
            conn = psycopg2.connect(**self.db_config)
            cursor = conn.cursor()
            
            # Проверяем количество записей
            count_query = """
            SELECT COUNT(*) as record_count 
            FROM user_analytics_warehouse 
            WHERE date_created = %s
            """
            
            cursor.execute(count_query, (date.today(),))
            result = cursor.fetchone()
            record_count = result[0] if result else 0
            
            if record_count == 0:
                logger.warning("Нет данных для загруженной даты")
                return False
            
            # Проверяем качество данных
            quality_query = """
            SELECT 
                COUNT(*) as total_records,
                COUNT(CASE WHEN avg_accuracy IS NULL THEN 1 END) as null_accuracy,
                COUNT(CASE WHEN avg_battery_level IS NULL THEN 1 END) as null_battery,
                COUNT(CASE WHEN total_sessions = 0 THEN 1 END) as zero_sessions
            FROM user_analytics_warehouse 
            WHERE date_created = %s
            """
            
            cursor.execute(quality_query, (date.today(),))
            quality_result = cursor.fetchone()
            
            if quality_result:
                total_records, null_accuracy, null_battery, zero_sessions = quality_result
                
                logger.info(f"Валидация данных:")
                logger.info(f"  Всего записей: {total_records}")
                logger.info(f"  Записей с NULL accuracy: {null_accuracy}")
                logger.info(f"  Записей с NULL battery: {null_battery}")
                logger.info(f"  Записей с 0 сессий: {zero_sessions}")
                
                # Проверяем качество
                if null_accuracy > total_records * 0.1:  # Более 10% NULL
                    logger.error(f"Слишком много NULL значений в accuracy: {null_accuracy}")
                    return False
                
                if zero_sessions > total_records * 0.05:  # Более 5% с 0 сессий
                    logger.error(f"Слишком много записей с 0 сессий: {zero_sessions}")
                    return False
            
            logger.info("Валидация данных завершена успешно")
            return True
            
        except Exception as e:
            logger.error(f"Ошибка при валидации данных: {str(e)}")
            return False
        finally:
            if 'cursor' in locals():
                cursor.close()
            if 'conn' in locals():
                conn.close()
    
    def run_etl(self):
        """
        Запуск полного ETL-процесса
        """
        logger.info("=== ЗАПУСК ETL-ПРОЦЕССА BIONICPRO ===")
        
        try:
            # 1. Извлечение данных
            crm_data = self.extract_crm_data()
            telemetry_data = self.extract_telemetry_data()
            
            # 2. Трансформация данных
            transformed_data = self.transform_data(crm_data, telemetry_data)
            
            # 3. Загрузка данных
            load_success = self.load_data_warehouse(transformed_data)
            
            if load_success:
                # 4. Валидация данных
                validation_success = self.validate_data_quality()
                
                if validation_success:
                    logger.info("=== ETL-ПРОЦЕСС ЗАВЕРШЁН УСПЕШНО ===")
                    return True
                else:
                    logger.error("=== ETL-ПРОЦЕСС ЗАВЕРШЁН С ОШИБКАМИ ВАЛИДАЦИИ ===")
                    return False
            else:
                logger.error("=== ETL-ПРОЦЕСС ЗАВЕРШЁН С ОШИБКАМИ ЗАГРУЗКИ ===")
                return False
                
        except Exception as e:
            logger.error(f"=== ETL-ПРОЦЕСС ЗАВЕРШЁН С ОШИБКОЙ: {str(e)} ===")
            return False

def main():
    """
    Главная функция
    """
    etl = BionicProETL()
    success = etl.run_etl()
    
    if success:
        print("✅ ETL-процесс выполнен успешно!")
        exit(0)
    else:
        print("❌ ETL-процесс завершился с ошибками!")
        exit(1)

if __name__ == "__main__":
    main()
