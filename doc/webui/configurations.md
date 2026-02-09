# MOF Configurations

This page explains how to manage MOF configuration files in the web interface.

## Features
- Upload, list, and delete MOF files for UI, API and PowerShell module.
- MOF files are versioned: each update creates a new version, allowing  historical tracking (rollback support will be improved in future releases).
- Both admins and regular users can access and use MOF configurations (with permissions enforced by role).
- It is not possible to delete a MOF file that is currently assigned to a node (deletion is blocked for safety).
- Each time a MOF file is used/applied by a node, its "last usage" date is updated. This allows administrators to track which configurations are actively in use and when they were last applied.

## Storage
- MOF files are stored as BLOBs (binary large objects) in the database for reliability and traceability.

- **Microsoft recommends a maximum MOF file size of 1 MB** for optimal performance and compatibility with DSC agents.