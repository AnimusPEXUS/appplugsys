package appplugsys

import "github.com/AnimusPEXUS/utils/worker"

type Plugin struct {
	Name        string
	Title       string
	Description string

	Icon []byte

	Requires []string // unique names of required plugins

	Init    func(*PluginInterfaceToAppPlugSys) error
	Destroy func() error

	Worker *worker.Worker // set this to nil, if plugin does not support this

	GetController func() interface{}

	Display func() error // set this to nil, if plugin does not support this

	Applications []*PluginApplication
}

type BasicPluginInfo struct {
	Name    string
	BuiltIn bool
	Sha512  string
}

type PluginApplication struct {
	Name        string
	Title       string
	Description string

	Icon []byte

	Worker *worker.Worker

	GetController func() interface{}

	Display func() error // set this to nil, if app does not support this
}

type AppViewI interface {
	Clear()
	SetApplication(title, plugin_name, name string)
	RmApplication(plugin_name, name string)
}
