import React, { useEffect } from 'react';
import { ReactKeycloakProvider } from '@react-keycloak/web';
import Keycloak, { KeycloakConfig, KeycloakInitOptions } from 'keycloak-js';
import ReportPage from './components/ReportPage';
import { 
  generateCodeVerifier, 
  generateCodeChallenge, 
  saveCodeVerifier,
  clearCodeVerifier 
} from './utils/pkce';

// Конфигурация Keycloak
const keycloakConfig: KeycloakConfig = {
  url: process.env.REACT_APP_KEYCLOAK_URL,
  realm: process.env.REACT_APP_KEYCLOAK_REALM || "",
  clientId: process.env.REACT_APP_KEYCLOAK_CLIENT_ID || ""
};

// Создание экземпляра Keycloak
const keycloak = new Keycloak(keycloakConfig);


// @ts-ignore - отключаем SSL для разработки
keycloak.sslRequired = 'none';

/**
 * Настройка PKCE для Keycloak
 * 
 * PKCE (Proof Key for Code Exchange) защищает от атак перехвата authorization code:
 * 1. Генерируется случайный code_verifier
 * 2. Создается code_challenge = SHA256(code_verifier)
 * 3. code_challenge отправляется в authorization request
 * 4. code_verifier отправляется при обмене code на токены
 * 5. Keycloak проверяет: SHA256(code_verifier) == code_challenge
 */
const initOptions: KeycloakInitOptions = {
  onLoad: 'check-sso',
  silentCheckSsoRedirectUri: window.location.origin + '/silent-check-sso.html',
  // ВАЖНО: Включаем PKCE с методом S256 (SHA-256)
  pkceMethod: 'S256',
  checkLoginIframe: true,
  checkLoginIframeInterval: 5,
};

const App: React.FC = () => {
  useEffect(() => {
    // Настройка обработчиков событий Keycloak
    keycloak.onAuthSuccess = () => {
      console.log('[Keycloak] Аутентификация успешна');
      // Очищаем code_verifier после успешной аутентификации
      clearCodeVerifier();
    };

    keycloak.onAuthError = (error) => {
      console.error('[Keycloak] Ошибка аутентификации:', error);
      clearCodeVerifier();
    };

    keycloak.onTokenExpired = () => {
      console.log('[Keycloak] Токен истек, обновляем...');
      keycloak.updateToken(30).catch(() => {
        console.error('[Keycloak] Не удалось обновить токен');
      });
    };

    keycloak.onAuthLogout = () => {
      console.log('[Keycloak] Выход выполнен');
      clearCodeVerifier();
    };
  }, []);

  // Обработчик событий токена
  const onTokens = (tokens: any) => {
    if (tokens.token) {
      console.log('[Keycloak] Токены получены');
    }
  };

  return (
    <ReactKeycloakProvider 
      authClient={keycloak}
      initOptions={initOptions}
      onTokens={onTokens}
    >
      <div className="App">
        <ReportPage />
      </div>
    </ReactKeycloakProvider>
  );
};

export default App;