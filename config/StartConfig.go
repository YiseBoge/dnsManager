package config

import (
	"dnsManager/models"
)

func Start() {
	models.ServerModel{}.Migrate()
}
