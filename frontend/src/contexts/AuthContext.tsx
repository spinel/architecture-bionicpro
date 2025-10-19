/**
 * Контекст аутентификации для безопасной работы с PKCE
 * Предоставляет единую точку доступа к состоянию аутентификации
 */

import React, { createContext, useContext, ReactNode } from 'react';
import { useSecureAuth, AuthState, UseSecureAuthReturn } from '../hooks/useSecureAuth';

interface AuthContextType extends UseSecureAuthReturn {}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const authConfig = {
    keycloakUrl: process.env.REACT_APP_KEYCLOAK_URL || 'https://localhost:8443',
    realm: process.env.REACT_APP_KEYCLOAK_REALM || 'reports-realm',
    clientId: process.env.REACT_APP_KEYCLOAK_CLIENT_ID || 'reports-frontend',
    redirectUri: window.location.origin + window.location.pathname
  };

  const auth = useSecureAuth(authConfig);

  return (
    <AuthContext.Provider value={auth}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth должен использоваться внутри AuthProvider');
  }
  return context;
};
