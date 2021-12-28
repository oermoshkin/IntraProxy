package main

import (
	"gopkg.in/yaml.v2"
	"os"
)

type ConfigStruct struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`

	Proxy struct {
		Server string `yaml:"server"`
		Login  string `yaml:"login"`
		ApiGW  string `yaml:"apigw"`
		Doc    string `yaml:"doc"`
	} `yaml:"proxy"`

	Origin struct {
		Server string `yaml:"server"`
		Login  string `yaml:"login"`
		ApiGW  string `yaml:"apigw"`
		Doc    string `yaml:"doc"`
	} `yaml:"origin"`
}

//LoadConfig Загружаем конфиг
func LoadConfig(configPath string) (*ConfigStruct, error) {
	config := &ConfigStruct{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
