package configuration

import (
	"github.com/KaiserWerk/CertMaker-Bot/internal/assets"
	"github.com/KaiserWerk/CertMaker-Bot/internal/entity"
	helper2 "github.com/KaiserWerk/CertMaker-Bot/internal/helper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
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
	if err != nil {
		return nil, err
	}

	return &c, err
}

func SetupConfiguration() (*entity.Configuration, error) {
	if !helper2.FileExists(configFile) {
		configAssets := assets.GetConfigurationAssets()
		cont, err := configAssets.ReadFile("assets/config.dist.yaml")
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
