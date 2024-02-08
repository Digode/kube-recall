package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func (config *Config) file() error {
	file, err := ioutil.ReadFile("../../config.yaml")
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	return nil
}
