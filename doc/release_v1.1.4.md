# Release v1.1.4

Date: 2026-06-14

Stabilization and UX/API improvement release, focused on the audit page, dashboard load reduction, Windows installation workflow, and startup robustness.

## Changes

### API and Performance

- Added `GET /api/v1/agents?stats` to return consolidated counters (compliant, failed, pending_enroll, pending_apply).
- Reduced dashboard load by using precomputed stats instead of client-side aggregations.

Impacted files:
- handlers/agent_api.go
- web/reports.js
- doc/APIv1.md

### Audit (WebUI + API)

- Added server-side DataTables ordering support (column mapping + direction).
- Added combined filters: users, actions, date_from, date_to.
- Aligned CSV export with active filters (same filtering logic as list endpoint).
- Added `GET /api/v1/audit/filters` to load filter options.
- UI adjustments for the filters bar (alignment improvements).

Impacted files:
- handlers/audit_list.go
- internal/routes/web_routes.go
- web/admin_audit.js
- templates/audit.tmpl
- doc/APIv1.md

### Scheduler / Startup Robustness

- Fixed handling of scheduler runs stuck in running state at startup.
- For MSSQL, switched to typed date parameters to avoid format-sensitive implicit conversions.

Impacted file:
- internal/db/scheduler_history.go

### Packaging and Windows Scripts

- Added/improved scripts:
  - `scripts/install-windows.ps1`
  - `scripts/update-windows.ps1`
  - `scripts/uninstall-windows.ps1`
- Added registry metadata handling (install path, service, version) and Appwiz integration.
- Included the `scripts/` folder in the release ZIP artifact.

Impacted files:
- scripts/install-windows.ps1
- scripts/update-windows.ps1
- scripts/uninstall-windows.ps1
- .github/workflows/release.yml

### UI

- Style adjustments (navigation/left sidebar and active states) and improved overall visual consistency.

Impacted files:
- templates/menu.tmpl
- web/style.css

## Database

- No new mandatory migration specific to v1.1.4.
- If upgrading from a version older than v1.1.3p2, apply previously required migrations before moving to v1.1.4.

## Update Procedure

Run these steps in order:

1. Stop the service.
2. Back up the database and `config.json`.
3. Replace application files with those from the v1.1.4 ZIP (including `scripts/`).
4. If required by your version history, run missing SQL migrations.
5. Restart the service.

## Post-Update Checks (Recommended)

- Verify access to `/web/admin/audit`.
- Verify ordering/filtering/CSV export on audit.
- Verify dashboard behavior (fast counters and chart loading).
- Verify Windows scripts are present in the deployed package.
- Verify startup logs and absence of critical SQL errors.

## Notes

- A `401` on `/api/v1/audit` indicates a missing or expired authenticated session.
- The release check may report an older remote version depending on the publication repository state.
