/**
 * PKCE (Proof Key for Code Exchange) утилиты для BionicPRO
 * RFC 7636: https://datatracker.ietf.org/doc/html/rfc7636
 * 
 * PKCE защищает от атак перехвата authorization code, что критично для:
 * - SPA приложений (как это)
 * - Мобильных приложений
 * - Любых публичных клиентов, которые не могут безопасно хранить client secret
 */

/**
 * Генерирует случайный code_verifier
 * 
 * Требования RFC 7636:
 * - Длина: 43-128 символов
 * - Символы: [A-Z] [a-z] [0-9] - . _ ~
 * - Энтропия: минимум 256 бит
 * 
 * @returns {string} Base64-URL закодированная случайная строка
 */
export function generateCodeVerifier(): string {
  // Генерируем 32 байта (256 бит) случайных данных
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  
  // Кодируем в Base64-URL формат
  return base64UrlEncode(array);
}

/**
 * Создает code_challenge из code_verifier используя SHA-256
 * 
 * Формула: code_challenge = BASE64URL(SHA256(ASCII(code_verifier)))
 * 
 * @param {string} verifier - code_verifier для которого генерируется challenge
 * @returns {Promise<string>} Base64-URL закодированный SHA-256 хеш
 */
export async function generateCodeChallenge(verifier: string): Promise<string> {
  // Конвертируем строку в байты
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  
  // Вычисляем SHA-256 хеш
  const hash = await crypto.subtle.digest('SHA-256', data);
  
  // Кодируем в Base64-URL формат
  return base64UrlEncode(new Uint8Array(hash));
}

/**
 * Base64-URL кодирование (RFC 4648 Section 5)
 * 
 * Отличия от обычного Base64:
 * - Символ '+' заменяется на '-'
 * - Символ '/' заменяется на '_'
 * - Удаляется padding (символы '=')
 * 
 * @param {Uint8Array} buffer - Данные для кодирования
 * @returns {string} Base64-URL закодированная строка
 */
function base64UrlEncode(buffer: Uint8Array): string {
  // Конвертируем Uint8Array в строку
  const base64 = btoa(String.fromCharCode.apply(null, Array.from(buffer)));
  
  // Применяем Base64-URL преобразования
  return base64
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

/**
 * Сохраняет code_verifier в sessionStorage
 * 
 * Важно: Используем sessionStorage вместо localStorage, потому что:
 * - sessionStorage автоматически очищается при закрытии вкладки
 * - code_verifier должен быть одноразовым и не должен переживать сессию
 * - Это уменьшает риск утечки или повторного использования verifier
 * 
 * @param {string} verifier - code_verifier для сохранения
 */
export function saveCodeVerifier(verifier: string): void {
  sessionStorage.setItem('pkce_code_verifier', verifier);
  console.log('[PKCE] Code verifier сохранен в sessionStorage');
}

/**
 * Получает code_verifier из sessionStorage
 * 
 * @returns {string | null} Сохраненный code_verifier или null если не найден
 */
export function getCodeVerifier(): string | null {
  const verifier = sessionStorage.getItem('pkce_code_verifier');
  if (verifier) {
    console.log('[PKCE] Code verifier получен из sessionStorage');
  }
  return verifier;
}

/**
 * Удаляет code_verifier из sessionStorage
 * 
 * Должно вызываться после успешного обмена authorization code на токены,
 * чтобы предотвратить повторное использование verifier
 */
export function clearCodeVerifier(): void {
  sessionStorage.removeItem('pkce_code_verifier');
  console.log('[PKCE] Code verifier удален из sessionStorage');
}

/**
 * Валидация code_verifier
 * 
 * Проверяет соответствие требованиям RFC 7636:
 * - Длина 43-128 символов
 * - Только разрешенные символы [A-Za-z0-9\-._~]
 * 
 * @param {string} verifier - code_verifier для проверки
 * @returns {boolean} true если verifier валиден
 */
export function isValidCodeVerifier(verifier: string): boolean {
  // Проверка длины
  if (verifier.length < 43 || verifier.length > 128) {
    console.error('[PKCE] Invalid verifier length:', verifier.length);
    return false;
  }
  
  // Проверка допустимых символов
  const validCharacters = /^[A-Za-z0-9\-._~]+$/;
  if (!validCharacters.test(verifier)) {
    console.error('[PKCE] Invalid characters in verifier');
    return false;
  }
  
  return true;
}

/**
 * Генерирует state для защиты от CSRF атак
 * 
 * State параметр используется для валидации того, что authorization response
 * соответствует нашему request (защита от CSRF)
 * 
 * @returns {string} Случайная строка для использования как state
 */
export function generateState(): string {
  const array = new Uint8Array(16);
  crypto.getRandomValues(array);
  return base64UrlEncode(array);
}

/**
 * Сохраняет state в sessionStorage
 * 
 * @param {string} state - State для сохранения
 */
export function saveState(state: string): void {
  sessionStorage.setItem('oauth_state', state);
  console.log('[PKCE] State сохранен');
}

/**
 * Проверяет state при возврате с IdP
 * 
 * @param {string} state - State из authorization response
 * @returns {boolean} true если state соответствует сохраненному
 */
export function validateState(state: string): boolean {
  const savedState = sessionStorage.getItem('oauth_state');
  sessionStorage.removeItem('oauth_state');
  
  const isValid = savedState === state;
  if (!isValid) {
    console.error('[PKCE] State validation failed');
  } else {
    console.log('[PKCE] State validated successfully');
  }
  
  return isValid;
}

/**
 * Очищает все PKCE данные из sessionStorage
 * 
 * Используется при logout или при возникновении ошибок
 */
export function clearAllPKCEData(): void {
  clearCodeVerifier();
  sessionStorage.removeItem('oauth_state');
  console.log('[PKCE] Все PKCE данные очищены');
}

