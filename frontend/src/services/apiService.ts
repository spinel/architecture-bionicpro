// Сервис для работы с API аналитики
export interface UserReport {
  user_id: string;
  date_created: string;
  total_sessions: number;
  avg_accuracy: number;
  avg_battery_level: number;
  avg_temperature: number;
  first_activity: string;
  last_activity: string;
  total_movements: number;
  accuracy_status: string;
  battery_status: string;
  session_duration_hours: number;
}

export interface DailySummary {
  date_created: string;
  active_users: number;
  avg_sessions_per_user: number;
  avg_accuracy_overall: number;
  avg_battery_overall: number;
  total_movements_all_users: number;
}

export interface ReportResponse {
  success: boolean;
  data: UserReport[];
  summary?: DailySummary;
  meta: {
    generated_at: string;
    user_id?: string;
    date_range?: string;
    record_count: number;
  };
  error?: string;
}

export interface ErrorResponse {
  success: false;
  error: string;
  code?: number;
}

class ApiService {
  private baseUrl: string;

  constructor() {
    this.baseUrl = process.env.REACT_APP_API_URL || 'http://localhost:8000';
  }

  // Получение отчётов пользователя
  async getUserReports(
    userId: string, 
    startDate: string, 
    endDate: string,
    useDemo: boolean = true
  ): Promise<ReportResponse> {
    const endpoint = useDemo ? '/api/v1/demo/reports' : '/api/v1/analytics/reports';
    
    const params = new URLSearchParams({
      user_id: userId,
      start_date: startDate,
      end_date: endDate
    });

    const headers: Record<string, string> = {
      'Content-Type': 'application/json'
    };

    // Для демо-режима добавляем заголовок X-User-ID
    if (useDemo) {
      headers['X-User-ID'] = userId;
    }

    const response = await fetch(`${this.baseUrl}${endpoint}?${params}`, {
      method: 'GET',
      headers
    });

    if (!response.ok) {
      const errorData: ErrorResponse = await response.json();
      throw new Error(errorData.error || `Ошибка сервера: ${response.status}`);
    }

    return await response.json();
  }

  // Получение отчёта пользователя по ID
  async getUserReportById(
    userId: string, 
    days: number = 7,
    useDemo: boolean = true
  ): Promise<ReportResponse> {
    const endpoint = useDemo ? `/api/v1/demo/user/${userId}` : `/api/v1/analytics/user/${userId}`;
    
    const params = new URLSearchParams({
      days: days.toString()
    });

    const headers: Record<string, string> = {
      'Content-Type': 'application/json'
    };

    // Для демо-режима добавляем заголовок X-User-ID
    if (useDemo) {
      headers['X-User-ID'] = userId;
    }

    const response = await fetch(`${this.baseUrl}${endpoint}?${params}`, {
      method: 'GET',
      headers
    });

    if (!response.ok) {
      const errorData: ErrorResponse = await response.json();
      throw new Error(errorData.error || `Ошибка сервера: ${response.status}`);
    }

    return await response.json();
  }

  // Получение ежедневной сводки
  async getDailySummary(
    startDate: string, 
    endDate: string
  ): Promise<DailySummary[]> {
    const params = new URLSearchParams({
      start_date: startDate,
      end_date: endDate
    });

    const response = await fetch(`${this.baseUrl}/api/v1/analytics/summary?${params}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      const errorData: ErrorResponse = await response.json();
      throw new Error(errorData.error || `Ошибка сервера: ${response.status}`);
    }

    return await response.json();
  }

  // Проверка здоровья API
  async checkHealth(): Promise<{ status: string; timestamp: string; version: string; database: string }> {
    const response = await fetch(`${this.baseUrl}/api/v1/health`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`Ошибка проверки здоровья API: ${response.status}`);
    }

    return await response.json();
  }

  // Проверка готовности API
  async checkReady(): Promise<{ status: string; timestamp: string; version: string; database: string }> {
    const response = await fetch(`${this.baseUrl}/api/v1/ready`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`Ошибка проверки готовности API: ${response.status}`);
    }

    return await response.json();
  }
}

// Экспортируем единственный экземпляр сервиса
export const apiService = new ApiService();
export default apiService;
