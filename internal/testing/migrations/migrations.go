package migrations

import "embed"

//go:embed 00001_init.up.sql
//go:embed 00002_seed_tasks.sql
//go:embed 00003_category_cooldown_and_questionnaire.sql
//go:embed 00004_fsp_integration.sql
var FS embed.FS

