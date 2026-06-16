# Prompt Complet a Fournir a une IA

Tu es un expert Go et Terraform/OpenTofu Plugin Framework.

Ta mission est de generer un provider Terraform/OpenTofu pour provisionner des nodes DSC Pull Server.

Contrainte fondamentale: le cycle de vie Terraform doit s appuyer sur un identifiant immuable `internal_dsc_id` (stable), et jamais sur `agent_id` (car il peut evoluer de `TEMP-...` vers un GUID apres registration DSC).

Le provider doit implementer un resource principal `dscpull_node` avec create, read, update, delete, import.

---

## 1. Contexte API et Auth

Base URL configurable, exemple: `https://dsc-dev.local`

Tous les appels proteges utilisent un header Authorization:

- JWT:

```http
Authorization: Bearer <jwt>
```

- Ou API token:

```http
Authorization: Token <api_token>
```

`Content-Type: application/json` pour les appels avec body.

---

## 2. Endpoints a Utiliser (obligatoires)

### Nodes

1. `GET /api/v1/agents`
2. `POST /api/v1/agents/preenroll`
3. `GET /api/v1/agents/{id}`
4. `DELETE /api/v1/agents/{id}`

### Configurations d un node

5. `GET /api/v1/agents/{id}/configs`
6. `POST /api/v1/agents/{id}/configs`
7. `DELETE /api/v1/agents/{id}/configs`

### Tags d un node

8. `GET /api/v1/agents/{id}/tags`
9. `PUT /api/v1/agents/{id}/tags`
10. `DELETE /api/v1/agents/{id}/tags`

### Optionnels utiles

11. `GET /api/v1/agents?count=1`
12. `GET /api/v1/agents?stats`

---

## 3. Schema JSON Actuel Retourne par les APIs

Important: le backend actuel renvoie `agent_id`, pas `internal_dsc_id` dans les payloads.
Tu dois implementer le provider pour **exiger** `internal_dsc_id` comme identite Terraform, et proposer l adaptation backend necessaire (section 8).

### 3.1 GET /api/v1/agents (response)

```json
[
  {
    "agent_id": "TEMP-8f6a1b2c...",
    "node_name": "node-01",
    "lcm_version": "2.0",
    "registration_type": "ConfigurationRepositoryWeb",
    "certificate_thumbprint": "ABC...",
    "certificate_subject": "CN=node-01",
    "certificate_issuer": "CN=CA",
    "certificate_notbefore": "2026-06-01T00:00:00Z",
    "certificate_notafter": "2027-06-01T00:00:00Z",
    "registered_at": "2026-06-14 22:31:48",
    "last_communication": "2026-06-16 09:12:33",
    "has_error_last_report": false,
    "state": "pending_apply",
    "tags": {
      "env": ["prod"],
      "role": ["web"]
    }
  }
]
```

### 3.2 POST /api/v1/agents/preenroll

Request:

```json
{
  "node_name": "node-01"
}
```

Response:

```json
{
  "agent_id": "TEMP-8f6a1b2c...",
  "node_name": "node-01",
  "last_communication": "0000-00-01 00:00:00",
  "state": "waiting_for_registration"
}
```

### 3.3 GET /api/v1/agents/{id}

Response:

```json
{
  "agent_id": "864FB6D0-4583-4D26-A06E-8F278F698975",
  "node_name": "node-01",
  "lcm_version": "2.0",
  "registration_type": "ConfigurationRepositoryWeb",
  "certificate_thumbprint": "ABC...",
  "certificate_subject": "CN=node-01",
  "certificate_issuer": "CN=CA",
  "certificate_notbefore": "2026-06-01T00:00:00Z",
  "certificate_notafter": "2027-06-01T00:00:00Z",
  "registered_at": "2026-06-14 22:31:48",
  "last_communication": "2026-06-16 09:12:33",
  "has_error_last_report": false,
  "state": "success",
  "configurations": ["Baseline", "Hardening"],
  "tags": {
    "env": ["prod"],
    "role": ["web"]
  }
}
```

### 3.4 GET /api/v1/agents/{id}/configs

Response:

```json
["Baseline", "Hardening"]
```

### 3.5 POST /api/v1/agents/{id}/configs

Request:

```json
{
  "configuration_name": "Baseline"
}
```

Response code attendu: `201 Created`

### 3.6 DELETE /api/v1/agents/{id}/configs

Request:

```json
{
  "configuration_name": "Baseline"
}
```

Response code attendu: `204 No Content`

### 3.7 GET /api/v1/agents/{id}/tags

Response (map string -> array de string):

```json
{
  "env": ["prod"],
  "role": ["web", "api"]
}
```

### 3.8 PUT /api/v1/agents/{id}/tags

Request (ajout d une valeur de tag, operation additive par paire key/value):

```json
{
  "key": "env",
  "value": "prod"
}
```

Response code attendu: `204 No Content`

### 3.9 DELETE /api/v1/agents/{id}/tags

2 modes supportes:

- Supprimer une paire key/value precise via query params:

```http
DELETE /api/v1/agents/{id}/tags?key=env&value=prod
```

- Ou via body:

```json
{
  "key": "env",
  "value": "prod"
}
```

- Si ni key ni value ne sont fournis, supprime tous les tags du node.

Response code attendu: `204 No Content`

### 3.10 DELETE /api/v1/agents/{id}

Response codes:

- `200 OK` si supprime
- `404 Not Found` si deja absent

---

## 4. Resource Terraform/OpenTofu a Generer

Nom resource: `dscpull_node`

Schema cible:

- `id` (Computed): doit contenir `internal_dsc_id`
- `internal_dsc_id` (Computed): identifiant immuable
- `agent_id` (Computed): identifiant runtime non stable
- `node_name` (Required, ForceNew si rename non supporte)
- `state` (Computed)
- `has_error_last_report` (Computed)
- `configs` (Optional, Set(String))
- `tags` (Optional, Map(Set(String)))
- `registration_type` (Computed)
- `lcm_version` (Computed)
- `certificate_thumbprint` (Computed)
- `certificate_subject` (Computed)
- `certificate_issuer` (Computed)
- `certificate_notbefore` (Computed)
- `certificate_notafter` (Computed)
- `registered_at` (Computed)
- `last_communication` (Computed)

---

## 5. Regles CRUD + Import (obligatoires)

### 5.1 Create

1. Appeler `POST /api/v1/agents/preenroll` avec `node_name`.
2. Recuperer le node cree.
3. Resoudre son `internal_dsc_id` (voir section 8 si non expose).
4. Appliquer les `configs` desires:
- pour chaque config du plan, appeler `POST /api/v1/agents/{agent_id}/configs`.
5. Appliquer les `tags` desires:
- pour chaque paire key/value, appeler `PUT /api/v1/agents/{agent_id}/tags`.
6. Faire un Read final et enregistrer `id = internal_dsc_id`.

### 5.2 Read

1. Toujours partir de `internal_dsc_id` du state.
2. Resoudre `agent_id` courant correspondant.
3. Lire details + configs + tags.
4. Si ressource introuvable: remove state.
5. Tolerer le changement TEMP -> GUID sans recreation Terraform.

### 5.3 Update

- `node_name`: ForceNew (si API rename absente).
- `configs`: diff add/remove.
1. add -> `POST /agents/{id}/configs`
2. remove -> `DELETE /agents/{id}/configs`
- `tags`: synchronisation deterministe.
Strategie recommandee:
1. lire tags existants
2. supprimer ecarts (DELETE key/value)
3. ajouter manquants (PUT key/value)

### 5.4 Delete

1. Resoudre `agent_id` courant via `internal_dsc_id`.
2. Appeler `DELETE /api/v1/agents/{agent_id}`.
3. Considerer `404` comme succes idempotent.

### 5.5 Import

- Import ID accepte: `internal_dsc_id`

Commande exemple:

```bash
terraform import dscpull_node.web01 <internal_dsc_id>
```

Import flow:
1. retrouver node par `internal_dsc_id`
2. peupler state complet
3. conserver `id = internal_dsc_id`

---

## 6. Comportement Spécifique Tags

Le backend tags est additif avec `PUT key/value` (pas un remplacement global).
Le provider doit garantir l etat desire complet en calculant les deltas.

Regles:

1. Normaliser les sets pour eviter le drift (tri deterministe).
2. Supporter plusieurs valeurs par key.
3. Eviter les doublons.

Exemple HCL attendu:

```hcl
resource "dscpull_node" "web01" {
  node_name = "node-01"

  configs = [
    "Baseline",
    "Hardening"
  ]

  tags = {
    env  = ["prod"]
    role = ["web", "api"]
  }
}
```

---

## 7. Provider Block et Config

```hcl
terraform {
  required_providers {
    dscpull = {
      source  = "openinfraops/dscpull"
      version = ">= 0.1.0"
    }
  }
}

provider "dscpull" {
  base_url = "https://dsc-dev.local"
  token    = var.dscpull_token
  insecure = true
  timeout  = "30s"
}
```

---

## 8. Exigence Contractuelle internal_dsc_id

Le provider doit etre construit pour utiliser `internal_dsc_id` comme ID Terraform.

Si l API actuelle ne renvoie pas `internal_dsc_id`, tu dois:

1. Proposer le patch backend minimal:
- Ajouter `internal_dsc_id` dans reponses JSON de:
  - `GET /api/v1/agents`
  - `GET /api/v1/agents/{id}`
  - idealement `POST /api/v1/agents/preenroll`

2. Conserver la compatibilite:
- fallback temporaire possible, mais lever un warning explicite si `internal_dsc_id` absent.

3. Garantir le lifecycle:
- `internal_dsc_id` persiste quand `agent_id` change TEMP -> GUID.

---

## 9. Exigences Implementation

- Go + terraform-plugin-framework.
- Client API dedie avec gestion retry/timeouts.
- Erreurs detaillees (status + body).
- Tests acceptance:
1. create/read/delete
2. update configs/tags
3. import by internal_dsc_id
4. disappears out-of-band
- Compatibilite Terraform et OpenTofu.
- README complet + examples.

---

## 10. Livrables Demandes

1. Arborescence projet complete
2. Code provider complet (compilable)
3. Resource `dscpull_node` CRUD+Import
4. Client HTTP API
5. Tests acceptance
6. Documentation utilisateur
7. Exemples Terraform/OpenTofu

Ne fournis pas une spec abstraite: fournis du code concret et executable.
