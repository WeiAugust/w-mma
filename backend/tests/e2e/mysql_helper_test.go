package e2e

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func setupMySQLDSNForTest(t *testing.T) string {
	t.Helper()

	containerName := fmt.Sprintf("bajiaozhi-e2e-mysql-%d", time.Now().UnixNano())
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
	deadline := time.Now().Add(90 * time.Second)
	for {
		cmd := exec.Command("/bin/sh", "-lc", fmt.Sprintf("mysqladmin --protocol=tcp --host=%s --port=%s --user=root --password=root ping >/dev/null 2>&1", strings.Split(addr, ":")[0], strings.Split(addr, ":")[1]))
		if cmd.Run() == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("mysql did not become ready")
		}
		time.Sleep(time.Second)
	}

	return dsn
}
