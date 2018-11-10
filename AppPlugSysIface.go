package appplugsys

import "errors"

type AppPlugSysIface struct {
	plugin_wrap *PluginWrap
}

func NewAppPlugSysIface(plugin_wrap *PluginWrap) (*AppPlugSysIface, error) {
	self := new(AppPlugSysIface)
	self.plugin_wrap = plugin_wrap
	return self, nil
}

func (self *AppPlugSysIface) GetPlugin(name string) (*Plugin, error) {
	pw, err := self.plugin_wrap.sys.GetPluginByName(name)
	if err != nil {
		return nil, err
	}

	if pw.Plugin == nil {
		return nil, errors.New("plugin not loaded into wrap")
	}

	return pw.Plugin, nil
}

func (self *AppPlugSysIface) GetPluginController(name string) (interface{}, error) {

	plugin, err := self.GetPlugin(name)
	if err != nil {
		return nil, err
	}

	ret := plugin.GetController()
	if err != nil {
		return nil, errors.New("plugin have no controller")
	}

	return ret, nil
}

func (self *AppPlugSysIface) GetPluginApplication(plugname string, name string) (*PluginApplication, error) {

	plugin, err := self.GetPlugin(name)
	if err != nil {
		return nil, err
	}

	for k, v := range plugin.Applications {
		if k == name {
			return v, nil
		}
	}

	return nil, errors.New("plugin's application not found")
}

func (self *AppPlugSysIface) GetPluginApplicationController(plugname string, name string) (interface{}, error) {

	pluginapplication, err := self.GetPluginApplication(plugname, name)
	if err != nil {
		return nil, err
	}

	ret := pluginapplication.GetController()
	if err != nil {
		return nil, errors.New("application have no controller")
	}

	return ret, nil
}
