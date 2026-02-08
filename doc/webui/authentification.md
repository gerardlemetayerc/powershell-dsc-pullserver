# Authentication

This page explains the authentication mechanisms available in the DSC Pull Server web interface.

## Supported Methods
- Local authentication (username/password)
- SAML Single Sign-On (SSO)

## SAML Configuration

SAML Single Sign-On (SSO) allows integration with enterprise identity providers (IdP) for centralized authentication.

### Example Configuration for Azure (config.json)

```json
"saml": {
	"enabled": true,
	"entity_id": "https://dsc-dev.local",
	"idp_metadata_url": "https://login.microsoftonline.com/yourtenantID/FederationMetadata/2007-06/FederationMetadata.xml?appid=appId",
	"sp_key_file": "sp.key",
	"sp_cert_file": "sp.crt.pem",
	"user_mapping": {
		"email": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
		"sn": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
		"givenName": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
	},
	"group_mapping": {
		"attribute": "http://schemas.microsoft.com/ws/2008/06/identity/claims/groups",
		"admin_value": "9f59531c-ae20-42b6-be94-a4dae3623eb0",
		"user_value": "b79fbf4d-3ef9-4689-8143-76b194e85509"
	}
}
```

#### Field Descriptions

- `enabled`: Enables or disables SAML authentication (true/false).
- `entity_id`: The unique identifier (URI) for your Service Provider (SP).
- `idp_metadata_url`: URL to fetch the IdP's SAML metadata (XML with endpoints and certificates).
- `sp_key_file`: Path to the private key file for the SP (used to sign SAML requests).
- `sp_cert_file`: Path to the public certificate file for the SP (used for SAML assertions).
- `user_mapping`: Maps SAML attributes to local user fields:
  - `email`: SAML attribute for user email.
  - `sn`: SAML attribute for surname (last name).
  - `givenName`: SAML attribute for given name (first name).
- `group_mapping`: Maps SAML group/role claims to application roles:
  - `attribute`: SAML attribute containing group or role IDs.
  - `admin_value`: Value identifying admin group(s).
  - `user_value`: Value identifying user group(s).

### Setup Steps
- Obtain SAML metadata (XML) from your IdP administrator.
- Upload the metadata file in the web interface (SAML config section).
- Configure the Service Provider (SP) settings: entity ID, ACS URL, certificate, and private key.
- Map SAML groups/roles to application roles (admin/user) in the configuration file.
- Optionally, enable or disable certificate signature validation for flexibility.

### User Mapping
- SAML attributes (e.g., email, group, role) are mapped to local user accounts.
- Admins can define mapping rules in config.json for automatic role assignment.

### Security Notes
- SAML authentication is enforced for all web sessions when enabled.
- All SAML login attempts and errors are logged for auditing.

## Security Notes
- All sensitive routes require authentication.
- Admins can manage user roles and tokens.