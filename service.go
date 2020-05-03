package main

import (
	"io/ioutil"
	"log"
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

func loadFromFile(path string) Forward {

	filename, _ := filepath.Abs(path)
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Panic(err.Error())
	}

	var forward Forward
	err = yaml.Unmarshal(yamlFile, &forward)
	if err != nil {
		log.Panic(err.Error())
	}
	return forward
}
