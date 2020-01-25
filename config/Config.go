package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Timeout int `yaml:"timeout"`
}

func LoadConfig() Config {
	workingDirectory, _ := os.Getwd()
	f, err := os.Open(workingDirectory + "/config/config.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func SaveConfig(cfg Config) {
	workingDirectory, _ := os.Getwd()
	f, err := os.OpenFile(workingDirectory+"/config/config.yml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	//x := io.Writer(f)

	encoder := yaml.NewEncoder(f)
	err = encoder.Encode(cfg)
	if err != nil {
		fmt.Println("here")
		log.Fatal(err)
	}
	defer encoder.Close()
}
