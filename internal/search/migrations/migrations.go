package migrations

import "embed"

//go:embed 00001_init.up.sql
//go:embed 00002_ranking_and_explainability.sql
var FS embed.FS

