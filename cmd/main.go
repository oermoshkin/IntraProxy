package main

import "log"

var Config *ConfigStruct

func main() {
	var err error
	Config, err = LoadConfig("config.yml")
	if err != nil {
		log.Fatalln("Error load config file:", err)
	}

	MyServer()
}
