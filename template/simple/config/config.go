package config

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/fengjx/go-halo/fskit"
	"github.com/fengjx/luchen"
	"github.com/fengjx/luchen/env"
	"github.com/fengjx/luchen/log"
	"go.uber.org/zap"
)

var appConfig AppConfig

func init() {
	configArg := flag.String("c", "", "custom config file path")
	flag.Parse()
	appCfg := "conf/app.yml"
	if !env.IsDev() {
		exePath, err := os.Executable()
		if err != nil {
			log.Panic("can not get exec file path", zap.Error(err))
		}
		appCfg = filepath.Join(filepath.Dir(exePath), "conf/app.yml")
	}
	log.Infof("app conf: %s", appCfg)
	appConfigFile, err := fskit.Lookup(appCfg, 5)
	if err != nil {
		log.Panic("config file not found", zap.String("path", appCfg), zap.Error(err))
	}
	configs := []string{appConfigFile}
	var configFile string
	if configArg != nil {
		configFile = *configArg
	}
	if configFile == "" {
		envConfig := os.Getenv("APP_CONFIG")
		if envConfig != "" {
			configFile = envConfig
		}
	}
	if configFile != "" {
		configFile, err = fskit.Lookup(configFile, 5)
		if err != nil {
			log.Panic("config file not found", zap.String("path", configFile), zap.Error(err))
		}
		log.Infof("load custom config file: %s", configFile)
		configs = append(configs, configFile)
	}
	appConfig = luchen.MustLoadConfig[AppConfig](configs...)
}

type AppConfig struct {
	Server Server `json:"server"`
}

type Server struct {
	HTTP HTTPServerConfig
}

// CorsConfig 跨域配置
type CorsConfig struct {
	AllowOrigins []string `json:"allow-origins"`
}

type HTTPServerConfig struct {
	ServerName string     `json:"server-name"`
	Listen     string     `json:"listen"`
	Cors       CorsConfig `json:"cors"`
}

func GetConfig() AppConfig {
	return appConfig
}
