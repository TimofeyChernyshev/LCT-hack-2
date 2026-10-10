package migrations

import "embed"

//go:embed 00001_init.up.sql
var FS embed.FS
