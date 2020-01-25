package api

import (
	"dnsManager/config"
	"dnsManager/db"
	"dnsManager/models"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strings"
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

func (a *API) GetDescriptor(_ string, result *string) error {
	log.Println("Get Descriptor Called______")

	val := config.LoadConfig().Server.Descriptor

	if val == "." {
		*result = ""
	} else {
		client := ParentClient()
		var r string
		err := client.Call("API.GetDescriptor", "", r)
		if err != nil {
			log.Println("Could not get parent descriptor")
			r = ""
		}
		*result = "." + val + r
	}

	log.Println("______Get Descriptor Returning")
	return nil
}

func (a *API) FindAll(_ string, result *[]models.DomainName) error {
	log.Println("Find All Called______")

	database := db.GetOpenDatabase()
	domains := models.DomainName{}.FindAll(database)
	*result = domains
	defer database.Close()

	log.Println("______Find All Returning")
	return nil
}

func (a *API) Lookup(name string, result *[]models.DomainName) error {
	log.Println("Lookup Called______")

	descriptor := GetMyDescriptor()
	log.Println("Server descriptor:", descriptor)

	cacheDatabase := db.GetOpenCacheDatabase()
	cacheResults := models.DomainName{}.FindByName(cacheDatabase, name)

	if len(cacheResults) > 0 {
		log.Println("Using Cache...")
		*result = cacheResults
		models.DomainName{}.SaveAll(cacheDatabase, cacheResults)
		defer cacheDatabase.Close()

	} else {
		if strings.HasSuffix(name, descriptor) {
			foundChild := false
			for _, node := range ChildrenNodes {
				if strings.HasSuffix(name, node.Descriptor) {
					client := GetClient(node)
					var r []models.DomainName
					err := client.Call("API.Lookup", name, &r)
					if err != nil {
						log.Println("______Returning Error")
						return err
					}
					log.Println("Using Child Server...")
					*result = r
					if len(r) > 0 {
						log.Println("Saving to Cache...")
						models.DomainName{}.SaveAll(cacheDatabase, r)
					}
					foundChild = true
					break
				}
			}

			if !foundChild {
				localDatabase := db.GetOpenDatabase()
				localResults := models.DomainName{}.FindByName(localDatabase, name)
				*result = localResults
				models.DomainName{}.SaveAll(localDatabase, localResults)
				log.Println("Using Local Database...")
				defer localDatabase.Close()
			}
		} else {
			client := ParentClient()
			var r []models.DomainName
			err := client.Call("API.Lookup", name, &r)
			if err != nil {
				log.Println("______Returning Error")
				return err
			}
			log.Println("Using Parent Server...")
			*result = r
			if len(r) > 0 {
				log.Println("Saving to Cache...")
				models.DomainName{}.SaveAll(cacheDatabase, r)
			}
		}
	}

	log.Println("______Lookup Returning")
	return nil
}

func (a *API) FindByName(name string, result *[]models.DomainName) error {
	log.Println("Find By Name Called______")

	database := db.GetOpenDatabase()
	domains := models.DomainName{}.FindByName(database, name)
	fmt.Println(domains)
	*result = domains
	defer database.Close()

	log.Println("______Find By Name Returning")
	return nil
}

func (a *API) RegisterChild(child models.ServerNode, result *bool) error {
	log.Println("Register Child Called______")

	ChildrenNodes = append(ChildrenNodes, child)
	MoveUnfittingData(GetClient(child))
	*result = true

	log.Println("______Register Child Returning")
	return nil
}

func (a *API) Register(domain models.DomainName, result *bool) error {
	log.Println("Register Called______")

	name := domain.Name
	descriptor := GetMyDescriptor()
	log.Printf("Server descriptor: %s", descriptor)

	if strings.HasSuffix(name, descriptor) {
		foundChild := false
		for _, node := range ChildrenNodes {
			if strings.HasSuffix(name, node.Descriptor) {
				client := GetClient(node)
				var r bool
				err := client.Call("API.Register", domain, &r)
				if err != nil {
					log.Println("______Returning Error")
					return err
				}
				log.Println("Using Child Server...")
				*result = r
				foundChild = true
				break
			}
		}

		if !foundChild {
			localDatabase := db.GetOpenDatabase()
			domain.Save(localDatabase)
			*result = true
			log.Println("Using Local Database...")
			defer localDatabase.Close()
		}
	} else {
		client := ParentClient()
		var r bool
		err := client.Call("API.Register", domain, &r)
		if err != nil {
			return err
		}
		log.Println("Using Parent Server...")
		*result = r
	}

	log.Println("______Register Returning")
	return nil
}

func (a *API) Remove(domain models.DomainName, result *bool) error {
	log.Println("Remove Called______")

	name := domain.Name
	descriptor := GetMyDescriptor()
	log.Println("Server descriptor:", descriptor)

	if strings.HasSuffix(name, descriptor) {
		foundChild := false
		for _, node := range ChildrenNodes {
			if strings.HasSuffix(name, node.Descriptor) {
				client := GetClient(node)
				var r bool
				err := client.Call("API.Remove", domain, &r)
				if err != nil {
					log.Println("______Returning Error")
					return err
				}
				log.Println("Using Child Server...")
				*result = r
				foundChild = true
				break
			}
		}

		if !foundChild {
			localDatabase := db.GetOpenDatabase()
			domain.Delete(localDatabase)
			*result = true
			log.Println("Using Local Database...")
			defer localDatabase.Close()
		}
	} else {
		client := ParentClient()
		var r bool
		err := client.Call("API.Remove", domain, &r)
		if err != nil {
			log.Println("______Returning Error")
			return err
		}
		log.Println("Using Parent Server...")
		*result = r
	}

	log.Println("______Remove Returning")
	return nil
}

func (a *API) RemoveCache(domain models.DomainName, result *bool) error {
	log.Println("Remove Cache Called______")

	cacheDatabase := db.GetOpenCacheDatabase()
	domain.Delete(cacheDatabase)
	*result = true
	defer cacheDatabase.Close()

	log.Println("______Remove Cache Returning")
	return nil
}

func (a *API) RemoveChild(child models.ServerNode, result *bool) error {
	log.Println("Remove Child Called______")

	delIndex := -1
	for i, node := range ChildrenNodes {
		if node.Address == child.Address && node.Port == child.Port {
			delIndex = i
		}
	}

	if delIndex < 0 {
		log.Println("______Remove Child Found Nothing")
		*result = false
		return nil
	}

	ChildrenNodes[delIndex] = ChildrenNodes[len(ChildrenNodes)-1]
	ChildrenNodes[len(ChildrenNodes)-1] = models.ServerNode{}
	ChildrenNodes = ChildrenNodes[:len(ChildrenNodes)-1]
	*result = true

	log.Println("______Remove Child Returning")
	return nil
}

func (a *API) SwitchParent(parent models.ServerNode, result *bool) error {
	log.Println("Switch Parent Called______")

	configuration := config.LoadConfig()
	configuration.Parent.Address = parent.Address
	configuration.Parent.Port = parent.Port
	config.SaveConfig(configuration)

	InformParent()
	MoveUnfittingData(ParentClient())
	*result = true

	log.Println("______Switch Parent Returning")
	return nil
}
