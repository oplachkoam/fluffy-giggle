// Package migrations embeds the schema files applied at start.
package migrations

import "embed"

//go:embed *.sql
var Files embed.FS
