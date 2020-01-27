package models

import (
	"dnsManager/db"
	"github.com/jinzhu/gorm"
	"time"
)

type ServerNode struct {
	Address    string
	Port       string
	Descriptor string
}

type ServerModel struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Address       string
	Port          string
	ParentAddress string
	ParentPort    string
}

func (model *ServerModel) Save(database *gorm.DB) {
	var prev ServerModel
	database.Where("address = ? and port = ?", model.Address, model.Port).First(&prev)

	if model.ID == 0 && prev.ID == 0 {
		database.Create(model)
	} else if model.ID > 0 {
		database.First(&prev, model.ID)
		database.Model(&prev).Update(model)
	} else if prev.ID > 0 {
		database.Model(&prev).Update(model)
	}
}

func (ServerModel) SaveAll(database *gorm.DB, models []ServerModel) {
	for _, model := range models {
		d := ServerModel{Address: model.Address, Port: model.Port, ParentAddress: model.ParentAddress, ParentPort: model.ParentPort}
		d.Save(database)
	}
}

func (ServerModel) FindAll(database *gorm.DB) []ServerModel {
	var models []ServerModel
	database.Find(&models)
	return models
}

func (ServerModel) FindById(database *gorm.DB, id int) ServerModel {
	var model ServerModel
	database.First(&model, id)
	return model
}

func (ServerModel) FindByServerNode(database *gorm.DB, serverNode ServerNode) ServerModel {
	var model ServerModel
	database.First(&model, "address = ? and port = ?", serverNode.Address, serverNode.Port)
	return model
}

func (model *ServerModel) ToServerNode() ServerNode {
	return ServerNode{Address: model.Address, Port: model.Port}
}

func (model *ServerModel) GetParent() ServerNode {
	return ServerNode{Address: model.ParentAddress, Port: model.ParentPort}
}
func (model *ServerModel) GetChildren(database *gorm.DB) []ServerModel {
	var models []ServerModel
	database.Where("parent_address = ? AND parent_port = ?", model.Address, model.Port).Find(&models)
	return models
}

func (model *ServerModel) Delete(database *gorm.DB) {
	database.Where("address = ? and port = ?", model.Address, model.Port).Delete(ServerModel{})
}

func (ServerModel) Migrate() {
	database := db.GetOpenDatabase()
	database.AutoMigrate(&ServerModel{})
	defer database.Close()
}
