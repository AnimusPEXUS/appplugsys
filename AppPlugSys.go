package appplugsys

import (
	"errors"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
)

type DBPluginInfo struct {
	Name        string `gorm:"primary_key"`
	Builtin     bool
	Sha512      string
	Enabled     bool
	LastDBReKey *time.Time
	DBKey       string
}

func (DBPluginInfo) TableName() string {
	return "plugin_info"
}

type AppPlugSys struct {
	db *gorm.DB

	passthrough_data interface{}

	plugin_searcher PluginSearcherI
	plugin_opener   PluginOpenerI

	// builtin safe trusted plugins
	internal_plugins map[string]*PluginWrap

	// heterogeneous plugins
	external_plugins map[string]*PluginWrap

	lock *sync.Mutex
}

func NewAppPlugSys(
	db *gorm.DB,
	plugin_searcher PluginSearcherI,
	plugin_opener PluginOpenerI,
	passthrough_data interface{},
) (*AppPlugSys, error) {
	self := new(AppPlugSys)

	self.db = db
	self.plugin_searcher = plugin_searcher

	self.lock = &sync.Mutex{}

	self.internal_plugins = make(map[string]*PluginWrap)

	self.external_plugins = make(map[string]*PluginWrap)

	if !db.HasTable(&DBPluginInfo{}) {
		if err := db.CreateTable(&DBPluginInfo{}).Error; err != nil {
			return nil, err
		}
	}

	return self, nil
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
	if plugin_status.Builtin {
		res, err := self.plugin_searcher.FindBuiltIn(plugin_status.Name)
		if err != nil {
			return err
		}

		pw, err := NewPluginWrapFromSearchItem(res, self)
		if err != nil {
			return err
		}

		pw.Plugin.Init(pw.AppPlugSysIface())

		self.internal_plugins[res.Name] = pw
	} else {
		res, err := self.plugin_searcher.FindBySha512(plugin_status.Sha512)
		if err != nil {
			return err
		}

		pw, err := NewPluginWrapFromSearchItem(res, self)
		if err != nil {
			return err
		}

		pw.Plugin.Init(pw.AppPlugSysIface())

		self.external_plugins[res.Name] = pw
	}

	return nil
}

func (self *AppPlugSys) AcceptInternalPlugin(plugwrap *PluginWrap) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.acceptInternalPlugin(plugwrap)
}

func (self *AppPlugSys) acceptInternalPlugin(plugwrap *PluginWrap) error {

	if plugwrap == nil {
		return errors.New("plugwrap must be not nil")
	}

	err := self.removeExternalPlugin(plugwrap.Name)
	if err != nil {
		return err
	}

	self.internal_plugins[plugwrap.Name] = plugwrap

	return nil
}

func (self *AppPlugSys) RemoveInternalPlugin(unique_name string) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.removeInternalPlugin(unique_name)
}

func (self *AppPlugSys) removeInternalPlugin(unique_name string) error {
	plugwrap, ok := self.internal_plugins[unique_name]
	if !ok {
		return nil
	}

	if plugwrap.Plugin.Worker != nil {
		plugwrap.Plugin.Worker.Stop()
	}

	delete(self.internal_plugins, unique_name)

	return nil
}

// leave name empty to automatically discover it
// leave path empty to automatically discover it
func (self *AppPlugSys) AcceptExternalPlugin(sha512 string, name string, path string) {

	if sha512 != "" && name != "" {
		ps := &DBPluginInfo{
			Name:        name,
			Builtin:     false,
			Sha512:      sha512,
			Enabled:     true,
			LastDBReKey: nil,
			DBKey:       "",
		}
		self.db.Create(ps)
	}
	return
}

func (self *AppPlugSys) RemoveExternalPlugin(name string) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.removeExternalPlugin(name)
}

func (self *AppPlugSys) removeExternalPlugin(name string) error {
	plugwrap, ok := self.external_plugins[name]
	if !ok {
		return nil
	}

	if plugwrap.Plugin.Worker != nil {
		plugwrap.Plugin.Worker.Stop()
	}

	delete(self.external_plugins, name)

	return nil
}
