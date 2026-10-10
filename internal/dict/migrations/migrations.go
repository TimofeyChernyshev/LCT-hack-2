package migrations

import "embed"

//go:embed 00001_init.up.sql
//go:embed 00002_seed.sql
//go:embed 00003_more_technologies.sql
//go:embed 00004_all_spec_grades.sql
var FS embed.FS
