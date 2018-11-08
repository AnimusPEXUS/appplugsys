package appplugsys

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
	return pw.Plugin, nil
}
