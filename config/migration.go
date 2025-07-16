package config

import (
	"os"
	"path/filepath"
)

func RunMigrations() error {
	schemaPath := filepath.Join(".", "schema.sql")
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	_, err = DB.Exec(string(content))
	return err
}
