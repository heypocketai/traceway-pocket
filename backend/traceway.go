// Package tracewaybackend provides an embeddable Traceway backend that can be
// run inside your own Go application. Uses SQLite for all storage — no external
// databases required.
//
package tracewaybackend

import "github.com/tracewayapp/traceway/backend/cmd"

type Option = cmd.Option

var (
	Run                = cmd.Run
	WithPort           = cmd.WithPort
	WithServerURL      = cmd.WithServerURL
	WithSQLitePath     = cmd.WithSQLitePath
	WithDefaultUser    = cmd.WithDefaultUser
	WithDefaultProject = cmd.WithDefaultProject
	DisableLogging     = cmd.DisableLogging
)
