#!/bin/bash

# Скрипт для быстрого запуска BionicPRO Reports

set -e

echo "🚀 Starting BionicPRO Reports..."
echo ""

# Проверка Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker не установлен. Установите Docker Desktop."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose не установлен."
    exit 1
fi

# Остановка существующих контейнеров
echo "📦 Stopping existing containers..."
docker-compose down 2>/dev/null || true

# Очистка старых данных ClickHouse (опционально)
if [ "$1" == "--clean" ]; then
    echo "🧹 Cleaning ClickHouse data..."
    rm -rf clickhouse-data
fi

# Запуск сервисов
echo "🔧 Building and starting services..."
docker-compose up -d --build

echo ""
echo "⏳ Waiting for services to be ready..."
sleep 5

# Проверка статуса сервисов
echo ""
echo "📊 Service Status:"
docker-compose ps

# Проверка health checks
echo ""
echo "🏥 Health Checks:"

# Backend health check
for i in {1..30}; do
    if curl -s http://localhost:8000/api/health > /dev/null 2>&1; then
        echo "✅ Backend API is healthy"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "⚠️  Backend API is not responding (timeout)"
    fi
    sleep 1
done

# Frontend check
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    echo "✅ Frontend is accessible"
else
    echo "⚠️  Frontend is not responding yet (may still be starting)"
fi

# Keycloak check
if curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo "✅ Keycloak is accessible"
else
    echo "⚠️  Keycloak is not responding yet (may still be starting)"
fi

# ClickHouse check
if curl -s http://localhost:8123/ping > /dev/null 2>&1; then
    echo "✅ ClickHouse is healthy"
else
    echo "⚠️  ClickHouse is not responding yet"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 BionicPRO Reports is starting up!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📱 Access the application:"
echo "   • Frontend:      http://localhost:3000"
echo "   • Backend API:   http://localhost:8000/api/health"
echo "   • Keycloak:      http://localhost:8080"
echo "   • ClickHouse:    http://localhost:8123"
echo ""
echo "👤 Test users:"
echo "   • prothetic1 / prothetic123 (5 reports)"
echo "   • prothetic2 / prothetic123 (3 reports)"
echo "   • prothetic3 / prothetic123 (4 reports)"
echo ""
echo "📋 View logs:"
echo "   docker-compose logs -f [service]"
echo ""
echo "🛑 Stop services:"
echo "   docker-compose down"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

