package sqlpostgres

import (
	"embed"
)

//go:embed seeds/*.sql
var SeedsFs embed.FS
