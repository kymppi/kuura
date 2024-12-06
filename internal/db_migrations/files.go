package db_migrations

import (
	"embed"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed *.sql
var Files embed.FS

var Migrations = &migrate.EmbedFileSystemMigrationSource{
	FileSystem: Files,
	Root:       ".",
}
