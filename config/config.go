package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pelletier/go-toml/v2"
)

// Config представляет структуру конфигурации
type Config struct {
	Orchestrator OrchestratorConfig `toml:"Orchestrator"`
	Agent        AgentConfig        `toml:"Agent"`
}

// OrchestratorConfig представляет параметры сервера оркестратора
type OrchestratorConfig struct {
	Port int `toml:"port"`
}

// AgentConfig представляет параметры сервера агента
type AgentConfig struct {
	Port int `toml:"port"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Orchestrator: OrchestratorConfig{
			Port: 8080,
		},
		Agent: AgentConfig{
			Port: 8081,
		},
	}
}

// LoadConfig загружает конфигурацию из TOML файла
func LoadConfig(filename string) (*Config, error) {
	// Создаем конфиг по умолчанию
	cfg := DefaultConfig()
	// Получаем абсолютный путь до файла конфигурации
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return cfg, fmt.Errorf("ошибка получния абсолютного пути для файла конфигурации %s: %w", filename, err)
	}

	// Проверка существования файла
	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		return cfg, fmt.Errorf("файл конфигурации не найден: %s", absPath)
	}
	// Открываем файл
	file, err := os.Open(absPath)
	if err != nil {
		return cfg, fmt.Errorf("не удалось открыть файл конфигурации: %w", err)
	}
	defer file.Close()

	// Декодируем TOML файл в структуру
	decoder := toml.NewDecoder(file)

	if err := decoder.Decode(cfg); err != nil {
		return cfg, fmt.Errorf("не удалось декодировать TOML конфигурацию: %w", err)
	}
	return cfg, nil
}

var (
	// Глобальная переменная для общего использования
	Cfg       *Config
	once      sync.Once
	loadError error
)

func InitConfig() error {
	once.Do(func() {
		Cfg, loadError = LoadConfig("config\\default.toml")
	})
	return loadError
}
