package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config представляет структуру конфигурации
type Config struct {
	Server ServersConfig `yaml:"server"`
	Math   MathConfig    `yaml:"math"`
	CORS   CORSConfig    `yaml:"cors"`
	Logger LoggerConfig  `yaml:"logger"`
}

// Config представляет общую структуру серверов
type ServersConfig struct {
	Orchestrator ServerConfig    `yaml:"orchestrator"`
	Agent        ServerConfig    `yaml:"agent"`
	Web          WebServerConfig `yaml:"web"`
}

// ServerConfig представляет параметры серверов
type ServerConfig struct {
	Port int `toml:"port"`
}

type WebServerConfig struct {
	ServerConfig
	Port      int    `yaml:"port"`
	StaticDir string `yaml:"static"`
}

// MathConfig представляет длительность математически хопераций
type MathConfig struct {
	TIME_ADDITION_MS       int `yaml:"TIME_ADDITION_MS"`
	TIME_SUBTRACTION_MS    int `yaml:"TIME_SUBTRACTION_MS"`
	TIME_MULTIPLICATION_MS int `yaml:"TIME_MULTIPLICATION_MS"`
	TIME_DIVISION_MS       int `yaml:"TIME_DIVISION_MS"`
	TIME_UNARY_MINUS_MS    int `yaml:"TIME_UNARY_MINUS_MS"`
	TIME_POWER_MS          int `yaml:"TIME_POWER_MS"`
}

// CORSConfig представляет параметры CORS
type CORSConfig struct {
	AllowOrigin []string `yaml:"allow_origin"`
}

// LoggerConfig представляет параметры Логгера
type LoggerConfig struct {
	Level        int    `yaml:"level"`
	TimeFormat   string `yaml:"time_format"`
	CallDepth    int    `yaml:"call_depth"`
	DisableCall  bool   `yaml:"disable_call"`
	DisableTime  bool   `yaml:"disable_time"`
	DisableColor bool   `yaml:"disable_color"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Server: ServersConfig{
			Orchestrator: ServerConfig{
				Port: 8080,
			},
			Agent: ServerConfig{
				Port: 8081,
			},
			Web: WebServerConfig{
				Port:      8082,
				StaticDir: "\\static",
			},
		},
		Math: MathConfig{
			TIME_ADDITION_MS:       0,
			TIME_SUBTRACTION_MS:    0,
			TIME_MULTIPLICATION_MS: 0,
			TIME_DIVISION_MS:       0,
			TIME_UNARY_MINUS_MS:    0,
			TIME_POWER_MS:          0,
		},
		CORS: CORSConfig{
			AllowOrigin: []string{"*"},
		},
		Logger: LoggerConfig{
			Level:        0,
			TimeFormat:   "2006-01-02 15:04:05",
			CallDepth:    2,
			DisableCall:  false,
			DisableTime:  false,
			DisableColor: false,
		},
	}
}

// LoadConfig загружает конфигурацию из TOML файла
func LoadConfig(filename string) (*Config, error) {

	// Создаем конфиг по умолчанию
	cfg := DefaultConfig()
	// Получаем абсолютный путь до файла конфигурации (для запуска из любой директории)
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return cfg, fmt.Errorf("ошибка получния абсолютного пути для файла конфигурации %s: %w", filename, err)
	}

	// Проверка существования файла
	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		return cfg, fmt.Errorf("файл конфигурации %s не найден", filename)
	}
	// Открываем файл
	file, err := os.Open(absPath)
	if err != nil {
		return cfg, fmt.Errorf("не удалось открыть файл конфигурации: %w", err)
	}
	defer file.Close()

	// Декодируем TOML файл в структуру

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		if err == io.EOF {
			return cfg, fmt.Errorf("файл конфигурации пуст")
		} else {
			return cfg, fmt.Errorf("не удалось декодировать YAML конфигурацию: %w", err)
		}
	}
	return cfg, nil
}

var (
	// Глобальная переменная для общего использования
	Cfg       *Config
	once      sync.Once
	loadError error
)

func InitConfig(env string) error {
	once.Do(func() {
		filename := fmt.Sprintf("config/%s.yaml", env)
		Cfg, loadError = LoadConfig(filename)
	})
	return loadError
}
