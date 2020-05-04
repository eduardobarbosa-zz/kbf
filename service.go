package main

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Forward struct {
	Services []struct {
		Name       string `yaml:"name"`
		Namespace  string `yaml:"namespace"`
		Port       int    `yaml:"port"`
		TargetPort int    `yaml:"targetPort"`
	} `yaml:"services"`
}

func loadFromFile(path string) (Forward, error) {

	filename, _ := filepath.Abs(path)
	yamlFile, err := ioutil.ReadFile(filename)
	var forward Forward

	if err != nil {
		return forward, errors.New(err.Error())
	}

	err = yaml.Unmarshal(yamlFile, &forward)
	if err != nil {
		return forward, errors.New(err.Error())
	}
	return forward, nil
}
