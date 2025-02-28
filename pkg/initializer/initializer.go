package initializer

import (
	"github.com/OinkiePie/calc_2/config"
	"github.com/OinkiePie/calc_2/pkg/logger"
)

// Init инициализирует конфигурацию и логгер.
func Init() {
	err := config.InitConfig()

	logger.InitLogger(logger.Options{
		Level:        logger.Level(config.Cfg.Logger.Level),
		TimeFormat:   config.Cfg.Logger.TimeFormat,
		CallDepth:    config.Cfg.Logger.CallDepth,
		DisableCall:  config.Cfg.Logger.DisableCall,
		DisableTime:  config.Cfg.Logger.DisableTime,
		DisableColor: config.Cfg.Logger.DisableColor,
	})

	if err != nil {
		logger.Log.Errorf(err.Error())
		logger.Log.Warnf("Ошибка при загрузке конфигурации, используется конфигурация по умолчанию")
	}
	logger.Log.Infof("Загружена конфигурация: %s", config.Name)
}
