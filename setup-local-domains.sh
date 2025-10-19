#!/bin/bash

echo "=== НАСТРОЙКА ЛОКАЛЬНЫХ ДОМЕНОВ ==="

# Проверяем, есть ли уже записи
if grep -q "bionicpro.local" /etc/hosts; then
    echo "✅ Домены уже добавлены в /etc/hosts"
else
    echo "Добавляем домены в /etc/hosts..."
    echo "" >> /etc/hosts
    echo "# BionicPRO Local Development Domains" >> /etc/hosts
    echo "127.0.0.1 bionicpro.local" >> /etc/hosts
    echo "127.0.0.1 keycloak.bionicpro.local" >> /etc/hosts
    echo "127.0.0.1 api.bionicpro.local" >> /etc/hosts
    echo "✅ Домены добавлены в /etc/hosts"
fi

echo ""
echo "🌐 Локальные домены настроены:"
echo "• https://bionicpro.local - Frontend"
echo "• https://keycloak.bionicpro.local - Keycloak"
echo "• https://api.bionicpro.local - Backend API"
echo ""
echo "⚠️  Для работы с HTTPS нужно:"
echo "1. Добавить сертификаты в доверенные (macOS):"
echo "   sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ssl_certs_local/bionicpro.local.crt"
echo ""
echo "2. Или использовать Chrome с флагом --ignore-certificate-errors"
echo ""
echo "🚀 Готово к запуску!"
