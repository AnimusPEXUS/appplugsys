package appplugsys

type PluginInterfaceToAppPlugSys struct {
	connected_plugin_wrap *PluginWrap
	sys                   *AppPlugSys
}

func NewPluginInterfaceToAppPlugSys(
	connected_plugin_wrap *PluginWrap,
	sys *AppPlugSys,
) (
	*PluginInterfaceToAppPlugSys,
	error,
) {
	self := new(PluginInterfaceToAppPlugSys)
	self.connected_plugin_wrap = connected_plugin_wrap
	self.sys = sys
	return self, nil
}

func (self *PluginInterfaceToAppPlugSys) GetControlerInstance() (interface{}, error) {
	return nil, nil
}
