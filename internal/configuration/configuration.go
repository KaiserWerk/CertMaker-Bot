package configuration

import (
	"os"
	"path/filepath"
	"time"

	"github.com/KaiserWerk/CertMaker-Bot/internal/assets"
	"github.com/KaiserWerk/CertMaker-Bot/internal/helper"

	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	App       App       `yaml:"app"`
	CertMaker CertMaker `yaml:"certmaker"`
}

type (
	App struct {
		MinValidity time.Duration `yaml:"min_validity"`
		Interval    time.Duration `yaml:"interval"`
	}
	CertMaker struct {
		Host       string `yaml:"host"`
		SkipVerify bool   `yaml:"skip_verify"`
		ApiKey     string `yaml:"apikey"`
	}
)

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

		if err := os.WriteFile(file, cont, 0700); err != nil {
			return nil, false, err
		}
		created = true
	}

	cont, err := os.ReadFile(file)
	if err != nil {
		return nil, created, err
	}

	var c AppConfig
	err = yaml.Unmarshal(cont, &c)

	return &c, created, err
}
