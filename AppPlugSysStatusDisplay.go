package appplugsys

import "github.com/AnimusPEXUS/utils/worker/workerstatus"

type AppPlugSysStatusDisplayLine struct {
	Sha512  string
	BuiltIn bool
	Found   bool

	Enabled   bool
	AutoStart bool

	WorkerStatus *workerstatus.WorkerStatus // nil if not worker
}

type AppPlugSysStatusDisplayI interface {
	Clear()
	SetLine(unique_name string)
	DelLine(unique_name string)
}
