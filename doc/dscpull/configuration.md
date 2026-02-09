# DSC Pull Server Configuration

This page describes the expected structure and options for the DSC Pull Server configuration file (`config.json`).

## Main Section: `dsc_pullserver`

Example:
```json
{
  "dsc_pullserver": {
    "cert_file": "dsc-dev.local.crt",
    "enable_https": true,
    "enable_client_cert_validation": false,
    "bypass_ca_validation": true,
    "key_file": "dsc-dev.local.key",
    "port": 8484,
    "registrationKey": "AnyString"
  }
}
```

### Option Descriptions

- **cert_file**: Path to the server's SSL certificate file. Required for HTTPS.
- **key_file**: Path to the server's private key file. Required for HTTPS.
- **enable_https**: Set to `true` to enable HTTPS. If `false`, the server will run in HTTP mode.
- **enable_client_cert_validation**: If `true`, the server will require client certificates for authentication. Recommended for production.
- **bypass_ca_validation**: If `true`, disables CA validation for client certificates (useful for testing/self-signed certs).
- **port**: Port number the server listens on (default: 8484).
- **registrationKey**: Key required for agent registration. Set any string value; agents must provide this key to register.

## Additional Sections

Other configuration sections (database, SAML, JWT, web) may be present depending on your deployment. See the main documentation for details.

---

> Ensure all paths are correct and certificates are valid for secure operation.
