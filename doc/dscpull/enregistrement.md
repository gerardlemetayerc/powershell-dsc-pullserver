# Node Registration

This page explains how to register nodes (agents) with the DSC Pull Server.

## Registration Process

1. **Obtain the registration key:**
   - The registration key is defined in the `config.json` file under the `dsc_pullserver` section (option: `registrationKey`).

2. **Agent configuration:**
   - Configure your DSC agent to use the server URL and provide the registration key during registration.
      - Example agent configuration (LCM) for DSC Pull:
      
         The Local Configuration Manager (LCM) on Windows nodes is configured via a MOF file or PowerShell script. Here is a sample configuration for DSC Pull with reporting enabled:
      
         ```powershell
         [DSCLocalConfigurationManager()]
         Configuration LCMConfig {
            Node "localhost" {
               Settings {
                  RefreshMode = "Pull"
                  RefreshFrequencyMins = 30
                  ConfigurationMode = "ApplyAndMonitor"
                  ConfigurationModeFrequencyMins = 15
                  RebootNodeIfNeeded = $true
                  ReportServerWeb = "https://your-server:8484/PSDSCPullServer.svc"
                  DownloadManagerName = "WebDownloadManager"
                  DownloadManagerCustomData = @{ 
                     ServerUrl = "https://your-server:8484/PSDSCPullServer.svc";
                     RegistrationKey = "AnyString"
                  }
               }
            }
         }
         LCMConfig
         Set-DscLocalConfigurationManager -Path ./LCMConfig
         ```
      
             - `ReportServerWeb`: URL for reporting to the DSC Pull Server.
                > Note: For now, the path must include `PSDSCPullServer.svc` (the official Microsoft service URL is retained). This requirement may be revisited in the future to allow custom endpoints.
         - `DownloadManagerCustomData`: Contains the server URL and registration key.
         - `RefreshMode`: Set to "Pull" for DSC Pull mode.
         - `RefreshFrequencyMins`: How often the node checks for new configurations.
         - `ConfigurationMode`: "ApplyAndMonitor" enables reporting.
         - `ConfigurationModeFrequencyMins`: How often the node applies and monitors configuration.
         - `RebootNodeIfNeeded`: Allows automatic reboot if required.
      
             Apply this configuration using PowerShell to set up the LCM for DSC Pull and reporting.

             ### Example: Configuration with ConfigurationRepositoryWeb and ReportServerWeb

             Here is a typical LCM configuration using ConfigurationRepositoryWeb and ReportServerWeb blocks:

             ```powershell
             [DSCLocalConfigurationManager()]
             Configuration PullLCMConfig {
                Node "localhost" {
                   Settings {
                      RefreshMode = "Pull"
                      ConfigurationMode = "ApplyOnly"
                   }
                   ConfigurationRepositoryWeb PullSrv {
                      ServerURL = "http://127.0.0.1:8080/PSDSCPullServer.svc"
                      AllowUnsecureConnection = $true
                      RegistrationKey = "Test"
                   }

                   ReportServerWeb PullReport {
                      ServerURL = "http://127.0.0.1:8080/PSDSCPullServer.svc"
                      AllowUnsecureConnection = $true
                   }
                }
             }
             PullLCMConfig
             Set-DscLocalConfigurationManager -Path ./PullLCMConfig
             ```

             - `ConfigurationRepositoryWeb`: Defines the pull server endpoint and registration key for configuration retrieval.
             - `ReportServerWeb`: Defines the endpoint for reporting compliance and status.
             - `AllowUnsecureConnection`: Set to `$true` to allow HTTP (not recommended for production).
             - `ServerURL`: Must include `PSDSCPullServer.svc`.
             - `RegistrationKey`: Must match the server configuration.

             > Note: Using HTTP is only suitable for testing or isolated environments. Always use HTTPS in production for security.

3. **Certificate requirements:**
   - If client certificate validation is enabled, ensure the agent presents a valid certificate.
   - The server may require HTTPS and certificate validation depending on your config.json settings.

4. **Successful registration:**
   - Upon successful registration, the agent will appear in the web interface and can receive configurations.

## Troubleshooting

- Check server logs for registration errors.
- Ensure the registration key matches the value in config.json.
- Verify network connectivity and certificate validity.

---

> For advanced scenarios (SAML, JWT, custom roles), see the authentication and security documentation.
