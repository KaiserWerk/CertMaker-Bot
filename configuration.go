package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

var configFile = "config.yaml"

func setConfigurationFile(f string) {
	configFile = f
}

func getConfiguration() (*configuration, error) {
	cont, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var c configuration
	err = yaml.Unmarshal(cont, &c)
	if err != nil {
		return nil, err
	}

	return &c, err
}

func setupConfiguration() (*configuration, error) {
	if !fileExists(configFile) {
		cont, err := assetsFS.ReadFile("assets/config.dist.yaml")
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

	c, err := getConfiguration()
	if err != nil {
		return nil, err
	}

	return c, nil
}
