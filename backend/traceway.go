package tracewaybackend

import "github.com/tracewayapp/traceway/backend/cmd"

type Option = cmd.Option

var (
	Run                = cmd.Run
	WithPort           = cmd.WithPort
	WithServerURL      = cmd.WithServerURL
	WithSQLitePath     = cmd.WithSQLitePath
	WithClickhousePath = cmd.WithClickhousePath
	WithDefaultUser    = cmd.WithDefaultUser
	WithDefaultProject = cmd.WithDefaultProject
)
