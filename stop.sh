#!/bin/bash

# Скрипт для остановки BionicPRO Reports

echo "🛑 Stopping BionicPRO Reports..."

docker-compose down

echo ""
echo "✅ All services stopped"
echo ""
echo "To remove volumes and data:"
echo "  docker-compose down -v"
echo ""
echo "To clean ClickHouse data:"
echo "  rm -rf clickhouse-data"

