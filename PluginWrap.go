package appplugsys

type PluginWrap struct {
	BasicPluginInfo
	Enabled bool

	Plugin *Plugin

	sys    *AppPlugSys
	pitaps *PluginInterfaceToAppPlugSys
}

func NewPluginWrapFromSearchItem(
	psri *PluginSearcherSearchResultItem,
	sys *AppPlugSys,
) (*PluginWrap, error) {

	self := new(PluginWrap)

	if psri.BuiltIn {
		self.BasicPluginInfo = BasicPluginInfo{
			Name:    psri.Name,
			BuiltIn: true,
			Sha512:  "",
		}

	} else {
		self.BasicPluginInfo = BasicPluginInfo{
			Name:    psri.Name,
			BuiltIn: false,
			Sha512:  psri.Sha512,
		}
	}

	self.sys = sys

	return self, nil
}

func (self *PluginWrap) AppPlugSysIface() *PluginInterfaceToAppPlugSys {
	if self.pitaps == nil {
		self.pitaps = &PluginInterfaceToAppPlugSys{}
	}
	return self.pitaps
}
