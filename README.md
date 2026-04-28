# go-dsc-pull

A secure, versioned DSC Pull Server with a modern web interface.

## Purpose
Centralized management of DSC configurations, modules, users, and reports, with advanced access controls and a user-friendly web UI.

## Documentation

Detailed documentation is organized into two main sections:

- [WEB UI Interface](doc/webui/README.md)
- [DSC Pull Server](doc/dscpull/README.md)

## Quick Navigation

- [Installation](doc/installation.md)
- [Node Registration](doc/dscpull/enregistrement.md)
- [Security](doc/dscpull/securite.md)
- [Web Authentication](doc/webui/authentification.md)
- [Resource Modules](doc/webui/modules.md)
- [MOF Configurations](doc/webui/configurations.md)
- [Roles & Access](doc/webui/droits.md)
- [Storage Modes](doc/webui/stockage.md)

## Version 1.2.0 Highlights

- Node-level scheduled configuration override (one-shot or recurring) with configurable tolerance window.
- Main header "A propos" menu showing current application version.
- Database migration scripts for upgrade from `1.1.1` to `1.2.0`:
	- `db/migration_v1.1.1_to_v1.2.0_sqlite.sql`
	- `db/migration_v1.1.1_to_v1.2.0_mssql.sql`

## Contribution
To contribute, open an issue or submit a pull request. Any help to improve the documentation or code is welcome.
# Go DSC Pull Server

A secure, modular DSC Pull Server written in Go, with a REST API and PowerShell module for remote management.

## License

MIT License
