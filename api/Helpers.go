package api

import (
	"dnsManager/db"
	"dnsManager/models"
	"github.com/jinzhu/gorm"
	"log"
	"net"
	"net/rpc"
	"time"
)

func GetClient(node models.ServerNode) (*rpc.Client, error) {
	address := node.Address
	port := node.Port

	timeout := 3 * time.Second
	_, err := net.DialTimeout("tcp", address+":"+port, timeout)
	if err != nil {
		//log.Println("Site unreachable, error: ", err)
		return nil, err
	}

	client, err := rpc.DialHTTP("tcp", address+":"+port)
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
	timeout := 5 * time.Second
	_, err := net.DialTimeout("tcp", node.Address+":"+node.Port, timeout)
	if err != nil {
		if first {
			log.Println("Suspecting node", node.ToServerNode())
			time.Sleep(3 * time.Second)
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
	parent := node.GetParent()
	children := node.GetChildren(database)
	var r bool

	client, err := GetClient(parent)
	if err != nil {
		log.Printf("Could not contact parent server %s", parent)
		return
	}
	err = client.Call("API.RemoveChild", node.ToServerNode(), &r)
	if err != nil {
		log.Printf("Could not call Remove Child %s", parent)
		return
	}

	for _, child := range children {
		client, err := GetClient(child.ToServerNode())
		if err != nil {
			log.Printf("Could not contact child server %s", child.ToServerNode())
			continue
		}
		log.Println("New Parent is", parent)
		err = client.Call("API.SwitchParent", parent, &r)
		if err != nil {
			log.Printf("Could not call Switch Parent for child %s", child.ToServerNode())
		}
	}
}

func RemoveServerCache(server models.ServerModel, domain models.DomainName) {
	var r bool
	client, err := GetClient(server.ToServerNode())
	if err != nil {
		log.Printf("Could not contact server %s", server.ToServerNode())
		return
	}
	err = client.Call("API.RemoveCache", domain, &r)
	if err != nil {
		log.Printf("Could not call Remove Cache %s", server.ToServerNode())
	}
}
