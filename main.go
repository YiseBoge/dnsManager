package main

import (
	"dnsManager/api"
	"dnsManager/config"
	"fmt"
	"gopkg.in/robfig/cron.v3"
	"log"
	"regexp"
	"strconv"
	"time"
)

func main() {
	config.Start()
	fmt.Println("Welcome to the DomaInator Manager.")
	configuration := config.LoadConfig()
	portRegex, _ := regexp.Compile("^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])(?::([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5]))?$")

	var res1 string
	for true {
		fmt.Printf("Current port = \"%s\" press 'Enter' to continue or provide new port: ", configuration.Server.Port)
		_, _ = fmt.Scanln(&res1)

		if res1 == "" {
			break
		}

		if portRegex.MatchString(res1) {
			configuration.Server.Port = res1
			break
		}
		fmt.Println("**Bad input, Please try again**")
	}

	var res2 string
	for true {
		fmt.Printf("Timeout value = \"%d\" press 'Enter' to continue or provide new timeout: ", configuration.Timeout)
		_, _ = fmt.Scanln(&res2)

		if res2 == "" {
			break
		}

		v, err := strconv.Atoi(res2)
		if err == nil {
			configuration.Timeout = v
			break
		}
		fmt.Println("**Bad input, Please try again**")
	}

	config.SaveConfig(configuration)
	log.Printf("Port set to: %s", configuration.Server.Port)

	c := cron.New(cron.WithSeconds())
	timeString := fmt.Sprintf("@every %ds", configuration.Timeout)
	_, _ = c.AddFunc(timeString, func() {
		api.KeepHealthy()
	})
	c.Start()

	go api.Serve()
	time.Sleep(1 * time.Second)

	var res string
	for true {
		fmt.Println("Type 'exit' or 'stop' to stop serving.")
		_, _ = fmt.Scanln(&res)

		if res == "exit" || res == "stop" {
			break
		}
		fmt.Println("**Bad input, Please try again**")
	}
}
