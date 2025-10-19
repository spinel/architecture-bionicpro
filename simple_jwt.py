#!/usr/bin/env python3
"""
Простой JWT токен для тестирования
"""

import jwt
import time

# Секретный ключ
SECRET_KEY = "your-secret-key"

# Простой payload
payload = {
    "sub": "user_001",
    "email": "user001@example.com",
    "exp": int(time.time()) + 3600,  # 1 час
    "iat": int(time.time())
}

# Создаём токен
token = jwt.encode(payload, SECRET_KEY, algorithm="HS256")
print(f"Простой токен: {token}")

# Проверяем токен
try:
    decoded = jwt.decode(token, SECRET_KEY, algorithms=["HS256"])
    print(f"Декодированный: {decoded}")
    print("✅ Токен валиден!")
except Exception as e:
    print(f"❌ Ошибка декодирования: {e}")

print(f"\nТестовый запрос:")
print(f'curl -H "Authorization: Bearer {token}" "http://localhost:8000/api/v1/analytics/reports?user_id=user_001&start_date=2025-10-17&end_date=2025-10-18"')
