/**
 * Утилиты для работы с PKCE (Proof Key for Code Exchange)
 * Обеспечивает безопасность OAuth2 Authorization Code Flow
 */

/**
 * Генерирует случайный code_verifier для PKCE
 * @returns {string} Base64URL-encoded code_verifier
 */
export const generateCodeVerifier = (): string => {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return base64URLEncode(array);
};

/**
 * Генерирует code_challenge из code_verifier
 * @param {string} verifier - code_verifier
 * @returns {Promise<string>} Base64URL-encoded SHA256 hash
 */
export const generateCodeChallenge = async (verifier: string): Promise<string> => {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return base64URLEncode(new Uint8Array(digest));
};

/**
 * Кодирует массив байтов в Base64URL
 * @param {Uint8Array} array - Массив байтов
 * @returns {string} Base64URL-encoded строка
 */
const base64URLEncode = (array: Uint8Array): string => {
  const binaryString = Array.from(array, byte => String.fromCharCode(byte)).join('');
  return btoa(binaryString)
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
};

/**
 * Сохраняет code_verifier в sessionStorage для последующего использования
 * @param {string} verifier - code_verifier
 */
export const storeCodeVerifier = (verifier: string): void => {
  sessionStorage.setItem('pkce_code_verifier', verifier);
};

/**
 * Получает сохраненный code_verifier из sessionStorage
 * @returns {string | null} code_verifier или null если не найден
 */
export const getStoredCodeVerifier = (): string | null => {
  return sessionStorage.getItem('pkce_code_verifier');
};

/**
 * Удаляет code_verifier из sessionStorage
 */
export const clearStoredCodeVerifier = (): void => {
  sessionStorage.removeItem('pkce_code_verifier');
};

/**
 * Генерирует полный набор PKCE параметров
 * @returns {Promise<{verifier: string, challenge: string}>} PKCE параметры
 */
export const generatePKCEParams = async (): Promise<{verifier: string, challenge: string}> => {
  const verifier = generateCodeVerifier();
  const challenge = await generateCodeChallenge(verifier);
  
  // Сохраняем verifier для последующего использования
  storeCodeVerifier(verifier);
  
  return { verifier, challenge };
};
