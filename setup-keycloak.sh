#!/bin/bash

echo "=== НАСТРОЙКА KEYCLOAK ДЛЯ HTTP ==="

# Ждем запуска Keycloak
echo "Ожидание запуска Keycloak..."
sleep 15

# Получаем токен администратора
echo "Получение токена администратора..."
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/realms/master/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin" \
  -d "password=admin" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" | jq -r '.access_token')

if [ "$ADMIN_TOKEN" = "null" ] || [ -z "$ADMIN_TOKEN" ]; then
  echo "❌ Не удалось получить токен администратора"
  exit 1
fi

echo "✅ Токен получен"

# Обновляем настройки клиента reports-frontend
echo "Обновление настроек клиента reports-frontend..."

curl -s -X PUT "http://localhost:8080/admin/realms/reports-realm/clients/$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "http://localhost:8080/admin/realms/reports-realm/clients?clientId=reports-frontend" | jq -r '.[0].id')" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "reports-frontend",
    "name": "Reports Frontend",
    "description": "Frontend application for BionicPRO reports",
    "enabled": true,
    "clientAuthenticatorType": "client-secret",
    "secret": "",
    "redirectUris": ["http://localhost:3000/*"],
    "webOrigins": ["http://localhost:3000"],
    "standardFlowEnabled": true,
    "implicitFlowEnabled": false,
    "directAccessGrantsEnabled": false,
    "serviceAccountsEnabled": false,
    "publicClient": true,
    "protocol": "openid-connect",
    "attributes": {
      "pkce.code.challenge.method": "S256"
    }
  }'

echo ""
echo "✅ Настройки клиента обновлены"

# Создаем тестового пользователя
echo "Создание тестового пользователя prothetic1..."

curl -s -X POST "http://localhost:8080/admin/realms/reports-realm/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "prothetic1",
    "email": "prothetic1@bionicpro.com",
    "firstName": "Иван",
    "lastName": "Петров",
    "enabled": true,
    "credentials": [{
      "type": "password",
      "value": "password123",
      "temporary": false
    }]
  }'

echo ""
echo "✅ Пользователь prothetic1 создан"

echo ""
echo "🎯 НАСТРОЙКА ЗАВЕРШЕНА!"
echo "Теперь можно тестировать аутентификацию на http://localhost:3000"
