# Role-Based Access Control (RBAC)

## Overview

The DSC Pull Server implements a Role-Based Access Control (RBAC) system to provide fine-grained security and governance. This system allows administrators to control user access to different features and operations within the server.

## Roles

The system includes two predefined roles:

### 1. Administrator

**Full access to all server features**

The Administrator role grants complete access to all server operations, including:

#### Read Permissions
- View all agents and their configurations
- View all reports and status information
- List and view modules
- View configuration models
- View properties
- List and view all users
- View role information

#### Write/Modify Permissions
- Create, update, and delete agents
- Manage agent configurations
- Upload, update, and delete modules
- Create, update, and delete configuration models
- Create, update, and delete properties
- Create, update, and delete users
- Assign and remove user roles
- Activate/deactivate users
- Change user passwords

### 2. Read-Only

**Limited to viewing configurations, nodes, and status**

The Read-Only role provides read-only access to monitor and inspect the system without the ability to make changes:

#### Read Permissions
- View all agents and their configurations
- View all reports and status information
- List and view modules
- View configuration models
- View properties
- List and view all users (without password hashes)
- View role information
- Manage own API tokens

#### Restrictions
- **Cannot** create, update, or delete agents
- **Cannot** modify agent configurations
- **Cannot** upload, update, or delete modules
- **Cannot** create, update, or delete configuration models
- **Cannot** create, update, or delete properties
- **Cannot** create, update, or delete users
- **Cannot** assign or remove roles
- **Cannot** activate/deactivate users
- **Cannot** change user passwords (including their own)

## Role Assignment

### Default Role Assignment

- **New Users**: When a user is created without specifying a role, they are automatically assigned the **Read-Only** role for security purposes.
- **Admin User**: The initial admin user (`admin@localhost`) created during server initialization is automatically assigned the **Administrator** role.

### Assigning Roles

Only users with the **Administrator** role can assign or remove roles from other users.

#### Via API

**Assign a role to a user:**
```bash
POST /api/v1/users/{id}/roles
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json

{
  "role_name": "Administrator"
}
```

**Remove a role from a user:**
```bash
DELETE /api/v1/users/{id}/roles
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json

{
  "role_name": "Read-Only"
}
```

**Get user's roles:**
```bash
GET /api/v1/users/{id}/roles
Authorization: Bearer <jwt_token>
```

**List all available roles:**
```bash
GET /api/v1/roles
Authorization: Bearer <jwt_token>
```

### Creating Users with Specific Roles

When creating a new user, an Administrator can specify the role:

```bash
POST /api/v1/users
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "password": "SecurePassword123",
  "is_active": "true",
  "role": "Administrator"
}
```

If the `role` field is omitted, the user will be assigned the **Read-Only** role by default.

## Authentication and Authorization Flow

1. **Authentication**: Users authenticate using their credentials via the `/api/v1/login` endpoint
2. **Token Generation**: Upon successful authentication, a JWT token is generated containing the user's email and roles
3. **Token Usage**: The JWT token must be included in the `Authorization` header as `Bearer <token>` for all protected endpoints
4. **Authorization Check**: The RBAC middleware validates the user's roles against the required roles for each endpoint
5. **Access Grant/Deny**: If the user has the required role, access is granted; otherwise, a `403 Forbidden` response is returned

## API Endpoint Permissions

### Public Endpoints (No Authentication Required)
- `POST /api/v1/login` - User login
- `GET /api/v1/saml/enabled` - Check if SAML is enabled

### Authenticated Endpoints (Both Roles)
- `GET /api/v1/my` - Get current user info
- `GET /api/v1/agents` - List agents
- `GET /api/v1/agents/{id}` - Get agent details
- `GET /api/v1/agents/{id}/configs` - Get agent configurations
- `GET /api/v1/agents/{id}/reports` - List agent reports
- `GET /api/v1/agents/{id}/reports/latest` - Get latest report
- `GET /api/v1/agents/{id}/reports/{jobid}` - Get report by job ID
- `GET /api/v1/modules` - List modules
- `GET /api/v1/properties` - List properties
- `GET /api/v1/properties/{id}` - Get property details
- `GET /api/v1/configuration_models` - List configuration models
- `GET /api/v1/configuration_models/{id}` - Get configuration model details
- `GET /api/v1/users` - List users
- `GET /api/v1/users/{id}` - Get user details
- `GET /api/v1/roles` - List all roles
- `GET /api/v1/users/{id}/roles` - Get user roles
- `GET /api/v1/users/{id}/tokens` - List user API tokens
- `POST /api/v1/users/{id}/tokens` - Create user API token
- `POST /api/v1/users/{id}/tokens/{tokenid}/revoke` - Revoke user API token
- `DELETE /api/v1/users/{id}/tokens/{tokenid}` - Delete user API token

### Administrator-Only Endpoints
- `POST /api/v1/agents/{id}/configs` - Add agent configuration
- `DELETE /api/v1/agents/{id}/configs` - Remove agent configuration
- `POST /api/v1/modules/upload` - Upload module
- `DELETE /api/v1/modules/delete` - Delete module
- `POST /api/v1/properties` - Create property
- `PUT /api/v1/properties/{id}` - Update property
- `DELETE /api/v1/properties/{id}` - Delete property
- `POST /api/v1/configuration_models` - Create configuration model
- `PUT /api/v1/configuration_models/{id}` - Update configuration model
- `DELETE /api/v1/configuration_models` - Delete configuration model
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user
- `POST /api/v1/users/{id}/active` - Activate/deactivate user
- `POST /api/v1/users/{id}/password` - Change user password
- `POST /api/v1/users/{id}/roles` - Assign role to user
- `DELETE /api/v1/users/{id}/roles` - Remove role from user

## Error Responses

### 401 Unauthorized
Returned when authentication fails or token is invalid/expired:
```json
{
  "error": "Unauthorized"
}
```

### 403 Forbidden
Returned when user lacks required permissions:
```json
{
  "error": "Forbidden: insufficient permissions"
}
```

## Security Best Practices

1. **Principle of Least Privilege**: Assign users the minimum role necessary to perform their job functions
2. **Regular Audits**: Periodically review user roles and access patterns
3. **Strong Passwords**: Enforce strong password policies for all users
4. **Token Management**: Use API tokens for programmatic access and rotate them regularly
5. **Monitor Access**: Review logs for unauthorized access attempts
6. **Separation of Duties**: Use Read-Only role for monitoring and reporting, Administrator for operational changes

## Future Enhancements

The RBAC system is designed to be extensible. Future enhancements may include:

- Additional custom roles (e.g., Operator, Auditor)
- Fine-grained permissions at the resource level
- Role hierarchies and inheritance
- Time-based role assignments
- Integration with external identity providers (LDAP, Active Directory)
- Audit logging for role changes and permission denials
