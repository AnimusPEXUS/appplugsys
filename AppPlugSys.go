package appplugsys

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
)

type DBPluginInfo struct {
	Name        string `gorm:"primary_key"`
	BuiltIn     bool
	Sha512      string
	Enabled     bool
	LastDBReKey time.Time
	Key         string
}

func (DBPluginInfo) TableName() string {
	return "plugin_info"
}

type AppPlugSysPluginDisplayItem struct {
	Name         string
	Sha512       string
	BuiltIn      bool
	Found        bool
	Enabled      bool
	AutoStart    bool
	WorkerStatus string
	LastDBReKey  time.Time
	//	NotInDB      bool
	//	DBError      bool
}

type AppPlugSysApplicationDisplayItem struct {
	Title       string
	Icon        []byte
	PluginName  string
	Name        string
	Description string
}

type AppPlugSys struct {
	db *gorm.DB

	plugin_searcher PluginSearcherI

	plugins map[string]*PluginWrap

	getPluginDB func(*DBPluginInfo) (*gorm.DB, bool, bool, error)

	lock *sync.Mutex
}

func NewAppPlugSys(
	db *gorm.DB,
	getPluginDB func(*DBPluginInfo) (*gorm.DB, bool, bool, error),
	plugin_searcher PluginSearcherI, // to find already accepted plugin
) (*AppPlugSys, error) {
	self := new(AppPlugSys)

	self.db = db
	self.plugin_searcher = plugin_searcher
	self.getPluginDB = getPluginDB
	//	self.plugin_opener = plugin_opener

	self.lock = &sync.Mutex{}

	self.plugins = make(map[string]*PluginWrap)

	if !db.HasTable(&DBPluginInfo{}) {
		if err := db.CreateTable(&DBPluginInfo{}).Error; err != nil {
			return nil, err
		}
	}

	self.internalMethodToLoadAllPlugins()

	return self, nil
}

func (self *AppPlugSys) PluginInfoTable() map[string]*AppPlugSysPluginDisplayItem {

	ret := make(map[string]*AppPlugSysPluginDisplayItem)

	for k, v := range self.plugins {

		ws := ""
		if v.Plugin.Worker != nil {
			ws = v.Plugin.Worker.Status().String()
		}

		ret[k] = &AppPlugSysPluginDisplayItem{
			Name:         v.Name,
			BuiltIn:      v.BuiltIn,
			Enabled:      v.Enabled,
			Sha512:       v.Sha512,
			WorkerStatus: ws,
			Found:        v.Plugin != nil,
			LastDBReKey:  v.LastBDReKey,
		}

	}

	return ret
}

func (self *AppPlugSys) ApplicationInfoTable() []*AppPlugSysApplicationDisplayItem {

	ret := make(map[string]*AppPlugSysPluginDisplayItem)

	for k, v := range self.plugins {

		ws := ""
		if v.Plugin.Worker != nil {
			ws = v.Plugin.Worker.Status().String()
		}

		ret[k] = &AppPlugSysPluginDisplayItem{
			Name:         v.Name,
			BuiltIn:      v.BuiltIn,
			Enabled:      v.Enabled,
			Sha512:       v.Sha512,
			WorkerStatus: ws,
			Found:        v.Plugin != nil,
			LastDBReKey:  v.LastBDReKey,
		}

	}

	return ret
}

func (self *AppPlugSys) GetPluginByName(name string) (ret *PluginWrap, err error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.getPluginByName(name)
}

func (self *AppPlugSys) getPluginByName(name string) (*PluginWrap, error) {

	ret, ok := self.plugins[name]
	if !ok {
		return nil, errors.New("not found")
	}

	return ret, nil
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

	err = self.acceptPlugin(pw, plugin_status, true)
	if err != nil {
		return err
	}

	return nil
}

func (self *AppPlugSys) AcceptPlugin(plugwrap *PluginWrap) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.acceptPlugin(plugwrap, nil, false)
}

func (self *AppPlugSys) acceptPlugin(
	plugwrap *PluginWrap,
	plugin_status *DBPluginInfo,
	no_create bool,
) error {

	if plugwrap == nil {
		return errors.New("plugwrap must be not nil")
	}

	_, ok := self.plugins[plugwrap.Name]
	if ok {
		return errors.New("already have accepted plugin with this name")
	}

	create_record := false

	if plugin_status == nil {

		plugin_status = new(DBPluginInfo)

		err := self.db.Where("name = ?", plugwrap.Name).Take(&plugin_status).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			} else {
				create_record = true

				plugin_status = &DBPluginInfo{
					Name:        plugwrap.Name,
					BuiltIn:     plugwrap.BuiltIn,
					Sha512:      plugwrap.Sha512,
					Enabled:     false,
					LastDBReKey: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
					Key:         "",
				}

			}
		}

	}

	plug_db, need_key, need_rekey, err := self.getPluginDB(plugin_status)
	if err != nil {
		return err
	}

	if need_key || need_rekey {

		buff := make([]byte, 50)
		rand.Read(buff)
		buff_str := base64.RawStdEncoding.EncodeToString(buff)

		cmd := "key"
		if need_rekey {
			cmd = "rekey"
		}

		p := "PRAGMA " + cmd + " = '" + plugin_status.Key + "';"
		err = plug_db.Exec(p).Error
		if err != nil {
			return err
		}

		plugin_status.Key = buff_str
		plugin_status.LastDBReKey = time.Now().UTC()

	}

	plugsysiface, err := NewAppPlugSysIface(plugwrap)
	if err != nil {
		return err
	}

	if plugwrap.Plugin == nil {
		panic("plugwrap.Plugin == nil")
	}

	err = plugwrap.Plugin.Init(plugsysiface, plug_db)
	if err != nil {
		return err
	}

	self.plugins[plugwrap.Name] = plugwrap

	if create_record {
		if !no_create {
			err = self.db.Create(&plugin_status).Error
			if err != nil {
				return err
			}
		}
	} else {
		err = self.db.Save(&plugin_status).Error
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
