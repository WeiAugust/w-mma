package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	mysqlerr "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMySQL(cfg Config) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
}

func RunMigrations(db *gorm.DB, migrationDir string) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
  name VARCHAR(255) PRIMARY KEY,
  applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`).Error; err != nil {
		return fmt.Errorf("ensure schema_migrations failed: %w", err)
	}

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
		var appliedCount int64
		if err := db.Table("schema_migrations").Where("name = ?", name).Count(&appliedCount).Error; err != nil {
			return fmt.Errorf("check migration %s failed: %w", name, err)
		}
		if appliedCount > 0 {
			continue
		}

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
				if isIgnorableMigrationError(err) {
					continue
				}
				return fmt.Errorf("apply migration %s: %w", name, err)
			}
		}

		if err := db.Exec("INSERT INTO schema_migrations (name) VALUES (?)", name).Error; err != nil {
			if isIgnorableMigrationError(err) {
				continue
			}
			return fmt.Errorf("mark migration %s as applied: %w", name, err)
		}
	}
	return nil
}

func isIgnorableMigrationError(err error) bool {
	var mysqlError *mysqlerr.MySQLError
	if !errors.As(err, &mysqlError) {
		return false
	}

	// 1060: duplicate column name (legacy DB reruns additive migration).
	// 1062: duplicate entry (concurrent process inserts same migration marker).
	return mysqlError.Number == 1060 || mysqlError.Number == 1062
}
