#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
LOCAL_PROXY_URL="${LOCAL_PROXY_URL:-http://127.0.0.1:7890}"
MYSQL_DSN="${MYSQL_DSN:-root:root@tcp(127.0.0.1:23306)/bajiaozhi?parseTime=true&multiStatements=true}"
REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:26379}"
PUBLIC_BASE_URL="${PUBLIC_BASE_URL:-http://localhost:8080}"
MEDIA_CACHE_DIR="${MEDIA_CACHE_DIR:-${ROOT_DIR}/.worktrees/media-cache}"

echo "[1/5] ensure mysql/redis containers are up"
(
  cd "${ROOT_DIR}"
  docker compose up -d mysql redis
)

echo "[2/5] stop containerized api/worker to free :8080"
(
  cd "${ROOT_DIR}"
  docker compose stop api worker >/dev/null 2>&1 || true
)

echo "[3/5] verify local proxy: ${LOCAL_PROXY_URL}"
proxy_hostport="${LOCAL_PROXY_URL#http://}"
proxy_host="${proxy_hostport%%:*}"
proxy_port="${proxy_hostport##*:}"
if ! nc -z "${proxy_host}" "${proxy_port}" >/dev/null 2>&1; then
  echo "proxy not reachable at ${LOCAL_PROXY_URL}" >&2
  exit 1
fi

echo "[4/5] start local api process with proxy"
API_LOG="${ROOT_DIR}/.worktrees/local-api.log"
API_BIN="${ROOT_DIR}/.worktrees/local-api-bin"
mkdir -p "${ROOT_DIR}/.worktrees"

cd "${ROOT_DIR}/backend"
go build -o "${API_BIN}" ./cmd/api
MYSQL_DSN="${MYSQL_DSN}" \
REDIS_ADDR="${REDIS_ADDR}" \
REDIS_PASSWORD="" \
REDIS_DB="0" \
PUBLIC_BASE_URL="${PUBLIC_BASE_URL}" \
MEDIA_CACHE_DIR="${MEDIA_CACHE_DIR}" \
HTTP_PROXY="${LOCAL_PROXY_URL}" \
HTTPS_PROXY="${LOCAL_PROXY_URL}" \
NO_PROXY="localhost,127.0.0.1,mysql,redis" \
nohup "${API_BIN}" >"${API_LOG}" 2>&1 < /dev/null &
echo $! > "${ROOT_DIR}/.worktrees/local-api.pid"
cd "${ROOT_DIR}"

echo "[5/5] wait for local api healthz"
for _ in $(seq 1 40); do
  if curl -fsS "${API_BASE_URL}/healthz" >/dev/null 2>&1; then
    echo "local api started: ${API_BASE_URL}"
    echo "log: ${API_LOG}"
    echo "pid: $(cat "${ROOT_DIR}/.worktrees/local-api.pid")"
    exit 0
  fi
  sleep 0.5
 done

echo "local api failed to start; tail log:" >&2
tail -n 80 "${API_LOG}" >&2
exit 1
