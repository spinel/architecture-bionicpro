#!/usr/bin/env python3
"""
Тест для проверки JWT токена
"""

import jwt
import time
from datetime import datetime, timedelta

# Секретный ключ
SECRET_KEY = "your-secret-key"

# Создаём простой токен
payload = {
    "sub": "user_001",
    "iss": "bionicpro-test",
    "aud": "bionicpro-api",
    "exp": int((datetime.utcnow() + timedelta(hours=24)).timestamp()),
    "iat": int(time.time()),
    "preferred_username": "user_001",
    "email": "user001@example.com",
    "realm_access": {
        "roles": ["prothetic_user"]
    }
}

# Создаём токен
token = jwt.encode(payload, SECRET_KEY, algorithm="HS256")
print(f"Токен: {token}")

# Проверяем токен
try:
    decoded = jwt.decode(token, SECRET_KEY, algorithms=["HS256"])
    print(f"Декодированный: {decoded}")
except Exception as e:
    print(f"Ошибка декодирования: {e}")

print(f"\nТестовый запрос:")
print(f'curl -H "Authorization: Bearer {token}" "http://localhost:8000/api/v1/analytics/reports?user_id=user_001&start_date=2025-10-17&end_date=2025-10-18"')
