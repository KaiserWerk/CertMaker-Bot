package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var configFile = "config.yml"

func setConfigurationFile(f string) {
	configFile = f
}

func getConfiguration() (*configuration, error) {
	cont, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var c *configuration
	err = yaml.Unmarshal(cont, c)
	if err != nil {
		return nil, err
	}

	return c, err
}
