package configuration

import (
	"github.com/KaiserWerk/CertMaker-Bot/internal/assets"
	helper "github.com/KaiserWerk/CertMaker-Bot/internal/helper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type AppConfig struct {
	App struct {
		MinValidity time.Duration `yaml:"min_validity"`
		Interval    time.Duration `yaml:"interval"`
	} `yaml:"app"`
	CertMaker struct {
		Host       string `yaml:"host"`
		SkipVerify bool   `yaml:"skip_verify"`
		ApiKey     string `yaml:"apikey"`
	} `yaml:"certmaker"`
}

func Setup(file string) (*AppConfig, bool, error) {
	var created bool
	if !helper.FileExists(file) {
		if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
			return nil, false, err
		}

		cont, err := assets.GetConfigAsset("config.dist.yaml")
		if err != nil {
			return nil, false, err
		}

		if err := ioutil.WriteFile(file, cont, 0700); err != nil {
			return nil, false, err
		}
		created = true
	}

	cont, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, created, err
	}

	var c AppConfig
	err = yaml.Unmarshal(cont, &c)

	return &c, created, nil
}
