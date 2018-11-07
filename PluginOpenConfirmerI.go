package appplugsys

type PluginOpenConfirmerI interface {
	ConfirmPluginOpening(sha512, name, path string) (bool, error)
}
