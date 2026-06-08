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

## Contribution
To contribute, open an issue or submit a pull request. Any help to improve the documentation or code is welcome.
# Go DSC Pull Server

A secure, modular DSC Pull Server written in Go, with a REST API and PowerShell module for remote management.

## Build Windows versionne

Pour generer un exe Windows avec:
- metadonnees visibles dans les Proprietes du fichier (Product Version, File Version, etc.)
- metadonnees de build injectees dans les logs (version, commit, date)

Utilisez:

```powershell
./scripts/build-windows.ps1
```

Le fichier `build/version.json` est lu automatiquement par defaut et sert de source locale de versionning et de metadonnees produit.

Depuis un environnement non-Windows, vous pouvez aussi cibler Windows en definissant `GOOS=windows` et `GOARCH` avant d'executer le script (PowerShell 7 requis).

## License

MIT License

## Release Notes (v1.1.3)

- Added scheduling controls for periodic GitHub latest-release checks (`enable_release_check`, `release_check_interval_mins`).
- Added persistence of release-check results in infrastructure metadata (`latest_release`, `latest_release_url`, `update_available`, `release_check_ok`, `release_checked_at`).
- Added API/UI support to configure and surface release-check state from the scheduler and about endpoints.
