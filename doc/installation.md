# Installation Guide

This page explains how to install and deploy the Go DSC Pull Server.

## Prerequisites
- Go (latest stable version)
- Microsoft SQL Server or SQLite
- PowerShell (for module usage)
- SSL certificate and private key (for HTTPS)

## Steps

1. **Download the latest release:**
   - Go to the [GitHub Releases page](https://github.com/your-org/go-dsc-pull/releases) and download the pre-built executable for your platform.
   - Extract the archive and place the executable in your desired directory.

2. **Configure the application:**
   - Edit `config.json` for database, SAML, and web UI settings.
   - Place your SSL certificate and key files in the specified paths.

3. **Deploy the server:**
    - Run the executable:
       ```sh
       ./dsc-pull-server.exe
       ```
    - (Optional) Register as a Windows service using `sc.exe`:
      1. Open a command prompt as Administrator.
      2. Run the following command (adapt the path as needed):
         ```sh
         sc.exe create DSCPullServer binPath= "C:\\Path\\To\\dsc-pull-server.exe" start= auto
         ```
      3. Start the service:
         ```sh
         sc.exe start DSCPullServer
         ```
      4. To stop or delete the service:
         ```sh
         sc.exe stop DSCPullServer
         sc.exe delete DSCPullServer
         ```

5. **Database setup:**
   - For MSSQL: run `db/schema_mssql.sql` to create tables.
   - For SQLite: run `db/schema_sqlite.sql`.

6. **Certificate management:**
   - Use a trusted certificate for production.
   - You can disable client certificate validation in `config.json` for testing.

7. **Web UI access:**
   - Open your browser and go to `https://your-server:443` (or the port specified).

8. **PowerShell module:**
   - Import the module from `powershell/DSCPullServer/`.
   - Authenticate and manage agents, modules, and configurations.

## Troubleshooting
- Check logs for errors in the `logs/` directory.
- Ensure all ports are open and certificates are valid.
- Review `config.json` for correct paths and settings.

## Upgrade/Migration
- Backup your database before upgrading.
- Run migration scripts if schema changes are present.
- Restart the server after upgrade.

## Security Notes
- Always use HTTPS in production.
- Configure SAML or JWT authentication for secure access.
- Restrict admin access and review logs regularly.
