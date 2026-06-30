// Package migrations embeds the SQL schema files so they ship inside the binary.
package migrations

import "embed"

// MySQL holds the MySQL migration files (NNNN_name.up.sql / .down.sql).
//
//go:embed mysql/*.sql
var MySQL embed.FS
