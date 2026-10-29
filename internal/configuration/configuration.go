package configuration

import (
	"errors"
	"os"
	"time"

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
		ApiKey     string `yaml:"api_key"`
	}
)

func defaults() AppConfig {
	return AppConfig{
		App: App{
			MinValidity: 72 * time.Hour,
			Interval:    6 * time.Hour,
		},
		CertMaker: CertMaker{
			Host:       "http://certmaker:8880",
			SkipVerify: false,
			ApiKey:     "",
		},
	}
}

func Setup(file string) (*AppConfig, error) {
	c := defaults()
	st, err := os.Stat(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &c, nil
		}
		return nil, err
	}
	if st.IsDir() {
		return nil, errors.New("configuration path points to a directory")
	}

	cont, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(cont, &c)

	return &c, err
}
