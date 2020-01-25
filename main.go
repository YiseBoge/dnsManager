package main

import (
	"dnsManager/api"
	"dnsManager/config"
	"dnsManager/models"
	"fmt"
	"gopkg.in/robfig/cron.v3"
	"log"
	"regexp"
	"time"
)

func main() {
	fmt.Println("Welcome to the DomaInator Manager.")
	configuration := config.LoadConfig()
	portRegex, _ := regexp.Compile("^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])(?::([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5]))?$")

	var res1 string
	for true {
		log.Printf("Current port = \"%s\" press 'Enter' to continue or provide new port:", configuration.Server.Port)
		_, _ = fmt.Scanln(&res1)

		if res1 == "" {
			break
		}

		if portRegex.MatchString(res1) {
			configuration.Server.Port = res1
			break
		}
		log.Printf("**Bad input, Please try again**")
	}

	config.SaveConfig(configuration)
	log.Printf("Port set to: %s", configuration.Server.Port)

	c := cron.New(cron.WithSeconds())
	timeString := fmt.Sprintf("@every %dh", configuration.Timeout)
	_, _ = c.AddFunc(timeString, func() {
		api.KeepHealthy()
	})
	c.Start()

	go api.Serve()

	time.Sleep(1 * time.Second)

	var res string
	for true {
		log.Printf("Type 'exit' or 'stop' to stop serving.")
		_, _ = fmt.Scanln(&res)

		if res == "exit" || res == "stop" {
			break
		}
		log.Printf("**Bad input, Please try again**")
	}
}
