package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMySQL(cfg Config) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
}

func RunMigrations(db *gorm.DB, migrationDir string) error {
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return err
	}

	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".up.sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		path := filepath.Join(migrationDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		for _, stmt := range strings.Split(string(content), ";") {
			trimmed := strings.TrimSpace(stmt)
			if trimmed == "" {
				continue
			}
			if err := db.Exec(trimmed).Error; err != nil {
				return fmt.Errorf("apply migration %s: %w", name, err)
			}
		}
	}
	return nil
}
