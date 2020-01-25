package api

import (
	"dnsManager/config"
	"dnsManager/db"
	"dnsManager/models"
	"github.com/jinzhu/gorm"
	"log"
	"net/rpc"
	"time"
)

func GetClient(node models.ServerModel) *rpc.Client {
	parentAddress := node.Address
	parentPort := node.Port

	client, err := rpc.DialHTTP("tcp", parentAddress+":"+parentPort)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func CheckHeartbeat(node models.ServerModel, first bool) {
	client := GetClient(node)
	database := db.GetOpenDatabase()
	var r bool
	if first && client.Call("API.Heartbeat", "", &r) != nil {
		t := config.LoadConfig().Timeout
		time.Sleep(time.Duration(t) * time.Minute)
		CheckHeartbeat(node, false)
	}

	if !r {
		GiveChildrenToParent(database, node)
		node.Delete(database)
	}
}

func GiveChildrenToParent(database *gorm.DB, node models.ServerModel) {
	parent := node.GetParent(database)
	children := node.GetChildren(database)
	var r bool

	client := GetClient(parent)
	err := client.Call("API.RemoveChild", node, &r)
	if err != nil {
		log.Println("Could not contact server", parent)
		return
	}

	for _, child := range children {
		client := GetClient(child)
		err := client.Call("API.SwitchParent", child, &r)
		if err != nil {
			log.Println("Could not contact server", parent)
		}
	}
}
