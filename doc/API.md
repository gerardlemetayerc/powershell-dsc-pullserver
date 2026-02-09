# API Endpoints Reference

This page documents the main REST API endpoints provided by the DSC Pull Server. It covers authentication, agent management, configuration, reporting, and SAML integration. Use this as a reference for integrating clients, automation, or for troubleshooting.

---

## Authentication

### 1. JWT Login
- **POST** `/api/login`
  - Request: `{ "username": "...", "password": "..." }`
  - Response: `{ "token": "<JWT>" }`
  - Description: Authenticates a user and returns a JWT token for API access.

### 2. SAML Login
- **GET** `/api/saml/login`
  - Redirects to SAML IdP for authentication.
- **POST** `/api/saml/acs`
  - Assertion Consumer Service endpoint for SAML responses.

---

## Agent Management

### 1. Register Agent
- **POST** `/api/agent/register`
  - Request: `{ "agent_id": "...", "hostname": "...", "registration_key": "..." }`
  - Response: Agent info or error.
  - Description: Registers a new DSC agent.

### 2. Get Agent by ID
- **GET** `/api/agent/{agent_id}`
  - Returns agent details.

### 3. List Agents
- **GET** `/api/agents`
  - Returns a list of all registered agents.

### 4. Update Agent Tags
- **POST** `/api/agent/{agent_id}/tags`
  - Request: `{ "tags": ["tag1", "tag2"] }`
  - Updates agent tags.

---

## Configuration Management

### 1. List Configurations
- **GET** `/api/configurations`
  - Returns all configuration models.

### 2. Get Configuration Content
- **GET** `/api/configuration/{name}`
  - Returns the MOF/DSC content for a configuration.

### 3. Upload Configuration
- **POST** `/api/configuration/upload`
  - Multipart/form-data: MOF file upload.
  - Adds or updates a configuration model.

### 4. Delete Configuration
- **DELETE** `/api/configuration/{name}`
  - Deletes a configuration model.

---

## Reporting

### 1. List Reports
- **GET** `/api/reports`
  - Returns all DSC reports.

### 2. Get Latest Report for Agent
- **GET** `/api/agent/{agent_id}/report/latest`
  - Returns the latest report for a given agent.

### 3. Get Reports by Job ID
- **GET** `/api/reports/job/{job_id}`
  - Returns all reports for a specific job.

### 4. Submit Report
- **POST** `/api/report`
  - Request: DSC report JSON.
  - Submits a new report from an agent.

---

## User Management

### 1. List Users
- **GET** `/api/users`
  - Returns all users (admin only).

### 2. Create User
- **POST** `/api/user`
  - Request: `{ "username": "...", "password": "...", "email": "...", "is_admin": true }`
  - Creates a new user (admin only).

### 3. Update User
- **PUT** `/api/user/{id}`
  - Updates user details (admin only).

### 4. Delete User
- **DELETE** `/api/user/{id}`
  - Deletes a user (admin only).

---

## SAML & Security

### 1. SAML Status
- **GET** `/api/saml/status`
  - Returns SAML configuration and status.

### 2. SAML User Info
- **GET** `/api/saml/userinfo`
  - Returns SAML user attributes for the current session.

---

## Notes
- All endpoints (except login and SAML) require a valid JWT in the `Authorization: Bearer <token>` header.
- Admin endpoints require the user to have `is_admin` privileges.
- For full request/response details, see the OpenAPI/Swagger documentation if available, or refer to the handler code in `handlers/`.

---

> For troubleshooting, check server logs and ensure your JWT or SAML session is valid.
