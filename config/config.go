package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

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
	ADDR_ORCHESTRATOR string `yaml:"ADDR_ORCHESTRATOR"`
	PORT_ORCHESTRATOR int    `yaml:"PORT_ORCHESTRATOR"`
}

// AgentServiceConfig структура параметров агента
type AgentServiceConfig struct {
	COMPUTING_POWER  int `yaml:"COMPUTING_POWER"`
	AGENT_REPEAT     int `yaml:"AGENT_REPEAT"`
	AGENT_REPEAT_ERR int `yaml:"AGENT_REPEAT_ERR"`
}

// WebServiceConfig структура параметров веб сервиса
type WebServiceConfig struct {
	ADDR_WEB  string `yaml:"ADDR_WEB"`
	PORT_WEB  int    `yaml:"PORT_WEB"`
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
func defaultConfig() *Config {
	return &Config{
		Server: ServicesConfig{
			Orchestrator: OrchestratorServiceConfig{
				ADDR_ORCHESTRATOR: "127.0.0.1",
				PORT_ORCHESTRATOR: 8080,
			},
			Agent: AgentServiceConfig{
				COMPUTING_POWER:  4,
				AGENT_REPEAT:     5000,
				AGENT_REPEAT_ERR: 2000,
			},
			Web: WebServiceConfig{
				ADDR_WEB:  "127.0.0.1",
				StaticDir: "web/static",
				PORT_WEB:  8081,
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
	Filename string
	Cfg      *Config
)

func loadName() {
	// Загрузка env переменных из файла .env
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден")
	}
	// Определение типа приложения - prod или dev
	path := os.Getenv("APP_CFG")

	if path == "" {
		log.Println("Переменная среды APP_CFG отстутствует или пуста, используется конфигурация по умолчанию")
		path = "config/configs/dev.yml" // По умолчанию - разработка
	} else if path == "CFG_FALSE" {
		log.Println(`Переменная среды APP_CFG равна "CFG_FALSE". Файл конфигурации отключен.`)
		Filename = ""
		return
	}

	Filename = path

}

func loadEnv() error {
	// ADDR_ORCHESTRATOR
	addrOrchestrator := os.Getenv("ADDR_ORCHESTRATOR")
	if addrOrchestrator != "" {

		Cfg.Server.Orchestrator.ADDR_ORCHESTRATOR = addrOrchestrator
	}

	// ADDR_WEB
	addrWeb := os.Getenv("ADDR_WEB")
	if addrWeb != "" {
		Cfg.Server.Web.ADDR_WEB = addrWeb
	}

	// PORT_ORCHESTRATOR
	portOrchestratorStr := os.Getenv("PORT_ORCHESTRATOR")
	if portOrchestratorStr != "" {
		portOrchestrator, err := strconv.Atoi(portOrchestratorStr)
		if err != nil {
			return fmt.Errorf("ошибка преобразования PORT_ORCHESTRATOR в int: %w", err)
		}
		Cfg.Server.Orchestrator.PORT_ORCHESTRATOR = portOrchestrator
	}

	// PORT_WEB
	portWebStr := os.Getenv("PORT_WEB")
	if portWebStr != "" {
		portWeb, err := strconv.Atoi(portWebStr)
		if err != nil {
			return fmt.Errorf("ошибка преобразования PORT_WEB в int: %w", err)
		}
		Cfg.Server.Web.PORT_WEB = portWeb
	}

	// COMPUTING_POWER
	computingPowerStr := os.Getenv("COMPUTING_POWER")
	if computingPowerStr != "" {
		computingPower, err := strconv.Atoi(computingPowerStr)
		if err != nil {
			return fmt.Errorf("ошибка преобразования TIME_POWER_MS в int: %w", err)
		}
		Cfg.Math.TIME_POWER_MS = computingPower
	}

	// AGENT_REPEAT_ERR
	agentRepeatErrStr := os.Getenv("AGENT_REPEAT_ERR")
	if agentRepeatErrStr != "" {
		agentRepeatErr, err := strconv.Atoi(agentRepeatErrStr)
		if err != nil {
			return fmt.Errorf("ошибка преобразования AGENT_REPEAT_ERR в int: %w", err)
		}
		Cfg.Server.Agent.AGENT_REPEAT_ERR = agentRepeatErr
	}

	// AGENT_REPEAT
	agentRepeatStr := os.Getenv("AGENT_REPEAT")
	if agentRepeatStr != "" {
		agentRepeat, err := strconv.Atoi(agentRepeatStr)
		if err != nil {
			return fmt.Errorf("ошибка преобразования AGENT_REPEAT в int: %w", err)
		}
		Cfg.Server.Agent.AGENT_REPEAT = agentRepeat
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
	// Создаем конфиг по умолчанию
	Cfg = defaultConfig()

	// Ищем название файла конфигурации
	loadName()

	if Filename != "" {
		// Проверка существования файла
		_, err := os.Stat(Filename)
		if os.IsNotExist(err) {
			return fmt.Errorf("файл конфигурации %s не найден", Filename)
		}
		// Открываем файл
		file, err := os.Open(Filename)
		// Проверка прав
		if os.IsPermission(err) {
			return fmt.Errorf("недостаточно прав чтобы открыть %s", Filename)
		}

		if err != nil {
			return fmt.Errorf("не удалось открыть файл конфигурации: %w", err)
		}
		defer file.Close()

		// Декодируем YAML файл в структуру
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(Cfg); err != nil {
			if err == io.EOF {
				return fmt.Errorf("файл конфигурации пуст")

			} else {
				return fmt.Errorf("не удалось декодировать YAML конфигурацию: %w", err)

			}
		}
	}

	// Записываем переменные среды поверх других
	err := loadEnv()
	return err
}
