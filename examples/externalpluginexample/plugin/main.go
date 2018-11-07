package main

import (
	"github.com/AnimusPEXUS/appplugsys"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	db *gorm.DB
}

func NewController() *Controller {
	self := new(Controller)
	return self
}

func (self *Controller) SetDB(db *gorm.DB) {
	self.db = db
	return
}

func GetPlugin() *appplugsys.Plugin {

	c := NewController()

	ret := &appplugsys.Plugin{
		SetDB: c.SetDB,
	}
	return ret
}

func main() {

}
