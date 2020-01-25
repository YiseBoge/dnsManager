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

func GetClient(node models.ServerModel) (*rpc.Client, error) {
	parentAddress := node.Address
	parentPort := node.Port

	client, err := rpc.DialHTTP("tcp", parentAddress+":"+parentPort)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func KeepHealthy() {
	log.Println("Managing Health...")
	database := db.GetOpenDatabase()
	servers := models.ServerModel{}.FindAll(database)

	for _, server := range servers {
		go CheckPulse(database, server, true)
	}
}

func CheckPulse(database *gorm.DB, node models.ServerModel, first bool) {
	client, err := GetClient(node)
	var r bool

	if err != nil || client.Call("API.Heartbeat", "", &r) != nil {
		if first {
			log.Println("Suspecting node", node.ToServerNode())
			t := config.LoadConfig().Timeout / 2
			time.Sleep(time.Duration(t) * time.Second)
			CheckPulse(database, node, false)
		} else {
			log.Println("Removing node", node.ToServerNode())
			GiveChildrenToParent(database, node)
			node.Delete(database)
		}
	} else {
		log.Println("Healthy node", node.ToServerNode())
	}
}

func GiveChildrenToParent(database *gorm.DB, node models.ServerModel) {
	parent := node.GetParent(database)
	children := node.GetChildren(database)
	var r bool

	client, err := GetClient(parent)
	if err != nil {
		log.Printf("Could not contact parent server %s", parent.ToServerNode())
		return
	}
	err = client.Call("API.RemoveChild", node.ToServerNode(), &r)
	if err != nil {
		log.Printf("Could not call Remove Child %s", parent.ToServerNode())
		return
	}

	for _, child := range children {
		client, err := GetClient(child)
		if err != nil {
			log.Printf("Could not contact child server %s", child.ToServerNode())
			continue
		}
		err = client.Call("API.SwitchParent", child.ToServerNode(), &r)
		if err != nil {
			log.Printf("Could not call Switch Parent for child %s", child.ToServerNode())
		}
	}
}

func RemoveServerCache(server models.ServerModel, domain models.DomainName) {
	var r bool
	client, err := GetClient(server)
	if err != nil {
		log.Printf("Could not contact server %s", server.ToServerNode())
		return
	}
	err = client.Call("API.RemoveCache", domain, &r)
	if err != nil {
		log.Printf("Could not call Remove Cache %s", server.ToServerNode())
	}
}
