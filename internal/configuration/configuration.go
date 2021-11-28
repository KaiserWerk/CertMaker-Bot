package configuration

import (
	"github.com/KaiserWerk/CertMaker-Bot/internal/assets"
	"github.com/KaiserWerk/CertMaker-Bot/internal/entity"
	helper "github.com/KaiserWerk/CertMaker-Bot/internal/helper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

var configFile = "config.yaml"

func SetConfigurationFile(f string) {
	configFile = f
}

func GetConfiguration() (*entity.Configuration, error) {
	cont, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var c entity.Configuration
	err = yaml.Unmarshal(cont, &c)
	return &c, err
}

func SetupConfiguration() (*entity.Configuration, error) {
	if !helper.FileExists(configFile) {
		configPath := filepath.Dir(configFile)
		_ = os.MkdirAll(configPath, 0755)

		configAssets := assets.GetConfigurationAssets()
		cont, err := configAssets.ReadFile("config/config.dist.yaml")
		if err != nil {
			return nil, err
		}

		fh, err := os.Create(configFile)
		if err != nil {
			return nil, err
		}

		_, err = fh.Write(cont)
		if err != nil {
			return nil, err
		}
	}

	c, err := GetConfiguration()
	if err != nil {
		return nil, err
	}

	return c, nil
}
