package appplugsys

type PluginOpenerI interface {
	Open(path string) (*Plugin, error)
}
