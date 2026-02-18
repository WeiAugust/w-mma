#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin123456}"

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

need_cmd curl
need_cmd node

json_pick() {
  local expr="$1"
  node -e "const fs=require('fs'); const raw=fs.readFileSync(0,'utf8'); const data=JSON.parse(raw); const fn=new Function('data', 'return (' + process.argv[1] + ');'); const out=fn(data); if (out === undefined || out === null) process.exit(9); process.stdout.write(String(out));" "$expr"
}

echo "[1/5] login admin API: ${API_BASE_URL}"
login_resp="$(curl -fsS -X POST "${API_BASE_URL}/admin/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"${ADMIN_USERNAME}\",\"password\":\"${ADMIN_PASSWORD}\"}")"

token="$(printf '%s' "${login_resp}" | json_pick 'data.token')"
if [[ -z "${token}" ]]; then
  echo "login failed: empty token" >&2
  exit 1
fi

auth_header="Authorization: Bearer ${token}"

echo "[2/5] locate UFC schedule source"
sources_resp="$(curl -fsS -H "${auth_header}" "${API_BASE_URL}/admin/sources?source_type=schedule&platform=ufc&enabled=true")"
source_id="$(printf '%s' "${sources_resp}" | json_pick 'data.items && data.items[0] && data.items[0].id')"
if [[ -z "${source_id}" ]]; then
  echo "no enabled UFC schedule source found" >&2
  exit 1
fi

echo "[3/5] trigger source sync: ${source_id}"
sync_resp="$(curl -fsS -X POST -H "${auth_header}" "${API_BASE_URL}/admin/sources/${source_id}/sync")"
sync_events="$(printf '%s' "${sync_resp}" | json_pick 'data.result && data.result.events')"
sync_bouts="$(printf '%s' "${sync_resp}" | json_pick 'data.result && data.result.bouts')"
echo "sync result: events=${sync_events}, bouts=${sync_bouts}"

echo "[4/5] load latest UFC event"
events_resp="$(curl -fsS "${API_BASE_URL}/api/events")"
candidate_ids="$(printf '%s' "${events_resp}" | node -e '
const fs=require("fs");
const raw=fs.readFileSync(0,"utf8");
const data=JSON.parse(raw);
const items=Array.isArray(data.items)?data.items:[];
const ids=items.filter((it)=>it.org==="UFC").map((it)=>it.id);
if(ids.length===0){process.exit(2)}
process.stdout.write(ids.join(" "));
')"
if [[ -z "${candidate_ids}" ]]; then
  echo "no UFC events returned from /api/events" >&2
  exit 1
fi

event_id=""
detail_resp=""
for id in ${candidate_ids}; do
  one_detail="$(curl -fsS "${API_BASE_URL}/api/events/${id}")"
  has_bout_data="$(printf '%s' "${one_detail}" | node -e '
const fs=require("fs");
const raw=fs.readFileSync(0,"utf8");
const data=JSON.parse(raw);
const mainCard=Array.isArray(data.main_card)?data.main_card:[];
const prelims=Array.isArray(data.prelims)?data.prelims:[];
const bouts=Array.isArray(data.bouts)?data.bouts:[];
const total=mainCard.length+prelims.length+bouts.length;
process.stdout.write(total>0?"1":"0");
')"
  if [[ "${has_bout_data}" == "1" ]]; then
    event_id="${id}"
    detail_resp="${one_detail}"
    break
  fi
done
if [[ -z "${event_id}" ]]; then
  echo "NO_BOUT_DATA" >&2
  exit 2
fi

echo "[5/5] validate event detail payload: event=${event_id}"
set +e
validation_out="$(printf '%s' "${detail_resp}" | node -e '
const fs=require("fs");
const raw=fs.readFileSync(0,"utf8");
const data=JSON.parse(raw);
const localPrefix=(process.argv[1]||"").replace(/\/$/,"");
const mainCard = Array.isArray(data.main_card) ? data.main_card : [];
const prelims = Array.isArray(data.prelims) ? data.prelims : [];
const bouts = Array.isArray(data.bouts) ? data.bouts : [];
if (!Array.isArray(data.main_card) || !Array.isArray(data.prelims)) {
  console.error("invalid payload: main_card/prelims should always be arrays");
  process.exit(1);
}
const total = mainCard.length + prelims.length;
if (total === 0) {
  if (bouts.length === 0) {
    console.error("NO_BOUT_DATA");
    process.exit(2);
  }
  console.error("no grouped card data; only legacy bouts found");
  process.exit(1);
}
const bout = mainCard[0] || prelims[0];
if (data.poster_url && !data.poster_url.startsWith(localPrefix + "/media-cache/ufc/")) {
  console.error("invalid payload: poster_url should use local media cache");
  process.exit(1);
}
for (const side of ["red_fighter", "blue_fighter"]) {
  const fighter = bout[side];
  if (!fighter || typeof fighter !== "object") {
    console.error(`invalid payload: missing ${side}`);
    process.exit(1);
  }
  for (const field of ["id", "name", "country", "rank", "weight_class", "avatar_url"]) {
    if (!(field in fighter)) {
      console.error(`invalid payload: ${side}.${field} is missing`);
      process.exit(1);
    }
  }
  if (fighter.avatar_url && !fighter.avatar_url.startsWith(localPrefix + "/media-cache/ufc/")) {
    console.error(`invalid payload: ${side}.avatar_url should use local media cache`);
    process.exit(1);
  }
}
if (!("weight_class" in bout)) {
  console.error("invalid payload: bout.weight_class is missing");
  process.exit(1);
}
console.log(`payload ok: main_card=${mainCard.length}, prelims=${prelims.length}`);
' "${API_BASE_URL}")"
status=$?
set -e
if [[ $status -eq 2 || "${validation_out}" == "NO_BOUT_DATA" ]]; then
  echo "no bout data available after sync." >&2
  echo "diagnosis: current source may be blocked by region redirect (ufc.com -> ufc.cn 404/403)." >&2
  echo "action: run with a network environment that can access ufc.com event pages, then retry." >&2
  exit 2
fi
if [[ $status -ne 0 ]]; then
  echo "${validation_out}" >&2
  exit $status
fi
echo "${validation_out}"

echo "done: UFC event detail payload is ready for miniapp rendering"
