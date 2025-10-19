import React, { useState } from 'react';

const SimpleAuth: React.FC = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedUser, setSelectedUser] = useState<string>('prothetic1');
  const [userRole, setUserRole] = useState<string>('prothetic_user');

  const handleLogin = async () => {
    setLoading(true);
    setError(null);
    
    try {
      // Простая имитация аутентификации с определением роли
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // Определяем роль пользователя
      if (selectedUser === 'prothetic1') {
        setUserRole('prothetic_user');
      } else if (selectedUser === 'user1') {
        setUserRole('regular_user');
      } else {
        setUserRole('admin');
      }
      
      setIsAuthenticated(true);
    } catch (err) {
      setError('Ошибка входа');
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    setIsAuthenticated(false);
  };

  const downloadReport = async () => {
    setLoading(true);
    setError(null);
    
    try {
      // Имитация скачивания отчета
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      // Создаем простой отчет
      const reportContent = `
ОТЧЕТ О РАБОТЕ ПРОТЕЗА
======================

Дата создания: ${new Date().toLocaleDateString('ru-RU')}
Пользователь: ${selectedUser}

СТАТИСТИКА ИСПОЛЬЗОВАНИЯ
------------------------
- Общее время использования: 8 часов 30 минут
- Количество движений: 1,247
- Средняя точность: 94.2%
- Количество сессий: 12

ТЕХНИЧЕСКОЕ СОСТОЯНИЕ
--------------------
- Уровень заряда батареи: 87%
- Температура: 23°C
- Статус датчиков: Норма
- Последняя калибровка: ${new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toLocaleDateString('ru-RU')}

РЕКОМЕНДАЦИИ
------------
1. Рекомендуется провести калибровку датчиков
2. Уровень заряда батареи в норме
3. Производительность в пределах нормы

---
Сгенерировано системой BionicPRO
Дата: ${new Date().toLocaleString('ru-RU')}
      `;
      
      // Создаем и скачиваем файл
      const blob = new Blob([reportContent], { type: 'text/plain;charset=utf-8' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `prothetic-report-${selectedUser}-${new Date().toISOString().split('T')[0]}.txt`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      
    } catch (err) {
      setError('Ошибка генерации отчета');
    } finally {
      setLoading(false);
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100">
        <div className="p-8 bg-white rounded-lg shadow-md text-center max-w-md w-full">
          <h1 className="text-2xl font-bold mb-6 text-gray-800">BionicPRO Reports</h1>
          <p className="text-gray-600 mb-6">Выберите пользователя для входа в систему</p>
          
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Выберите пользователя:
            </label>
            <select
              value={selectedUser}
              onChange={(e) => setSelectedUser(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="prothetic1">prothetic1 (Пользователь протеза) - Может скачивать отчеты</option>
              <option value="user1">user1 (Обычный пользователь) - Не может скачивать отчеты</option>
              <option value="admin">admin (Администратор) - Полный доступ</option>
            </select>
          </div>
          
          <button
            onClick={handleLogin}
            disabled={loading}
            className={`px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors ${
              loading ? 'opacity-50 cursor-not-allowed' : ''
            }`}
          >
            {loading ? 'Вход...' : `Войти как ${selectedUser}`}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">BionicPRO Reports</h1>
              <p className="text-sm text-gray-600">
                Добро пожаловать, {selectedUser} ({userRole === 'prothetic_user' ? 'Пользователь протеза' : userRole === 'regular_user' ? 'Обычный пользователь' : 'Администратор'})
              </p>
            </div>
            <button
              onClick={handleLogout}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            >
              Выйти
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <div className="bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Отчеты о работе протеза</h2>
            <p className="text-sm text-gray-600 mb-6">
              Скачайте подробный отчет о работе вашего протеза за выбранный период.
              Отчет включает данные о движениях, использовании и техническом состоянии.
            </p>
            
            <div className="space-y-4">
              {userRole === 'prothetic_user' || userRole === 'admin' ? (
                <button
                  onClick={downloadReport}
                  disabled={loading}
                  className={`w-full sm:w-auto px-6 py-3 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 ${
                    loading ? 'opacity-50 cursor-not-allowed' : ''
                  }`}
                >
                  {loading ? (
                    <div className="flex items-center justify-center">
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                      Генерация отчета...
                    </div>
                  ) : (
                    'Скачать отчет'
                  )}
                </button>
              ) : (
                <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4">
                  <div className="flex">
                    <div className="flex-shrink-0">
                      <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                      </svg>
                    </div>
                    <div className="ml-3">
                      <h3 className="text-sm font-medium text-yellow-800">Доступ ограничен</h3>
                      <div className="mt-2 text-sm text-yellow-700">
                        У вас нет прав для скачивания отчетов. Обратитесь к администратору для получения доступа.
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {error && (
                <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-md">
                  <div className="flex">
                    <div className="flex-shrink-0">
                      <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                      </svg>
                    </div>
                    <div className="ml-3">
                      <h3 className="text-sm font-medium text-red-800">Ошибка</h3>
                      <div className="mt-2 text-sm text-red-700">
                        {error}
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};

export default SimpleAuth;
