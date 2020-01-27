package api

import (
	"dnsManager/config"
	"dnsManager/db"
	"dnsManager/models"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type API int

func Serve() {
	serverPort := config.LoadConfig().Server.Port

	api := new(API)
	err := rpc.Register(api)
	if err != nil {
		log.Fatal("Error Registering API", err)
	}

	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", ":"+serverPort)

	if err != nil {
		log.Fatal("Listener Error", err)
	}
	log.Printf("Serving RPC on port \"%s\"", serverPort)
	err = http.Serve(listener, nil)

	if err != nil {
		log.Fatal("Error Serving: ", err)
	}
}

func (a *API) RemoveFromAllCache(domain models.DomainName, result *bool) error {
	log.Println("Remove From All Cache Called______")

	database := db.GetOpenDatabase()
	allServers := models.ServerModel{}.FindAll(database)

	for _, server := range allServers {
		go RemoveServerCache(server, domain)
	}
	*result = true

	log.Println("______Remove From All Cache Returning")
	return nil
}

func (a *API) RegisterServer(server models.ServerModel, result *bool) error {
	log.Println("Register Server Called______")

	database := db.GetOpenDatabase()
	server.Save(database)
	*result = true
	log.Println("New Server", server.ToServerNode())
	node := server.GetParent()
	log.Println("It's parent", node)

	log.Println("______Register Server Returning")
	return nil
}
