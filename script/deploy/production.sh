#!/usr/bin/env bash
# Initial layout must be verified before PPANEL_PRODUCTION_VERIFIED is enabled.
# Never recreate databases, replace config secrets, or restore a DB automatically.
set -euo pipefail
umask 077
sha=${1:?commit SHA required}
[[ "$sha" =~ ^[0-9a-f]{40}$ ]] || exit 2
root=/opt/1panel/docker/compose/ppanel
release=/opt/ppanel-releases/$sha
cd "$root"
exec 9>/var/lock/ppanel-production.lock
flock -n 9 || { echo 'Another backend deployment is active'; exit 1; }
[[ -f docker-compose.yml && -f config/ppanel.yaml ]]
# No unseen overrides: a stale binary bind mount would silently defeat an upgrade.
for f in compose.override.yml compose.override.yaml docker-compose.override.yml docker-compose.override.yaml; do
  [[ ! -f "$f" ]] || { echo "Unexpected Compose override: $f"; exit 1; }
done
docker compose -f docker-compose.yml config --format json > "$release/before.json"
python3 - "$release/before.json" <<'PY'
import json,sys
c=json.load(open(sys.argv[1])); s=c['services']
assert s['backend']['container_name']=='ppanel-backend'
assert s['database']['container_name']=='ppanel-postgres'
assert s['database']['image'].startswith('postgres:')
assert s['database']['environment']['POSTGRES_DB']=='ppanel'
assert s['database']['environment']['POSTGRES_USER']=='ppanel'
assert s['backend']['image'].startswith('ppanel-backend:')
for v in s['backend'].get('volumes',[]):
    if v['target']=='/app/ppanel':
        assert v['source'].endswith('/config/ppanel-mail-ehlo'), 'Unrecognized binary override'
PY
version=$(docker exec ppanel-postgres psql -U ppanel -d ppanel -Atc 'SELECT version FROM schema_migrations WHERE dirty = false')
# This release adds only one boolean column. Earlier/later or dirty schema needs review.
[[ "$version" == 3024 || "$version" == 3025 ]] || { echo 'Unreviewed or dirty migration state; refusing deployment'; exit 1; }
gzip -dc "$release/backend-image.tar.gz" | docker load >/dev/null
docker image inspect "ppanel-backend:$sha" >/dev/null
mkdir -m 700 -p "$root/backups"
backup=$(mktemp -d "$root/backups/$(date -u +%Y%m%dT%H%M%SZ)-$sha.XXXXXX")
cp -p docker-compose.yml "$backup/"
cp -p config/ppanel.yaml "$backup/"
cp -p .env "$backup/" 2>/dev/null || true
# Install recovery before stopping: even a partially successful stop must recover.
previous_image=$(docker inspect ppanel-backend --format '{{.Image}}')
started=false
cleanup() {
  status=$?
  trap - EXIT INT TERM HUP
  if [[ "$started" == false ]]; then
    set +e
    echo "Deployment failed; preserving backup at $backup. Attempting application rollback."
    cp -p "$backup/docker-compose.yml" docker-compose.yml
    docker compose -f docker-compose.yml up -d --no-deps backend
    recovered=false
    for attempt in $(seq 1 40); do
      if [[ $(docker inspect ppanel-backend --format '{{.State.Health.Status}}' 2>/dev/null) == healthy && $(docker inspect ppanel-backend --format '{{.Image}}' 2>/dev/null) == "$previous_image" ]]; then
        recovered=true
        break
      fi
      sleep 5
    done
    if [[ "$recovered" == true ]]; then
      echo 'Application rollback verified: previous image is healthy. Database was not restored.'
    else
      echo 'MANUAL RECOVERY REQUIRED: previous application health/image could not be verified.' >&2
    fi
    [[ "$status" != 0 ]] || status=1
  fi
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
trap 'exit 129' HUP
# Stop all application writers before taking a consistent, restorable snapshot.
docker compose -f docker-compose.yml stop backend
docker exec ppanel-postgres pg_dump -U ppanel -d ppanel -Fc > "$backup/database.dump"
test -s "$backup/database.dump"
docker exec -i ppanel-postgres pg_restore --list < "$backup/database.dump" > "$backup/database.contents"
test -s "$backup/database.contents"
# Use Compose's own normalized JSON to preserve all live config, including secrets,
# ports, resources and non-backend services. JSON is valid YAML.
python3 - "$release/before.json" "$sha" <<'PY'
import json,sys,os
c=json.load(open(sys.argv[1])); s=c['services']['backend']
s['image']='ppanel-backend:'+sys.argv[2]
s['volumes']=[v for v in s.get('volumes',[]) if v['target']!='/app/ppanel']
# Normalized Compose values are already interpolated. Escape literal dollars
# before writing them back, otherwise passwords/commands can be interpolated twice.
def escape(v):
    if isinstance(v,str): return v.replace('$','$$')
    if isinstance(v,list): return [escape(x) for x in v]
    if isinstance(v,dict): return {k:escape(x) for k,x in v.items()}
    return v
c=escape(c)
p='docker-compose.yml.next'
with open(p,'w') as f: json.dump(c,f,indent=2)
os.chmod(p,0o600)
os.replace(p,'docker-compose.yml')
PY
docker compose -f docker-compose.yml config --quiet
docker compose -f docker-compose.yml up -d --no-deps backend
for attempt in $(seq 1 40); do
  if [[ $(docker inspect ppanel-backend --format '{{.State.Health.Status}}') == healthy ]]; then
    actual=$(docker inspect ppanel-backend --format '{{.Image}}')
    expected=$(docker image inspect "ppanel-backend:$sha" --format '{{.Id}}')
    [[ "$actual" == "$expected" ]]
    [[ $(docker exec ppanel-postgres psql -U ppanel -d ppanel -Atc 'SELECT version FROM schema_migrations WHERE dirty = false') == 3025 ]]
    started=true
    printf '%s\n' "$sha" > "$root/deployed-backend-sha"
    echo "Healthy backend deployed: $sha; database backup: $backup"
    exit 0
  fi
  sleep 5
done
echo 'Backend failed health verification; database is NOT automatically downgraded/restored.'
exit 1
