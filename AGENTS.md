# Shipyard Project LLM-Assisted Development Guide

## Project Architecture Overview

```
shipyard/
├── cmd/                      # CLI Entry Points
│   ├── shipyard-cli/         # Lightweight client (communicates via API)
│   └── shipyard-server/      # Server (provides API + Web UI)
├── internal/                 # Internal Packages
│   ├── api/                  # HTTP API Layer
│   │   ├── handlers/         # API Handlers
│   │   ├── middleware/       # Middleware
│   │   ├── response/         # Response Envelopes
│   │   └── utils/            # ID Encoding/Decoding Tools
│   ├── client/               # API Client Wrapper
│   ├── database/             # Database Access Layer
│   │   ├── migrations/       # SQLite/Turso Migration Files
│   │   └── migrations_pg/    # PostgreSQL Migration Files
│   ├── deploy/               # Core Deployment Logic
│   ├── models/               # Data Models
│   ├── caddy/                # Caddy Configuration Management
│   └── cliutils/             # CLI Utility Functions
├── pkg/types/                # Shared DTO Types
└── webui/                    # Frontend (SolidJS + TypeScript)
    └── src/
        ├── api/              # API Layer
        │   ├── services/     # API Services (Categorized by feature)
        │   ├── hooks/        # React Query Hooks
        │   └── client.ts     # Axios Client Configuration
        ├── components/       # UI Components
        ├── routes/           # Routing Pages (File-based routing)
        │   ├── admin/        # Admin Dashboard Pages
        │   ├── login/        # Login Page
        │   └── setup/        # Initial Setup Page
        ├── contexts/         # React Contexts
        ├── i18n/             # Internationalization (en.ts, zh.ts)
        ├── lib/              # Utility Libraries
        └── types/            # TypeScript Type Definitions
```

---

## 1. Modifying Frontend Pages

### Files to Reference

| Purpose | File Path |
|------|----------|
| Check existing page structure | `webui/src/routes/**/*.tsx` |
| Check existing components | `webui/src/components/**/*.tsx` |
| Check API services | `webui/src/api/services/*.ts` |
| Check API hooks | `webui/src/api/hooks/*.ts` |
| Check TypeScript types | `webui/src/types/*.ts` |
| Check i18n strings | `webui/src/i18n/en.ts`, `webui/src/i18n/zh.ts` |

### Modification Workflow

1. **Modify TypeScript Types** ([webui/src/types/app.ts](webui/src/types/app.ts))
   - Add new fields to interface definitions.

2. **Modify API Services** ([webui/src/api/services/*.ts](webui/src/api/services/))
   - Add new API calls here if needed.
   - Available services: `applicationService`, `authService`, `cliService`, `dashboardService`, `setupService`, `sshHostService`, `systemService`.

3. **Modify Page Components** ([webui/src/routes/**/*.tsx](webui/src/routes/))
   - Use SolidJS syntax (`<Show>`, `<For>`, `createSignal`, etc.).
   - Use DaisyUI + Tailwind CSS for styling.
   - Use `@tanstack/solid-query` for data fetching.

4. **Add Internationalization Strings** (Must modify both files)
   - [webui/src/i18n/en.ts](webui/src/i18n/en.ts) - English
   - [webui/src/i18n/zh.ts](webui/src/i18n/zh.ts) - Chinese
   - Format: `category_key: "text"` (e.g., `deployments_port: "Port"`)

### Example: Adding a New Field to the Deployment List

```typescript
// 1. webui/src/types/app.ts - Add type
export interface DeploymentHistory {
  uid: string
  version: string
  port: number  // New field
  // ...
}

// 2. webui/src/i18n/en.ts - Add translation
export default {
  app_detail: {
    // ... other translations
    deployments_port: "Port",
  }
}

// 3. webui/src/i18n/zh.ts - Add translation
export default {
  app_detail: {
    // ... other translations
    deployments_port: "Port",
  }
}

// 4. webui/src/routes/admin/apps/[uid]/page.tsx - Use the field
<th>{t('app_detail.deployments_port')}</th>
<td>{item.port || '-'}</td>
```

---

## 2. Modifying Backend API

### Files to Reference

| Purpose | File Path |
|------|----------|
| Check route definitions | `internal/api/server.go` |
| Check handler implementations | `internal/api/handlers/*.go` |
| Check database operations | `internal/database/*.go` |
| Check data models | `internal/models/models.go` |
| Check DTO types | `pkg/types/models.go` |
| Check ID encoding/decoding | `internal/api/utils/id.go` |
| Check response envelopes | `internal/api/response/*.go` |

### Modification Workflow

1. **Modify Database Layer** ([internal/database/api_helpers.go](internal/database/api_helpers.go) or relevant file)
   - Add new fields to Row structs.
   - Update SQL queries to include new fields.

2. **Modify API Handlers** ([internal/api/handlers/*.go](internal/api/handlers/))
   - Add new fields to the JSON response.
   - Use the `response` package for consistent response formatting.

3. **(Optional) Modify DTO Definitions** ([pkg/types/models.go](pkg/types/models.go))
   - Used for data transfer between CLI client and server.

4. **(Optional) Modify API Client** ([internal/client/client.go](internal/client/client.go))
   - Ensure the client correctly parses the new fields returned by the server.

### Example: Adding a Field to an API Response

```go
// 1. internal/database/api_helpers.go - Modify struct and query
type DeploymentHistoryRow struct {
    ID          string    `db:"id"`
    Version     string    `db:"version"`
    Port        int       `db:"port"`  // New field
    CreatedAt   time.Time `db:"created_at"`
    // ... other fields
}

// Update SQL query to include new field
func (d *Database) GetDeploymentHistory(appID string) ([]DeploymentHistoryRow, error) {
    query := `
        SELECT 
            d.id,
            d.version,
            COALESCE(ai.active_port, 0) as port,  -- New
            d.created_at
        FROM deployments d
        LEFT JOIN application_instances ai ON ai.deployment_id = d.id
        WHERE d.application_id = ?
        ORDER BY d.created_at DESC
    `
    // ...
}

// 2. internal/api/handlers/deployments_handler.go - Return field
func (h *Handler) GetDeploymentHistory(c *gin.Context) {
    // ... query logic
    
    result := make([]gin.H, len(histories))
    for i, history := range histories {
        result[i] = gin.H{
            "uid":        utils.EncodeFriendlyID(utils.PrefixDeployment, history.ID),
            "version":    history.Version,
            "port":       history.Port,  // New field
            "created_at": history.CreatedAt,
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "data": result,
    })
}
```

### ID Encoding/Decoding Rules

The project uses friendly ID formats (e.g., `app_xxx`, `dpl_xxx`) to replace raw UUIDs:

| Prefix | Resource Type | Description |
|------|----------|------|
| `app_` | Application | Application |
| `dpl_` | Deployment | Deployment Record |
| `ssh_` | SSH Host | SSH Host |
| `inst_` | Application Instance | App Instance |
| `env_` | Environment Variable | Environment Variable |

**Usage Example:**

```go
// Encoding: UUID -> Friendly ID
friendlyID := utils.EncodeFriendlyID(utils.PrefixDeployment, history.ID)
// Result: dpl_abc123...

// Decoding: Friendly ID -> UUID
deployID, err := utils.DecodeFriendlyID(utils.PrefixDeployment, uid)
if err != nil {
    return response.Error(c, http.StatusBadRequest, "Invalid deployment ID")
}
```

---

## 3. Modifying Deployment Process

### Files to Reference

| Purpose | File Path |
|------|----------|
| Core deployment logic | `internal/deploy/deploy.go` |
| Build logic | `internal/deploy/build.go` |
| Remote command execution | `internal/deploy/remote.go` |
| Hooks execution | `internal/deploy/hooks.go` |
| Caddy configuration management | `internal/caddy/caddy.go` |
| Domain synchronization | `internal/deploy/domain_sync.go` |
| DTO conversion | `internal/deploy/dto_converters.go` |
