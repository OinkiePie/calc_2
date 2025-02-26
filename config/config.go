package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config представляет структуру конфигурации
type Config struct {
	Server     ServicesConfig   `yaml:"server"`
	Math       MathConfig       `yaml:"math"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	Logger     LoggerConfig     `yaml:"logger"`
}

// ServicesConfig представляет общую структуру сервисов
type ServicesConfig struct {
	Orchestrator OrchestratorServiceConfig `yaml:"orchestrator"`
	Agent        AgentServiceConfig        `yaml:"agent"`
	Web          WebServiceConfig          `yaml:"web"`
}

// OrchestratorServiceConfig структура параметров оркестратора
type OrchestratorServiceConfig struct {
	Port int `yaml:"port"`
}

// AgentServiceConfig структура параметров агента
type AgentServiceConfig struct {
	COMPUTING_POWER int `yaml:"COMPUTING_POWER"`
}

// WebServiceConfig структура параметров веб сервиса
type WebServiceConfig struct {
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
type MiddlewareConfig struct {
	ApiKeyPrefix  string   `yaml:"api_key_prefix"`
	Authorization string   `yaml:"authorization"`
	AllowOrigin   []string `yaml:"cors_allow_origin"`
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
		Server: ServicesConfig{
			Orchestrator: OrchestratorServiceConfig{
				Port: 8080,
			},
			Agent: AgentServiceConfig{
				COMPUTING_POWER: 4,
			},
			Web: WebServiceConfig{
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
		Middleware: MiddlewareConfig{
			ApiKeyPrefix:  "",
			Authorization: "",
			AllowOrigin:   []string{"*"},
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

var (
	// Глобальная переменная для общего использования
	Name      string
	Cfg       *Config
	once      sync.Once
	loadError error
)

func loadEnv() error {
	// Загрузка env переменных из файла .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("Файл .env не найден")
	}
	// Определение типа приложения - prod или dev
	env := os.Getenv("APP_ENV")

	if env == "" {
		fmt.Println("Переменная среды APP_ENV отстутствует или пуста, используется конфигурация по умолчанию")
		env = "dev" // По умолчанию - разработка
	}

	Name = env

	// COMPUTING_POWER
	computingPowerStr := os.Getenv("COMPUTING_POWER")
	if computingPowerStr != "" {
		computingPower, err := strconv.Atoi(computingPowerStr)
		if err != nil {
			return fmt.Errorf("ошибка преобразования TIME_POWER_MS в int: %w", err)
		}
		Cfg.Math.TIME_POWER_MS = computingPower
	}

	// TIME_ADDITION_MS
	timeAdditionMSStr := os.Getenv("TIME_ADDITION_MS")
	if timeAdditionMSStr != "" {
		timeAdditionMS, err := strconv.Atoi(timeAdditionMSStr)
		if err != nil {
			return fmt.Errorf("ошибка преобразования TIME_ADDITION_MS в int: %w", err)
		}
		Cfg.Math.TIME_ADDITION_MS = timeAdditionMS
	}

	// TIME_SUBTRACTION_MS
	timeSubtractionMSStr := os.Getenv("TIME_SUBTRACTION_MS")
	if timeSubtractionMSStr != "" {
		timeSubtractionMS, err := strconv.Atoi(timeSubtractionMSStr)
		if err != nil {
			return fmt.Errorf("ошибка преобразования TIME_SUBTRACTION_MS в int: %w", err)
		}
		Cfg.Math.TIME_SUBTRACTION_MS = timeSubtractionMS
	}

	// TIME_MULTIPLICATION_MS
	timeMultiplicationMSStr := os.Getenv("TIME_MULTIPLICATION_MS")
	if timeMultiplicationMSStr != "" {
		timeMultiplicationMS, err := strconv.Atoi(timeMultiplicationMSStr)
		if err != nil {
			return fmt.Errorf("ошибка преобразования TIME_MULTIPLICATION_MS в int: %w", err)
		}
		Cfg.Math.TIME_MULTIPLICATION_MS = timeMultiplicationMS
	}

	// TIME_DIVISION_MS
	timeDivisionMSStr := os.Getenv("TIME_DIVISION_MS")
	if timeDivisionMSStr != "" {
		timeDivisionMS, err := strconv.Atoi(timeDivisionMSStr)
		if err != nil {
			return fmt.Errorf("ошибка преобразования TIME_DIVISION_MS в int: %w", err)
		}
		Cfg.Math.TIME_DIVISION_MS = timeDivisionMS
	}
	// TIME_UNARY_MINUS_MS
	timeUnaryMinusMSStr := os.Getenv("TIME_UNARY_MINUS_MS")
	if timeUnaryMinusMSStr != "" {
		timeUnaryMinusMS, err := strconv.Atoi(timeUnaryMinusMSStr)
		if err != nil {
			return fmt.Errorf("ошибка преобразования TIME_UNARY_MINUS_MS в int: %w", err)
		}
		Cfg.Math.TIME_UNARY_MINUS_MS = timeUnaryMinusMS
	}
	// TIME_POWER_MS
	timePowerMSStr := os.Getenv("TIME_POWER_MS")
	if timePowerMSStr != "" {
		timePowerMS, err := strconv.Atoi(timePowerMSStr)
		if err != nil {
			return fmt.Errorf("ошибка преобразования TIME_POWER_MS в int: %w", err)
		}
		Cfg.Math.TIME_POWER_MS = timePowerMS
	}

	return nil

}

func InitConfig() error {
	once.Do(func() {
		// Создаем конфиг по умолчанию
		Cfg = DefaultConfig()

		// Загружает параметры из переменной среды
		// Перезаписывают значение по умолчанию, перезаписываются YAML конфигом
		err := loadEnv()
		if err != nil {
			fmt.Println(err.Error())
		}

		filename := fmt.Sprintf("config/%s.yaml", Name)
		// Получаем абсолютный путь до файла конфигурации (для запуска из любой директории)
		absPath, err := filepath.Abs(filename)
		if err != nil {
			loadError = fmt.Errorf("ошибка получния абсолютного пути для файла конфигурации %s: %w", filename, err)
			return
		}

		// Проверка существования файла
		_, err = os.Stat(absPath)
		if os.IsNotExist(err) {
			loadError = fmt.Errorf("файл конфигурации %s не найден", filename)
			return
		}
		// Открываем файл
		file, err := os.Open(absPath)
		if err != nil {
			loadError = fmt.Errorf("не удалось открыть файл конфигурации: %w", err)
			return
		}
		defer file.Close()

		// Декодируем YAML файл в структуру
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(Cfg); err != nil {
			if err == io.EOF {
				loadError = fmt.Errorf("файл конфигурации пуст")
				return
			} else {
				loadError = fmt.Errorf("не удалось декодировать YAML конфигурацию: %w", err)
				return
			}
		}
	})

	return loadError
}
