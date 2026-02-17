package storage

import (
	"database/sql"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func setupMySQLForTest(t *testing.T) *sql.DB {
	t.Helper()

	containerName := fmt.Sprintf("bajiaozhi-mysql-test-%d", time.Now().UnixNano())

	out, err := exec.Command(
		"docker", "run", "-d", "--rm",
		"--name", containerName,
		"-e", "MYSQL_ROOT_PASSWORD=root",
		"-e", "MYSQL_DATABASE=bajiaozhi_test",
		"-p", "127.0.0.1:0:3306",
		"public.ecr.aws/docker/library/mysql:8.0.36",
	).CombinedOutput()
	if err != nil {
		t.Fatalf("start mysql container failed: %v (%s)", err, strings.TrimSpace(string(out)))
	}

	t.Cleanup(func() {
		_, _ = exec.Command("docker", "rm", "-f", containerName).CombinedOutput()
	})

	portOut, err := exec.Command("docker", "port", containerName, "3306/tcp").CombinedOutput()
	if err != nil {
		t.Fatalf("discover mysql port failed: %v (%s)", err, strings.TrimSpace(string(portOut)))
	}

	addr := strings.TrimSpace(string(portOut))
	dsn := fmt.Sprintf("root:root@tcp(%s)/bajiaozhi_test?parseTime=true&multiStatements=true", addr)

	var db *sql.DB
	deadline := time.Now().Add(90 * time.Second)
	for {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("mysql did not become ready: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
