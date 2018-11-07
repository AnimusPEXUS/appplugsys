package appplugsys

type PluginSearcherSearchResultItem struct {
	BasicPluginInfo
	Path   string
	Plugin *Plugin
}

type PluginSearcherI interface {
	FindAll() ([]*PluginSearcherSearchResultItem, error)
	FindBuiltIn(name string) (*PluginSearcherSearchResultItem, error)
	FindBySha512(sha512 string) (*PluginSearcherSearchResultItem, error)
}

type PluginSearcher struct {
	builtin_blugins []*Plugin
}
