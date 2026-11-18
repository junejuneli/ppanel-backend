# GoodVPN main production pipeline

This pipeline is prepared locally, not enabled or verified against production.
The existing `master` checkout contains the GoodVPN changes; the recorded
`origin/main` is an ancestor, but refresh remote refs before fast-forwarding main.
Do not force-push or discard another developer's changes.

## Enable only after server verification

1. Inspect the live deployment at `/opt/1panel/docker/compose/ppanel` and confirm
   its backend service/container, PostgreSQL database, migration version and
   existing SMTP EHLO binary bind mount match `script/deploy/production.sh`.
2. Confirm no external application writers are using this database during deploy.
   The script stops the backend (including its workers) before backup. A short
   maintenance window is required.
3. Configure the GitHub `production` environment and secrets:
   `PPANEL_DEPLOY_HOST`, `PPANEL_DEPLOY_USER`, `PPANEL_DEPLOY_PASSWORD`,
   `PPANEL_DEPLOY_KNOWN_HOSTS`. Verify the SSH fingerprint through a trusted
   channel; do not generate trusted keys with an unauthenticated keyscan.
4. Inspect/restore a database backup in an isolated database before the first
   production run. `pg_restore --list` in the script detects an unreadable archive,
   but does not substitute for a restore drill.
5. Only then set repository variable `PPANEL_PRODUCTION_VERIFIED=true`. Until
   then main produces a checked image artifact, but deployment is skipped.
6. Publish the intended commit to main and verify the run for that exact SHA.

The complete reusable check workflow gates packaging and deployment. An immutable
image tagged with the commit SHA is transferred over SSH; server secrets never
enter the image. The script preserves live configuration and removes only the
known `config/ppanel-mail-ehlo` binary override, since its EHLO fix is now in source.
No static frontend directory or database container is recreated.

## Migration and failure policy

Only clean schemas 3024 and 3025 are accepted. Migration 3025 adds a non-null
`one_time` boolean with a false default. Packaging refuses later migration files
until the migration allowlist and rollback procedure have been reviewed.

Backups contain sensitive database/configuration data and remain in a mode-700
server directory. Each deployment retains the database archive, archive listing,
Compose config, application config and optional environment file. Retention and
off-host encrypted backups should be configured separately, without deleting the
last known-good backup.

On failed deployment restoration of the original Compose file/application image is attempted,
then the previous image ID and health are independently checked. A failed recovery
reports `MANUAL RECOVERY REQUIRED`; the deployment still exits as failed.
the database is deliberately never automatically restored or downgraded. This
release's additive column remains compatible with the prior binary. If the DB is
marked dirty, investigate under maintenance and restore a validated backup only
with an explicit recovery decision. Never perform down migrations automatically.

Success requires Docker health, exact image ID and clean schema 3025. It does not
send a test email, create a payment, or validate every customer flow.
