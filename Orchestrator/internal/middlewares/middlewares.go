package middlewares

import (
	"net/http"
	"strings"

	"github.com/OinkiePie/calc_2/pkg/logger"
)

// Handlers - структура для обработчиков запросов, зависит от TaskManager
type Middleware struct {
	// Префикс для ключа аунтификации
	apiKeyPrefix string
	// Ключ аунтификации
	authorization string
	// Список доступных источников
	allowOrigin []string
}

// NewOrchestratorMiddlewares - конструктор для структуры Middleware.
//
// Args:
//
//	ApiKeyPrefix:  string - Префикс API-ключа (например, "Bearer ").
//	Authorization: string - Ожидаемое значение API-ключа для авторизации.
//	AllowOrigin:   []string - Список разрешенных источников для CORS.
//
// Returns:
//
//	*Middleware - Указатель на новый экземпляр структуры Middleware.
func NewOrchestratorMiddlewares(apiKeyPrefix, authorization string, allowOrigin []string) *Middleware {
	if authorization == "" {
		logger.Log.Warnf("Задан пустой ключ авторизации")
	}
	return &Middleware{apiKeyPrefix: apiKeyPrefix, authorization: authorization, allowOrigin: allowOrigin}
}

// EnableAuthorization - проверяет ключ авторизации при запросе на internal endpoints.
//
// Проверяет наличие и корректность API-ключа в заголовке
// Authorization HTTP-запроса. Ключ должен соответствовать ожидаемому значению,
// заданному в поле Authorization структуры Middleware.
//
// Args:
//
//	next: http.Handler - Следующий обработчик в цепочке middleware.
//
// Returns:
//
//	http.Handler - Новый обработчик, который выполняет проверку авторизации перед
//	вызовом следующего обработчика.
//
// Headers:
//
//	Authorization: <ApiKeyPrefix><API-Ключ>
//	Пример: Authorization: Bearer mySecretApiKey
//
// Responses:
//
//	401 Unauthorized:
//	{
//		"error": "Неавторизован: отсутствует заголовок авторизации"
//	}
//
//	{
//		"error": "Неавторизован: неверный формат заголовка авторизации"
//	}
//
//	{
//		"error": "Неавторизован: пустой ключ API"
//	}
//
// 403 Forbidden:
//
//	{
//		"error": "Неавторизован: недействительный ключ API"
//	}
func (m *Middleware) EnableAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Если задан пустой ключ авторизации, пропускает всем запросы
		if m.authorization == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Получаем API-ключ из заголовка Authorization.
		authHeader := r.Header.Get("Authorization")

		// Проверяем, что заголовок Authorization присутствует.
		if authHeader == "" {
			logger.Log.Debugf("Отсутствует заголовок авторизации")
			http.Error(w, "Unauthorized: Missing authorization header", http.StatusUnauthorized) // 401
			return
		}

		// Проверяем, что заголовок начинается с префикса (например, "Bearer ").
		if !strings.HasPrefix(authHeader, m.apiKeyPrefix) {
			logger.Log.Debugf("Неверный формат заголовка авторизации")
			http.Error(w, "Unauthorized: Invalid authorization header format", http.StatusUnauthorized) // 401
			return
		}

		// Извлекаем API-ключ из заголовка.
		apiKey := strings.TrimPrefix(authHeader, m.apiKeyPrefix)

		//.Проверяем, что API-ключ не пустой.
		if apiKey == "" {
			logger.Log.Debugf("Пустой API-ключ")
			http.Error(w, "Unauthorized: Empty API key", http.StatusUnauthorized) // 401
			return
		}

		// Сравниваем полученный API-ключ с ожидаемым.
		if apiKey != m.authorization {
			logger.Log.Debugf("Неверный API-ключ: %s", apiKey)
			http.Error(w, "Unauthorized: Invalid API key", http.StatusForbidden) // 403
			return
		}

		// Если API-ключ валиден, вызываем следующий обработчик в цепочке.
		next.ServeHTTP(w, r)
	})
}

// EnableCORS - добавляет заголовки CORS для разрешения запросов с других доменов.
//
// Добавляет необходимые заголовки CORS (Cross-Origin
// Resource Sharing) для разрешения запросов с других доменов. Она проверяет
// наличие заголовка Origin в запросе и, если он присутствует и входит в список
// разрешенных источников (AllowOrigin), устанавливает соответствующие заголовки
// Access-Control-Allow-Origin, Access-Control-Allow-Methods и
// Access-Control-Allow-Headers в ответе.
//
// Args:
//
//	next: http.Handler - Следующий обработчик в цепочке middleware.
//
// Returns:
//
//	http.Handler - Новый обработчик, который добавляет заголовки CORS перед
//	вызовом следующего обработчика.
func (m *Middleware) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			// Проверяем, есть ли origin в списке разрешенных
			allowed := false
			for _, allowedOrigin := range m.allowOrigin {
				if strings.EqualFold(origin, allowedOrigin) { //Сравнение без учета регистра
					allowed = true
					break
				}
			}
			if allowed {
				// Если origin разрешен, устанавливаем заголовок Access-Control-Allow-Origin
				w.Header().Set("Access-Control-Allow-Origin", origin)
				// Дополнительные заголовки CORS
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
			}
		}

		next.ServeHTTP(w, r)
	})
}
