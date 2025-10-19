#!/usr/bin/env python3
"""
CRM API для BionicPRO
====================

Простой Flask API для имитации CRM-системы
"""

from flask import Flask, jsonify
import json

app = Flask(__name__)

@app.route("/api/customers")
def customers():
    """Получение списка клиентов"""
    return jsonify([
        {
            "customer_id": "001",
            "name": "Иван Петров",
            "email": "ivan@example.com",
            "phone": "+7-999-123-45-67",
            "registration_date": "2024-01-15"
        },
        {
            "customer_id": "002",
            "name": "Мария Сидорова",
            "email": "maria@example.com",
            "phone": "+7-999-234-56-78",
            "registration_date": "2024-02-20"
        }
    ])

@app.route("/api/prosthetics")
def prosthetics():
    """Получение списка протезов"""
    return jsonify([
        {
            "prosthetic_id": "P001",
            "customer_id": "001",
            "model": "BionicHand Pro",
            "serial_number": "BH-2024-001",
            "purchase_date": "2024-01-20"
        },
        {
            "prosthetic_id": "P002",
            "customer_id": "002",
            "model": "BionicLeg Advanced",
            "serial_number": "BL-2024-002",
            "purchase_date": "2024-02-25"
        }
    ])

@app.route("/api/sessions")
def sessions():
    """Получение списка сессий"""
    return jsonify([
        {
            "session_id": "S001",
            "prosthetic_id": "P001",
            "user_id": "user_001",
            "start_time": "2024-12-17T08:00:00Z",
            "end_time": "2024-12-17T18:00:00Z",
            "duration_minutes": 600
        },
        {
            "session_id": "S002",
            "prosthetic_id": "P002",
            "user_id": "user_002",
            "start_time": "2024-12-17T09:00:00Z",
            "end_time": "2024-12-17T17:00:00Z",
            "duration_minutes": 480
        }
    ])

@app.route("/health")
def health():
    """Проверка здоровья сервиса"""
    return jsonify({"status": "healthy", "service": "CRM API"})

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8001, debug=True)
