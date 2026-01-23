# Go DSC Pull Server

A secure, modular DSC Pull Server written in Go, with a REST API and PowerShell module for remote management.

## Features

- **Authentication & Security**
  - JWT (Bearer) and API Token authentication
  - SAML (Single Sign-On) support for web authentication
  - Middleware enforces authentication on all sensitive routes
  - Access control, token management, input validation

- **REST API Endpoints**
  - **Agents**: List, filter, get by ID
  - **Reports**: List agent reports, get latest report, get by job ID
  - **Modules**: Upload, list, delete DSC modules
  - **Configurations**: Upload, list, delete DSC configurations
  - **Advanced**: Endpoints for properties and configuration models

- **User & Session Management**
  - User management (API and web)
  - SAML session management

- **Web UI & Static Files**
  - Serves static files for the web interface (JS, CSS, HTML)
  - User profile and server property management

- **Data Storage**
  - SQLite database for all persistent data
  - Includes `.db` files and SQL migrations

- **Code Structure**
  - Modular design: handlers, internal routes, utilities, templates

---

## PowerShell Module: DSCPullServer

A PowerShell module is provided for remote management of the server.

### Exported Commands

- `Connect-DSCPullServer` — Authenticate with the server (token or PSCredential)
- `Get-DSCPullServerAgent` — List/filter DSC agents
- `Get-DSCPullServerReport` — List reports for an agent
- `Add-DSCPullServerModule` — Upload a DSC module
- `Get-DSCPullServerModule` — List available modules
- `Remove-DSCPullServerModule` — Delete a module
- `Add-DSCPullServerConfiguration` — Upload a DSC configuration
- `Get-DSCPullServerConfiguration` — List configurations
- `Remove-DSCPullServerConfiguration` — Delete a configuration

All commands require authentication via `Connect-DSCPullServer`.

---

## Quick Start

1. **Build and run the server:**
   ```sh
   go build -o dsc-pull-server.exe
   ./dsc-pull-server.exe
   ```
2. **Import the PowerShell module:**
   ```powershell
   Import-Module ./powershell/DSCPullServer -Force
   Connect-DSCPullServer -ServerUrl 'https://your-server' -Token 'your-token'
   # or with credentials
   Connect-DSCPullServer -ServerUrl 'https://your-server' -Credential (Get-Credential)
   ```
3. **Use the available commands to manage agents, modules, and configurations.**

---

## Project Structure

- `main.go` — Server entry point
- `handlers/` — API and web handlers
- `internal/` — App config, DB, routes, utils
- `templates/` — Web UI templates
- `web/` — Static files (JS, CSS, etc.)
- `powershell/DSCPullServer/` — PowerShell module

---

## License

MIT License
