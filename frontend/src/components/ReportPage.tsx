import React, { useState, useEffect } from 'react';
import { useKeycloak } from '@react-keycloak/web';

// Типы данных из backend API
interface ReportMetrics {
  total_usage_hours: number;
  average_daily_hours: number;
  battery_changes: number;
  maintenance_alerts: number;
  movement_activities: number;
  error_count: number;
  comfort_score: number;
  efficiency_score: number;
}

interface Report {
  report_id: string;
  user_id: string;
  username: string;
  device_id: string;
  device_type: string;
  period_start: string;
  period_end: string;
  metrics: ReportMetrics;
  created_at: string;
}

interface ReportResponse {
  reports: Report[];
  total_count: number;
  page: number;
  page_size: number;
}

const ReportPage: React.FC = () => {
  const { keycloak, initialized } = useKeycloak();
  const [reports, setReports] = useState<Report[]>([]);
  const [selectedReport, setSelectedReport] = useState<Report | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [totalCount, setTotalCount] = useState(0);
  const [page, setPage] = useState(1);
  const pageSize = 5;

  // API URL из переменных окружения
  const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8000';

  // Загрузка списка отчётов при монтировании компонента
  useEffect(() => {
    if (keycloak.authenticated) {
      fetchReports();
    }
  }, [keycloak.authenticated, page]);

  /**
   * Получение списка отчётов из backend API
   */
  const fetchReports = async () => {
    if (!keycloak?.token) {
      setError('Not authenticated');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const offset = (page - 1) * pageSize;
      const response = await fetch(
        `${API_URL}/api/reports?limit=${pageSize}&offset=${offset}`,
        {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${keycloak.token}`,
            'Content-Type': 'application/json'
          }
        }
      );

      if (!response.ok) {
        if (response.status === 401) {
          throw new Error('Unauthorized. Please login again.');
        } else if (response.status === 403) {
          throw new Error('Access denied.');
        } else {
          throw new Error(`Failed to fetch reports: ${response.statusText}`);
        }
      }

      const data: ReportResponse = await response.json();
      setReports(data.reports || []);
      setTotalCount(data.total_count || 0);
    } catch (err) {
      console.error('[Reports] Error fetching reports:', err);
      setError(err instanceof Error ? err.message : 'Failed to load reports');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Получение конкретного отчёта по ID
   */
  const fetchReportById = async (reportId: string) => {
    if (!keycloak?.token) {
      setError('Not authenticated');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await fetch(
        `${API_URL}/api/reports/${reportId}`,
        {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${keycloak.token}`,
            'Content-Type': 'application/json'
          }
        }
      );

      if (!response.ok) {
        if (response.status === 403) {
          throw new Error('Access denied. This report belongs to another user.');
        } else if (response.status === 404) {
          throw new Error('Report not found.');
        } else {
          throw new Error(`Failed to fetch report: ${response.statusText}`);
        }
      }

      const data: Report = await response.json();
      setSelectedReport(data);
    } catch (err) {
      console.error('[Reports] Error fetching report details:', err);
      setError(err instanceof Error ? err.message : 'Failed to load report details');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Форматирование даты
   */
  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  /**
   * Форматирование числа с разделителями
   */
  const formatNumber = (num: number): string => {
    return num.toLocaleString('ru-RU', { maximumFractionDigits: 1 });
  };

  // Загрузка приложения
  if (!initialized) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-100">
        <div className="text-xl text-gray-600">Загрузка...</div>
      </div>
    );
  }

  // Экран входа
  if (!keycloak.authenticated) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
        <div className="p-12 bg-white rounded-2xl shadow-xl text-center">
          <h1 className="text-3xl font-bold text-gray-800 mb-4">
            BionicPRO Reports
          </h1>
          <p className="text-gray-600 mb-8">
            Отчёты об использовании бионических протезов
          </p>
          <button
            onClick={() => keycloak.login()}
            className="px-8 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium"
          >
            Войти
          </button>
        </div>
      </div>
    );
  }

  // Главный экран с отчётами
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-8 px-4">
      <div className="max-w-6xl mx-auto">
        {/* Шапка */}
        <div className="bg-white rounded-2xl shadow-lg p-6 mb-6">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-bold text-gray-800">
                Мои отчёты
              </h1>
              <p className="text-gray-600 mt-1">
                Пользователь: <span className="font-medium">{keycloak.tokenParsed?.preferred_username}</span>
              </p>
            </div>
            <button
              onClick={() => keycloak.logout()}
              className="px-4 py-2 text-gray-600 hover:text-gray-800 transition-colors"
            >
              Выйти
            </button>
          </div>
        </div>

        {/* Ошибка */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
            <p className="text-red-800">{error}</p>
          </div>
        )}

        {/* Кнопка обновления */}
        <div className="mb-6">
          <button
            onClick={fetchReports}
            disabled={loading}
            className={`px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium ${
              loading ? 'opacity-50 cursor-not-allowed' : ''
            }`}
          >
            {loading ? '⏳ Загрузка...' : '🔄 Обновить отчёты'}
          </button>
          <span className="ml-4 text-gray-600">
            Всего отчётов: <span className="font-medium">{totalCount}</span>
          </span>
        </div>

        {/* Список отчётов */}
        {!loading && reports.length === 0 ? (
          <div className="bg-white rounded-2xl shadow-lg p-12 text-center">
            <p className="text-gray-600 text-lg">
              У вас пока нет отчётов
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            {reports.map((report) => (
              <div
                key={report.report_id}
                className="bg-white rounded-xl shadow-md hover:shadow-lg transition-shadow p-6"
              >
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <h3 className="text-xl font-bold text-gray-800 mb-2">
                      {report.device_type}
                    </h3>
                    <p className="text-gray-600 mb-1">
                      <span className="font-medium">Период:</span>{' '}
                      {formatDate(report.period_start)} - {formatDate(report.period_end)}
                    </p>
                    <p className="text-gray-600 mb-4">
                      <span className="font-medium">ID устройства:</span> {report.device_id}
                    </p>

                    {/* Краткая статистика */}
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                      <div className="bg-blue-50 rounded-lg p-3">
                        <p className="text-xs text-gray-600">Часов использования</p>
                        <p className="text-lg font-bold text-blue-600">
                          {formatNumber(report.metrics.total_usage_hours)}
                        </p>
                      </div>
                      <div className="bg-green-50 rounded-lg p-3">
                        <p className="text-xs text-gray-600">Комфорт</p>
                        <p className="text-lg font-bold text-green-600">
                          {formatNumber(report.metrics.comfort_score)} / 10
                        </p>
                      </div>
                      <div className="bg-purple-50 rounded-lg p-3">
                        <p className="text-xs text-gray-600">Эффективность</p>
                        <p className="text-lg font-bold text-purple-600">
                          {formatNumber(report.metrics.efficiency_score)} / 10
                        </p>
                      </div>
                      <div className="bg-orange-50 rounded-lg p-3">
                        <p className="text-xs text-gray-600">Активность</p>
                        <p className="text-lg font-bold text-orange-600">
                          {formatNumber(report.metrics.movement_activities)}
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Кнопка просмотра */}
                  <button
                    onClick={() => fetchReportById(report.report_id)}
                    className="ml-4 px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-medium whitespace-nowrap"
                  >
                    📊 Подробнее
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Пагинация */}
        {totalCount > pageSize && (
          <div className="mt-6 flex justify-center items-center space-x-4">
            <button
              onClick={() => setPage(Math.max(1, page - 1))}
              disabled={page === 1}
              className={`px-4 py-2 bg-white rounded-lg shadow ${
                page === 1
                  ? 'text-gray-400 cursor-not-allowed'
                  : 'text-gray-700 hover:bg-gray-50'
              }`}
            >
              ← Назад
            </button>
            <span className="text-gray-700">
              Страница {page} из {Math.ceil(totalCount / pageSize)}
            </span>
            <button
              onClick={() => setPage(page + 1)}
              disabled={page >= Math.ceil(totalCount / pageSize)}
              className={`px-4 py-2 bg-white rounded-lg shadow ${
                page >= Math.ceil(totalCount / pageSize)
                  ? 'text-gray-400 cursor-not-allowed'
                  : 'text-gray-700 hover:bg-gray-50'
              }`}
            >
              Вперёд →
            </button>
          </div>
        )}

        {/* Модальное окно с деталями отчёта */}
        {selectedReport && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
            <div className="bg-white rounded-2xl shadow-2xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
              <div className="p-8">
                <div className="flex justify-between items-start mb-6">
                  <h2 className="text-2xl font-bold text-gray-800">
                    Детальный отчёт
                  </h2>
                  <button
                    onClick={() => setSelectedReport(null)}
                    className="text-gray-400 hover:text-gray-600 text-2xl"
                  >
                    ×
                  </button>
                </div>

                {/* Информация об устройстве */}
                <div className="bg-gray-50 rounded-lg p-6 mb-6">
                  <h3 className="text-lg font-bold text-gray-800 mb-4">
                    Информация об устройстве
                  </h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <p className="text-sm text-gray-600">Тип устройства</p>
                      <p className="font-medium text-gray-800">{selectedReport.device_type}</p>
                    </div>
                    <div>
                      <p className="text-sm text-gray-600">ID устройства</p>
                      <p className="font-medium text-gray-800">{selectedReport.device_id}</p>
                    </div>
                    <div>
                      <p className="text-sm text-gray-600">Начало периода</p>
                      <p className="font-medium text-gray-800">{formatDate(selectedReport.period_start)}</p>
                    </div>
                    <div>
                      <p className="text-sm text-gray-600">Конец периода</p>
                      <p className="font-medium text-gray-800">{formatDate(selectedReport.period_end)}</p>
                    </div>
                  </div>
                </div>

                {/* Детальные метрики */}
                <div className="space-y-4">
                  <h3 className="text-lg font-bold text-gray-800 mb-4">
                    Метрики использования
                  </h3>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="bg-blue-50 rounded-lg p-4">
                      <p className="text-sm text-gray-600 mb-1">Общее время использования</p>
                      <p className="text-2xl font-bold text-blue-600">
                        {formatNumber(selectedReport.metrics.total_usage_hours)} часов
                      </p>
                    </div>

                    <div className="bg-indigo-50 rounded-lg p-4">
                      <p className="text-sm text-gray-600 mb-1">Среднее время в день</p>
                      <p className="text-2xl font-bold text-indigo-600">
                        {formatNumber(selectedReport.metrics.average_daily_hours)} часов
                      </p>
                    </div>

                    <div className="bg-yellow-50 rounded-lg p-4">
                      <p className="text-sm text-gray-600 mb-1">Замены батареи</p>
                      <p className="text-2xl font-bold text-yellow-600">
                        {selectedReport.metrics.battery_changes}
                      </p>
                    </div>

                    <div className="bg-red-50 rounded-lg p-4">
                      <p className="text-sm text-gray-600 mb-1">Предупреждения о обслуживании</p>
                      <p className="text-2xl font-bold text-red-600">
                        {selectedReport.metrics.maintenance_alerts}
                      </p>
                    </div>

                    <div className="bg-purple-50 rounded-lg p-4">
                      <p className="text-sm text-gray-600 mb-1">Активных движений</p>
                      <p className="text-2xl font-bold text-purple-600">
                        {formatNumber(selectedReport.metrics.movement_activities)}
                      </p>
                    </div>

                    <div className="bg-pink-50 rounded-lg p-4">
                      <p className="text-sm text-gray-600 mb-1">Количество ошибок</p>
                      <p className="text-2xl font-bold text-pink-600">
                        {selectedReport.metrics.error_count}
                      </p>
                    </div>

                    <div className="bg-green-50 rounded-lg p-4">
                      <p className="text-sm text-gray-600 mb-1">Оценка комфорта</p>
                      <p className="text-2xl font-bold text-green-600">
                        {formatNumber(selectedReport.metrics.comfort_score)} / 10
                      </p>
                    </div>

                    <div className="bg-teal-50 rounded-lg p-4">
                      <p className="text-sm text-gray-600 mb-1">Оценка эффективности</p>
                      <p className="text-2xl font-bold text-teal-600">
                        {formatNumber(selectedReport.metrics.efficiency_score)} / 10
                      </p>
                    </div>
                  </div>
                </div>

                {/* Закрыть */}
                <button
                  onClick={() => setSelectedReport(null)}
                  className="mt-8 w-full px-6 py-3 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors font-medium"
                >
                  Закрыть
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ReportPage;