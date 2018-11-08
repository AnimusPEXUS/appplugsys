package appplugsys

import (
	"errors"
	"sync"
	"time"

	"github.com/AnimusPEXUS/utils/worker/workerstatus"
	"github.com/jinzhu/gorm"
)

type DBPluginInfo struct {
	Name        string `gorm:"primary_key"`
	BuiltIn     bool
	Sha512      string
	Enabled     bool
	LastDBReKey *time.Time
	Key         string
}

func (DBPluginInfo) TableName() string {
	return "plugin_info"
}

type AppPlugSysStatusDisplayLine struct {
	Sha512  string
	BuiltIn bool
	Found   bool

	Enabled   bool
	AutoStart bool

	WorkerStatus *workerstatus.WorkerStatus // nil if not worker
}

type AppPlugSys struct {
	db *gorm.DB

	passthrough_data interface{}

	plugin_searcher PluginSearcherI

	plugins map[string]*PluginWrap

	getPluginDB func(*DBPluginInfo) (*gorm.DB, error)

	lock *sync.Mutex
}

func NewAppPlugSys(
	db *gorm.DB,

	plugin_searcher PluginSearcherI, // to find already accepted plugin
	//	plugin_opener PluginOpenerI, // to confirm and open external plugin
	//	plugin_acceptor PluginAcceptorI, // confirm acception of any plugin

	passthrough_data interface{},
) (*AppPlugSys, error) {
	self := new(AppPlugSys)

	self.db = db
	self.plugin_searcher = plugin_searcher
	//	self.plugin_opener = plugin_opener

	self.lock = &sync.Mutex{}

	self.plugins = make(map[string]*PluginWrap)

	if !db.HasTable(&DBPluginInfo{}) {
		if err := db.CreateTable(&DBPluginInfo{}).Error; err != nil {
			return nil, err
		}
	}

	return self, nil
}

func (self *AppPlugSys) PluginInfoTable() {

}

func (self *AppPlugSys) GetPluginByName(name string) (ret *PluginWrap, err error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.getPluginByName(name)
}

func (self *AppPlugSys) getPluginByName(name string) (ret *PluginWrap, err error) {

	ret = nil
	err = errors.New("not found")

	for k, v := range self.plugins {
		if k == name {
			ret = v
			return
		}
	}

	return
}

func (self *AppPlugSys) internalMethodToLoadAllPlugins() error {

	plugin_statuses := make([]*DBPluginInfo, 0)

	err := self.db.Find(&plugin_statuses).Error
	if err != nil {
		return err
	}

	errors := make([]error, 0)
	for _, i := range plugin_statuses {
		res := self.internalMethodToLoadPlugin(i)
		if res != nil {
			errors = append(errors, res)
		}
	}

	return nil
}

func (self *AppPlugSys) internalMethodToLoadPlugin(plugin_status *DBPluginInfo) error {

	var err error
	var res *PluginSearcherSearchResultItem

	if plugin_status.BuiltIn {
		res, err = self.plugin_searcher.FindBuiltIn(plugin_status.Name)
		if err != nil {
			return err
		}

	} else {
		res, err = self.plugin_searcher.FindBySha512(plugin_status.Sha512)
		if err != nil {
			return err
		}
	}

	pw, err := NewPluginWrapFromSearchItem(res, self)
	if err != nil {
		return err
	}

	pw.Plugin.Init(pw.AppPlugSysIface())

	self.plugins[res.Name] = pw

	return nil
}

func (self *AppPlugSys) AcceptPlugin(plugwrap *PluginWrap) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.acceptPlugin(plugwrap, false)
}

func (self *AppPlugSys) acceptPlugin(plugwrap *PluginWrap, no_register bool) error {

	if plugwrap == nil {
		return errors.New("plugwrap must be not nil")
	}

	_, ok := self.plugins[plugwrap.Name]
	if ok {
		return errors.New("already have accepted plugin with this name")
	}

	self.plugins[plugwrap.Name] = plugwrap

	if !no_register {

		prec := &DBPluginInfo{
			Name:        plugwrap.Name,
			BuiltIn:     plugwrap.BuiltIn,
			Sha512:      plugwrap.Sha512,
			Enabled:     false,
			LastDBReKey: nil,
			Key:         "",
		}

		err := self.db.Create(&prec).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (self *AppPlugSys) RemovePlugin(name string) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.removePlugin(name)
}

func (self *AppPlugSys) removePlugin(name string) error {

	pw, ok := self.plugins[name]
	if ok {
		if pw.Plugin.Worker != nil {
			pw.Plugin.Worker.Stop()
		}
		delete(self.plugins, name)
	}

	err := self.db.Where("name = ?", name).Delete(&DBPluginInfo{}).Error
	if err != nil {
		return err
	}

	return nil
}
