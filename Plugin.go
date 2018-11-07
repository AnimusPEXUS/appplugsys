package appplugsys

import "github.com/AnimusPEXUS/utils/worker"

type Plugin struct {
	Name        string
	Description string

	Requires []string // unique names of required plugins

	Init func(*PluginInterfaceToAppPlugSys)

	Worker *worker.Worker // set this to nil, if plugin does not support this
}

type BasicPluginInfo struct {
	Name    string
	BuiltIn bool
	Sha512  string
}
