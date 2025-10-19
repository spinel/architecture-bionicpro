#!/usr/bin/env python3
"""
Скрипт для создания тестового JWT токена для демонстрации ограничений доступа
"""

import jwt
import time
from datetime import datetime, timedelta

# Секретный ключ (в реальном приложении должен быть в переменных окружения)
SECRET_KEY = "your-secret-key"

def create_test_token(user_id: str, expires_in_hours: int = 24) -> str:
    """Создаёт тестовый JWT токен для указанного пользователя"""
    
    # Время истечения токена
    exp = datetime.utcnow() + timedelta(hours=expires_in_hours)
    
    # Payload токена
    payload = {
        "sub": user_id,  # Subject - ID пользователя
        "iss": "bionicpro-test",  # Issuer
        "aud": "bionicpro-api",  # Audience
        "exp": int(exp.timestamp()),  # Expiration time
        "iat": int(time.time()),  # Issued at
        "preferred_username": f"user_{user_id}",
        "email": f"user{user_id}@example.com",
        "realm_access": {
            "roles": ["prothetic_user"]
        }
    }
    
    # Создаём токен
    token = jwt.encode(payload, SECRET_KEY, algorithm="HS256")
    return token

def main():
    """Создаёт тестовые токены для разных пользователей"""
    
    print("=== ТЕСТОВЫЕ JWT ТОКЕНЫ ===\n")
    
    # Токен для user_001
    token_001 = create_test_token("user_001")
    print(f"Токен для user_001:")
    print(f"Authorization: Bearer {token_001}\n")
    
    # Токен для user_002
    token_002 = create_test_token("user_002")
    print(f"Токен для user_002:")
    print(f"Authorization: Bearer {token_002}\n")
    
    # Токен для user_003
    token_003 = create_test_token("user_003")
    print(f"Токен для user_003:")
    print(f"Authorization: Bearer {token_003}\n")
    
    print("=== ИНСТРУКЦИИ ПО ТЕСТИРОВАНИЮ ===")
    print("1. Используйте токены для тестирования API")
    print("2. Каждый пользователь может получить только свои отчёты")
    print("3. Попытка доступа к чужим данным вернёт ошибку 403")

if __name__ == "__main__":
    main()
