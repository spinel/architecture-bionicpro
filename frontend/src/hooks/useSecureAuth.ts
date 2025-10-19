/**
 * Безопасный хук для аутентификации с PKCE
 * Управляет состоянием аутентификации без передачи токенов на фронтенд
 */

import { useState, useEffect, useCallback } from 'react';
import { SecureAuthService, AuthConfig, TokenResponse } from '../services/authService';

export interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  user: {
    id: string;
    email: string;
    firstName: string;
    lastName: string;
    roles: string[];
  } | null;
}

export interface UseSecureAuthReturn {
  authState: AuthState;
  login: () => Promise<void>;
  logout: () => Promise<void>;
  refreshAuth: () => Promise<void>;
  clearError: () => void;
}

const AUTH_STORAGE_KEY = 'bionicpro_auth_state';
const TOKEN_REFRESH_THRESHOLD = 5 * 60 * 1000; // 5 минут до истечения

export const useSecureAuth = (config: AuthConfig): UseSecureAuthReturn => {
  const [authState, setAuthState] = useState<AuthState>({
    isAuthenticated: false,
    isLoading: true,
    error: null,
    user: null
  });

  const authService = new SecureAuthService(config);

  /**
   * Загружает состояние аутентификации из localStorage
   */
  const loadAuthState = useCallback(() => {
    try {
      const stored = localStorage.getItem(AUTH_STORAGE_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        const now = Date.now();
        
        // Проверяем, не истек ли токен
        if (parsed.expiresAt && now < parsed.expiresAt) {
          setAuthState({
            isAuthenticated: true,
            isLoading: false,
            error: null,
            user: parsed.user
          });
          return;
        }
      }
    } catch (error) {
      console.error('Ошибка при загрузке состояния аутентификации:', error);
    }
    
    // Если токен истек или не найден, очищаем состояние
    clearAuthState();
  }, []);

  /**
   * Сохраняет состояние аутентификации в localStorage
   */
  const saveAuthState = useCallback((tokens: TokenResponse, user: any) => {
    const expiresAt = authService.calculateExpiresAt(tokens.expires_in);
    
    const authData = {
      user,
      expiresAt,
      refreshToken: tokens.refresh_token
    };
    
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(authData));
  }, [authService]);

  /**
   * Очищает состояние аутентификации
   */
  const clearAuthState = useCallback(() => {
    localStorage.removeItem(AUTH_STORAGE_KEY);
    setAuthState({
      isAuthenticated: false,
      isLoading: false,
      error: null,
      user: null
    });
  }, []);

  /**
   * Получает информацию о пользователе из токена
   */
  const getUserInfo = useCallback(async (accessToken: string) => {
    try {
      const userInfoEndpoint = `${config.keycloakUrl}/realms/${config.realm}/protocol/openid-connect/userinfo`;
      const response = await fetch(userInfoEndpoint, {
        headers: {
          'Authorization': `Bearer ${accessToken}`
        }
      });

      if (!response.ok) {
        throw new Error('Не удалось получить информацию о пользователе');
      }

      return await response.json();
    } catch (error) {
      console.error('Ошибка при получении информации о пользователе:', error);
      throw error;
    }
  }, [config]);

  /**
   * Инициирует процесс входа
   */
  const login = useCallback(async () => {
    try {
      setAuthState(prev => ({ ...prev, isLoading: true, error: null }));
      
      const authUrl = await authService.initiateAuth();
      
      // Сохраняем URL для обработки после редиректа
      sessionStorage.setItem('auth_redirect_url', window.location.href);
      
      // Перенаправляем на Keycloak
      window.location.href = authUrl;
    } catch (error) {
      setAuthState(prev => ({
        ...prev,
        isLoading: false,
        error: error instanceof Error ? error.message : 'Ошибка входа'
      }));
    }
  }, [authService]);

  /**
   * Обрабатывает код авторизации после редиректа
   */
  const handleAuthCallback = useCallback(async (code: string) => {
    try {
      setAuthState(prev => ({ ...prev, isLoading: true, error: null }));
      
      const tokens = await authService.exchangeCodeForTokens(code);
      const userInfo = await getUserInfo(tokens.access_token);
      
      saveAuthState(tokens, userInfo);
      
      setAuthState({
        isAuthenticated: true,
        isLoading: false,
        error: null,
        user: {
          id: userInfo.sub,
          email: userInfo.email,
          firstName: userInfo.given_name,
          lastName: userInfo.family_name,
          roles: userInfo.realm_access?.roles || []
        }
      });
      
      // Возвращаемся на исходную страницу
      const redirectUrl = sessionStorage.getItem('auth_redirect_url');
      if (redirectUrl) {
        sessionStorage.removeItem('auth_redirect_url');
        window.location.href = redirectUrl;
      }
    } catch (error) {
      setAuthState(prev => ({
        ...prev,
        isLoading: false,
        error: error instanceof Error ? error.message : 'Ошибка аутентификации'
      }));
    }
  }, [authService, getUserInfo, saveAuthState]);

  /**
   * Выполняет выход из системы
   */
  const logout = useCallback(async () => {
    try {
      const stored = localStorage.getItem(AUTH_STORAGE_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        if (parsed.refreshToken) {
          await authService.logout(parsed.refreshToken);
        }
      }
    } catch (error) {
      console.error('Ошибка при выходе:', error);
    } finally {
      clearAuthState();
    }
  }, [authService, clearAuthState]);

  /**
   * Обновляет токены при необходимости
   */
  const refreshAuth = useCallback(async () => {
    try {
      const stored = localStorage.getItem(AUTH_STORAGE_KEY);
      if (!stored) return;

      const parsed = JSON.parse(stored);
      const now = Date.now();
      
      // Проверяем, нужно ли обновить токен
      if (parsed.expiresAt && (parsed.expiresAt - now) < TOKEN_REFRESH_THRESHOLD) {
        if (parsed.refreshToken) {
          const newTokens = await authService.refreshTokens(parsed.refreshToken);
          const userInfo = await getUserInfo(newTokens.access_token);
          
          saveAuthState(newTokens, userInfo);
          
          setAuthState(prev => ({
            ...prev,
            user: {
              id: userInfo.sub,
              email: userInfo.email,
              firstName: userInfo.given_name,
              lastName: userInfo.family_name,
              roles: userInfo.realm_access?.roles || []
            }
          }));
        }
      }
    } catch (error) {
      console.error('Ошибка при обновлении токенов:', error);
      clearAuthState();
    }
  }, [authService, getUserInfo, saveAuthState, clearAuthState]);

  /**
   * Очищает ошибки
   */
  const clearError = useCallback(() => {
    setAuthState(prev => ({ ...prev, error: null }));
  }, []);

  // Загружаем состояние при инициализации
  useEffect(() => {
    loadAuthState();
  }, [loadAuthState]);

  // Проверяем URL на наличие кода авторизации
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');
    const error = urlParams.get('error');
    
    if (error) {
      setAuthState(prev => ({
        ...prev,
        isLoading: false,
        error: `Ошибка аутентификации: ${error}`
      }));
    } else if (code) {
      handleAuthCallback(code);
    }
  }, [handleAuthCallback]);

  // Автоматическое обновление токенов
  useEffect(() => {
    if (authState.isAuthenticated) {
      const interval = setInterval(refreshAuth, 60000); // Проверяем каждую минуту
      return () => clearInterval(interval);
    }
  }, [authState.isAuthenticated, refreshAuth]);

  return {
    authState,
    login,
    logout,
    refreshAuth,
    clearError
  };
};
