package migrations

import "embed"

//go:embed 00001_init.up.sql
//go:embed 00003_company_owner_unique.sql
var FS embed.FS
