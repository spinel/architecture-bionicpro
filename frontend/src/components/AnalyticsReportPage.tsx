import React, { useState, useEffect, useCallback } from 'react';
import { apiService, UserReport, DailySummary } from '../services/apiService';

const AnalyticsReportPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [reports, setReports] = useState<UserReport[]>([]);
  const [summary, setSummary] = useState<DailySummary | null>(null);
  const [dateRange, setDateRange] = useState({
    startDate: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0], // 7 дней назад
    endDate: new Date().toISOString().split('T')[0] // сегодня
  });
  const [selectedUserId, setSelectedUserId] = useState('user_001'); // Для демонстрации

  // Загрузка отчётов
  const loadReports = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const data = await apiService.getUserReports(
        selectedUserId,
        dateRange.startDate,
        dateRange.endDate,
        true // Используем демо-режим
      );
      
      if (data.success) {
        setReports(data.data);
        setSummary(data.summary || null);
      } else {
        throw new Error(data.error || 'Неизвестная ошибка');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Произошла ошибка');
      setReports([]);
      setSummary(null);
    } finally {
      setLoading(false);
    }
  }, [selectedUserId, dateRange.startDate, dateRange.endDate]);

  // Загрузка сводки
  const loadSummary = async () => {
    try {
      const data = await apiService.getDailySummary(dateRange.startDate, dateRange.endDate);
      if (data.length > 0) {
        setSummary(data[0]); // Берём первую запись
      }
    } catch (err) {
      console.warn('Не удалось загрузить сводку:', err);
    }
  };

  // Загрузка данных при изменении параметров
  useEffect(() => {
    loadReports();
  }, [loadReports]);

  // Скачивание отчёта в виде файла
  const downloadReport = async () => {
    try {
      setLoading(true);
      setError(null);

      const reportContent = generateReportContent();
      const blob = new Blob([reportContent], { type: 'text/plain;charset=utf-8' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `bionicpro-report-${selectedUserId}-${dateRange.startDate}-${dateRange.endDate}.txt`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка генерации отчёта');
    } finally {
      setLoading(false);
    }
  };

  // Генерация содержимого отчёта
  const generateReportContent = (): string => {
    const now = new Date();
    let content = `ОТЧЁТ О РАБОТЕ ПРОТЕЗА BIONICPRO
=====================================

Дата создания: ${now.toLocaleDateString("ru-RU")} ${now.toLocaleTimeString("ru-RU")}
Пользователь: ${selectedUserId}
Период: ${dateRange.startDate} - ${dateRange.endDate}

`;

    if (reports.length === 0) {
      content += "Данные за указанный период не найдены.\n";
      return content;
    }

    // Общая статистика
    const totalSessions = reports.reduce((sum, r) => sum + r.total_sessions, 0);
    const totalMovements = reports.reduce((sum, r) => sum + r.total_movements, 0);
    const avgAccuracy = reports.reduce((sum, r) => sum + r.avg_accuracy, 0) / reports.length;
    const avgBattery = reports.reduce((sum, r) => sum + r.avg_battery_level, 0) / reports.length;

    content += `ОБЩАЯ СТАТИСТИКА
----------------
- Общее количество сессий: ${totalSessions}
- Общее количество движений: ${totalMovements}
- Средняя точность: ${avgAccuracy.toFixed(1)}%
- Средний уровень батареи: ${avgBattery.toFixed(1)}%
- Количество дней с данными: ${reports.length}

`;

    // Детальная статистика по дням
    content += `ДЕТАЛЬНАЯ СТАТИСТИКА ПО ДНЯМ
===============================

`;
    reports.forEach((report, index) => {
      const date = new Date(report.date_created).toLocaleDateString("ru-RU");
      content += `${index + 1}. ${date}
   Сессий: ${report.total_sessions}
   Движений: ${report.total_movements}
   Точность: ${report.avg_accuracy.toFixed(1)}% (${report.accuracy_status})
   Батарея: ${report.avg_battery_level.toFixed(1)}% (${report.battery_status})
   Температура: ${report.avg_temperature.toFixed(1)}°C
   Длительность: ${report.session_duration_hours.toFixed(1)} ч
   Активность: ${new Date(report.first_activity).toLocaleTimeString("ru-RU")} - ${new Date(report.last_activity).toLocaleTimeString("ru-RU")}

`;
    });

    // Рекомендации
    content += `РЕКОМЕНДАЦИИ
============
`;
    if (avgAccuracy < 85) {
      content += "- Рекомендуется провести калибровку датчиков (низкая точность)\n";
    }
    if (avgBattery < 40) {
      content += "- Обратите внимание на уровень заряда батареи\n";
    }
    if (totalSessions < 5) {
      content += "- Рекомендуется увеличить активность использования протеза\n";
    }
    if (avgAccuracy >= 95 && avgBattery >= 80) {
      content += "- Отличные показатели! Продолжайте в том же духе\n";
    }

    content += `
---
Сгенерировано системой BionicPRO
Дата: ${now.toLocaleString("ru-RU")}
`;

    return content;
  };

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">BionicPRO Analytics</h1>
              <p className="text-sm text-gray-600">Аналитические отчёты о работе протеза</p>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          
          {/* Фильтры */}
          <div className="bg-white shadow rounded-lg p-6 mb-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Параметры отчёта</h2>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Пользователь
                </label>
                <select
                  value={selectedUserId}
                  onChange={(e) => setSelectedUserId(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="user_001">user_001</option>
                  <option value="user_002">user_002</option>
                  <option value="user_003">user_003</option>
                </select>
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Начальная дата
                </label>
                <input
                  type="date"
                  value={dateRange.startDate}
                  onChange={(e) => setDateRange(prev => ({ ...prev, startDate: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Конечная дата
                </label>
                <input
                  type="date"
                  value={dateRange.endDate}
                  onChange={(e) => setDateRange(prev => ({ ...prev, endDate: e.target.value }))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>

            <div className="flex space-x-4">
              <button
                onClick={loadReports}
                disabled={loading}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
              >
                {loading ? 'Загрузка...' : 'Обновить данные'}
              </button>
              
              <button
                onClick={downloadReport}
                disabled={loading || reports.length === 0}
                className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 disabled:opacity-50"
              >
                Скачать отчёт
              </button>
            </div>
          </div>

          {/* Ошибки */}
          {error && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-md">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-red-800">Ошибка</h3>
                  <div className="mt-2 text-sm text-red-700">{error}</div>
                </div>
              </div>
            </div>
          )}

          {/* Сводка */}
          {summary && (
            <div className="bg-white shadow rounded-lg p-6 mb-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Общая сводка</h3>
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div className="bg-blue-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-blue-600">{summary.active_users}</div>
                  <div className="text-sm text-blue-800">Активных пользователей</div>
                </div>
                <div className="bg-green-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-green-600">{summary.avg_sessions_per_user.toFixed(1)}</div>
                  <div className="text-sm text-green-800">Среднее сессий на пользователя</div>
                </div>
                <div className="bg-yellow-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-yellow-600">{summary.avg_accuracy_overall.toFixed(1)}%</div>
                  <div className="text-sm text-yellow-800">Средняя точность</div>
                </div>
                <div className="bg-purple-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-purple-600">{summary.total_movements_all_users}</div>
                  <div className="text-sm text-purple-800">Общее количество движений</div>
                </div>
              </div>
            </div>
          )}

          {/* Таблица отчётов */}
          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Детальные отчёты</h3>
            </div>
            
            {loading ? (
              <div className="p-6 text-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
                <p className="text-gray-600">Загрузка данных...</p>
              </div>
            ) : reports.length === 0 ? (
              <div className="p-6 text-center text-gray-500">
                Нет данных за выбранный период
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Дата
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Сессии
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Движения
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Точность
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Батарея
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Длительность
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {reports.map((report, index) => (
                      <tr key={index} className="hover:bg-gray-50">
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {new Date(report.date_created).toLocaleDateString("ru-RU")}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {report.total_sessions}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {report.total_movements}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            report.avg_accuracy >= 95 ? 'bg-green-100 text-green-800' :
                            report.avg_accuracy >= 85 ? 'bg-yellow-100 text-yellow-800' :
                            'bg-red-100 text-red-800'
                          }`}>
                            {report.avg_accuracy.toFixed(1)}%
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            report.avg_battery_level >= 80 ? 'bg-green-100 text-green-800' :
                            report.avg_battery_level >= 40 ? 'bg-yellow-100 text-yellow-800' :
                            'bg-red-100 text-red-800'
                          }`}>
                            {report.avg_battery_level.toFixed(1)}%
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {report.session_duration_hours.toFixed(1)} ч
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
};

export default AnalyticsReportPage;
