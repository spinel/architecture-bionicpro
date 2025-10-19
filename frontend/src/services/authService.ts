/**
 * Безопасный сервис аутентификации с поддержкой PKCE
 * Обеспечивает безопасную работу с токенами без их передачи на фронтенд
 */

import { generatePKCEParams, getStoredCodeVerifier, clearStoredCodeVerifier } from '../utils/pkce';

export interface AuthConfig {
  keycloakUrl: string;
  realm: string;
  clientId: string;
  redirectUri: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

export class SecureAuthService {
  private config: AuthConfig;
  private tokenEndpoint: string;
  private authEndpoint: string;

  constructor(config: AuthConfig) {
    this.config = config;
    this.tokenEndpoint = `${config.keycloakUrl}/realms/${config.realm}/protocol/openid-connect/token`;
    this.authEndpoint = `${config.keycloakUrl}/realms/${config.realm}/protocol/openid-connect/auth`;
  }

  /**
   * Инициирует процесс аутентификации с PKCE
   * @returns {Promise<string>} URL для редиректа на аутентификацию
   */
  async initiateAuth(): Promise<string> {
    try {
      const { challenge } = await generatePKCEParams();
      
      const params = new URLSearchParams({
        client_id: this.config.clientId,
        redirect_uri: this.config.redirectUri,
        response_type: 'code',
        scope: 'openid profile email',
        code_challenge: challenge,
        code_challenge_method: 'S256',
        state: this.generateState()
      });

      return `${this.authEndpoint}?${params.toString()}`;
    } catch (error) {
      console.error('Ошибка при инициализации аутентификации:', error);
      throw new Error('Не удалось инициализировать аутентификацию');
    }
  }

  /**
   * Обменивает код авторизации на токены
   * @param {string} code - Код авторизации
   * @returns {Promise<TokenResponse>} Токены доступа
   */
  async exchangeCodeForTokens(code: string): Promise<TokenResponse> {
    const verifier = getStoredCodeVerifier();
    if (!verifier) {
      throw new Error('Code verifier не найден. Возможно, сессия истекла.');
    }

    try {
      const response = await fetch(this.tokenEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
          grant_type: 'authorization_code',
          client_id: this.config.clientId,
          client_secret: 'C60JxJONZAvMyoHHjVg84UbAwy4foEyr',
          code: code,
          redirect_uri: this.config.redirectUri,
          code_verifier: verifier
        })
      });

      if (!response.ok) {
        throw new Error(`Ошибка обмена кода на токены: ${response.status}`);
      }

      const tokens = await response.json();
      
      // Очищаем code_verifier после успешного обмена
      clearStoredCodeVerifier();
      
      return tokens;
    } catch (error) {
      console.error('Ошибка при обмене кода на токены:', error);
      throw new Error('Не удалось получить токены доступа');
    }
  }

  /**
   * Обновляет access token используя refresh token
   * @param {string} refreshToken - Refresh token
   * @returns {Promise<TokenResponse>} Новые токены
   */
  async refreshTokens(refreshToken: string): Promise<TokenResponse> {
    try {
      const response = await fetch(this.tokenEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
          grant_type: 'refresh_token',
          client_id: this.config.clientId,
          refresh_token: refreshToken
        })
      });

      if (!response.ok) {
        throw new Error(`Ошибка обновления токенов: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Ошибка при обновлении токенов:', error);
      throw new Error('Не удалось обновить токены доступа');
    }
  }

  /**
   * Выполняет выход из системы
   * @param {string} refreshToken - Refresh token для выхода
   */
  async logout(refreshToken: string): Promise<void> {
    try {
      const logoutEndpoint = `${this.config.keycloakUrl}/realms/${this.config.realm}/protocol/openid-connect/logout`;
      
      await fetch(logoutEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
          client_id: this.config.clientId,
          refresh_token: refreshToken
        })
      });
    } catch (error) {
      console.error('Ошибка при выходе из системы:', error);
      // Не выбрасываем ошибку, так как выход должен быть выполнен локально в любом случае
    }
  }

  /**
   * Генерирует случайный state параметр для защиты от CSRF
   * @returns {string} Случайный state
   */
  private generateState(): string {
    const array = new Uint8Array(16);
    crypto.getRandomValues(array);
    return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
  }

  /**
   * Проверяет, истек ли токен
   * @param {number} expiresAt - Время истечения токена (timestamp)
   * @returns {boolean} true если токен истек
   */
  isTokenExpired(expiresAt: number): boolean {
    return Date.now() >= expiresAt;
  }

  /**
   * Вычисляет время истечения токена
   * @param {number} expiresIn - Время жизни токена в секундах
   * @returns {number} Timestamp времени истечения
   */
  calculateExpiresAt(expiresIn: number): number {
    return Date.now() + (expiresIn * 1000);
  }
}
