# Shipyard CLI Commands Reference

> 📖 **Getting Started**: New to Shipyard? Start with the [README.md](README.md) for installation instructions and basic setup guide.

This document provides comprehensive documentation for all Shipyard CLI commands, including advanced options and detailed examples.

## Table of Contents

- [Shipyard CLI Commands Reference](#shipyard-cli-commands-reference)
  - [Table of Contents](#table-of-contents)
  - [Global Flags](#global-flags)
  - [Authentication](#authentication)
    - [login](#login)
    - [logout](#logout)
  - [Application Management](#application-management)
    - [deploy](#deploy)
    - [launch](#launch)
    - [status / info](#status--info)
    - [app](#app)
  - [Variable Management](#variable-management)
    - [vars](#vars)
  - [Logs](#logs)
  - [Build Artifacts](#build-artifacts)
    - [build](#build)
  - [Domain Management](#domain-management)
    - [domain](#domain)
  - [Utility](#utility)
    - [version](#version)
    - [help](#help)
  - [Common Workflows](#common-workflows)
    - [First-Time Application Setup](#first-time-application-setup)
    - [Regular Deployment](#regular-deployment)
    - [Managing Environment Variables](#managing-environment-variables)
    - [Troubleshooting](#troubleshooting)
    - [Fast Deployment with Build Reuse](#fast-deployment-with-build-reuse)
  - [Configuration File (shipyard.toml)](#configuration-file-shipyardtoml)
  - [Tips and Best Practices](#tips-and-best-practices)
  - [Testing Status](#testing-status)

---

## Global Flags

Global flags can be used with any command:

```bash
--config <path>   # Config file path (default: shipyard.toml)
--server <url>    # Shipyard Server URL (default: http://localhost:8080)
```

**Examples:**

```bash
# Use a custom config file
shipyard-cli --config shipyard.staging.toml deploy

# Connect to a different server
shipyard-cli --server https://shipyard.example.com status
```

---

## Authentication

### login

Login to the Shipyard Server using device flow authentication. The login session is saved and persists across CLI invocations.

**Usage:**

```bash
shipyard-cli login [--endpoint <url>]
```

**Flags:**

- `--endpoint <url>`: Server endpoint URL (e.g., http://localhost:8080)

**Examples:**

```bash
# Login with interactive endpoint prompt
shipyard-cli login

# Login to a specific server
shipyard-cli login --endpoint http://localhost:8080

# Login to a production server
shipyard-cli login --endpoint https://shipyard.example.com
```

**Process:**

1. CLI requests a device code from the server
2. A browser window opens automatically (or you can manually open the URL)
3. Enter the displayed user code in the browser
4. Authorize the device in the web interface
5. CLI automatically detects authorization and saves the access token

**Output:**

```
=== CLI Authorization ===
User Code: ABCD-EFGH

Please open the following URL in your browser to authorize:
  http://localhost:8080/auth/device?session_id=xxx

Waiting for authorization...

✅ Login successful!
Configuration saved to /home/user/.shipyard/config.json
```

**Status:** ✅ Tested and working

---

### logout

Logout from the Shipyard Server by removing the saved credentials.

**Usage:**

```bash
shipyard-cli logout
```

**Examples:**

```bash
# Logout from current session
shipyard-cli logout
```

**Output:**

```
Logged out successfully
```

**Status:** ✅ Tested and working

---

## Application Management

### deploy

Deploy an application to a remote host. Supports reusing existing build artifacts.

**Usage:**

```bash
shipyard-cli deploy [--app <name>] [--host <name>] [--use-build <identifier>]
```

**Flags:**

- `--app <name>`: Application name (optional, defaults to shipyard.toml)
- `--host <name>`: Host name (optional, defaults to interactive selection)
- `--use-build <identifier>`: Reuse build artifact by MD5 (short), git commit SHA, or version

**Examples:**

```bash
# Deploy with interactive host selection
shipyard-cli deploy

# Deploy to a specific host
shipyard-cli deploy --host vps-frankfurt

# Deploy a specific app to a specific host
shipyard-cli deploy --app chat-app --host vps-frankfurt

# Deploy using an existing build artifact (by version)
shipyard-cli deploy --host vps-frankfurt --use-build 1.0.0

# Deploy using an existing build artifact (by git SHA)
shipyard-cli deploy --host vps-frankfurt --use-build a1b2c3d4e5

# Deploy using an existing build artifact (by MD5 short hash)
shipyard-cli deploy --host vps-frankfurt --use-build f5e4d3c2b1
```

**Process:**

1. Reads app name from `shipyard.toml` (or uses `--app` flag)
2. Prompts for host selection if not specified
3. Builds the application (or reuses existing build with `--use-build`)
4. Uploads artifacts to remote host
5. Performs blue-green deployment with zero downtime
6. Updates Caddy configuration for traffic switching

**Output:**

```
--- 🚀 Starting deployment process ---
Reading configuration from shipyard.toml...
Detected runtime: phoenix
Building release...
✅ Build completed
Uploading artifacts...
✅ Upload completed
Starting new version on port 12345...
✅ New version started
Performing health check...
✅ Health check passed
Switching traffic via Caddy...
✅ Traffic switched
Stopping old version...
✅ Deployment completed successfully
```

---

### launch

Initialize and deploy a new application for the first time. This command handles the entire setup process.

**Usage:**

```bash
shipyard-cli launch [--app <name>] [--host <name>]
```

**Flags:**

- `--app <name>`: Application name (optional, defaults to current directory name or shipyard.toml)
- `--host <name>`: Host to deploy to (optional, defaults to interactive selection)

**Examples:**

```bash
# Launch with interactive prompts
shipyard-cli launch

# Launch with specific app name
shipyard-cli launch --app my-app

# Launch with specific app and host
shipyard-cli launch --app my-app --host vps-frankfurt
```

**Process:**

1. Creates or validates `shipyard.toml` file
2. Registers the app on the server (if not already registered)
3. Prompts for host selection
4. Links the app to the selected host
5. Initializes the remote host environment
6. Executes the first deployment

**Output:**

```
--- 🚀 Starting shipyard launch process (CLI Mode) ---
✅ App 'my-app' registered.
--- Please select a deployment host ---
1. localhost (127.0.0.1) [Local deployment]
2. vps-frankfurt (192.168.1.100:22)
Selection [1-2]: 2
Selected host: 'vps-frankfurt'
✅ App instance linked successfully.
--- ⚙️ Initializing remote host ---
✅ Remote host initialization completed.
--- 🚀 Executing first deployment ---
✅ Application deployed successfully!
--- 🎉 shipyard launch process completed ---
```

---

### status / info

Show the deployment status of the current project's application.

**Usage:**

```bash
shipyard-cli status
shipyard-cli info
```

**Examples:**

```bash
# Show status of app defined in shipyard.toml
shipyard-cli status

# Using info command (same as status)
shipyard-cli info
```

**Output:**

```
--- App Info: chat-app ---
(Domain info is stored on server; use 'shipyard-cli domain check' to view domain config)

--- Deployment Instances ---
- Host: vps-frankfurt, Status: active on port 12345
- Host: vps-london, Status: active on port 12346
```

---

### app

Manage running application instances (restart, stop, status).

**Usage:**

```bash
shipyard-cli app <subcommand> [--app <name>] [--host <name>]
```

**Subcommands:**

- `restart`: Restart application (stop then start)
- `stop`: Stop application
- `status`: View detailed application status

**Flags:**

- `--app <name>`: Application name (optional, defaults to shipyard.toml)
- `--host <name>`: Host name (optional, defaults to interactive selection)

**Examples:**

```bash
# Restart app (interactive host selection)
shipyard-cli app restart

# Restart specific app on specific host
shipyard-cli app restart --app chat-app --host vps-frankfurt

# Stop app
shipyard-cli app stop --app chat-app --host vps-frankfurt

# Check app status
shipyard-cli app status --app chat-app --host vps-frankfurt

# Using shorthand with interactive selection
shipyard-cli app restart
```

**Output (restart):**

```
--- Restarting app 'chat-app' (Host: vps-frankfurt) ---
Current active port: 12345
--- Stopping service (Port 12345) ---
✅ Service stopped
--- Starting service (Port 12345) ---
--- Executing health check ---
✅ Health check passed
✅ App 'chat-app' restarted successfully (Host: vps-frankfurt, Port: 12345)
```

**Output (stop):**

```
--- Stopping app 'chat-app' (Host: vps-frankfurt) ---
Current active port: 12345
✅ App 'chat-app' stopped successfully (Host: vps-frankfurt, Port: 12345)
```

**Output (status):**

```
--- App Status: chat-app ---
Host: vps-frankfurt (192.168.1.100:22)
Active Port: 12345
Service Status: ✅ Running (active)

Rollback Port: 12340
```

---

## Variable Management

### vars

Manage application environment variables and secrets.

**Usage:**

```bash
shipyard-cli vars <subcommand> [args] [--app <name>]
```

**Subcommands:**

- `list`: List all environment variable keys
- `set KEY=VALUE`: Set environment variable
- `unset KEY`: Delete environment variable

**Flags:**

- `--app <name>`: Application name (optional, defaults to shipyard.toml)

**Examples:**

```bash
# List all variables for app in shipyard.toml
shipyard-cli vars list

# List variables for specific app
shipyard-cli vars list --app chat-app

# Set a single variable
shipyard-cli vars set DATABASE_URL="postgres://user:pass@localhost/db"

# Set multiple variables at once
shipyard-cli vars set API_KEY="abc123" SECRET_KEY="xyz789"

# Set variable interactively (will prompt for value)
shipyard-cli vars set DATABASE_URL

# Set variable for specific app
shipyard-cli vars set --app chat-app REDIS_URL="redis://localhost:6379"

# Unset/delete a variable
shipyard-cli vars unset API_KEY

# Unset multiple variables
shipyard-cli vars unset API_KEY SECRET_KEY

# Unset variable for specific app
shipyard-cli vars unset --app chat-app REDIS_URL
```

**Output (list):**

```
--- Environment Variables for 'chat-app' ---

--- From shipyard.toml ---
PORT
MIX_ENV
PHX_HOST

--- Secrets (from Server) ---
DATABASE_URL
SECRET_KEY_BASE
API_KEY
```

**Output (set):**

```
✅ Secret 'DATABASE_URL' set for application 'chat-app'.
```

**Output (unset):**

```
✅ Secret 'API_KEY' deleted for application 'chat-app'.
```

**Notes:**

- Variables stored in `shipyard.toml` are not encrypted
- Variables managed via `vars set` are encrypted on the server
- During deployment, all variables are merged and injected into the application

---

## Logs

View application logs via SSH connection to the remote host.

**Usage:**

```bash
shipyard-cli logs [app-name] [--host <host>] [--port <port>] [--lines N] [-f|--follow] [--no-color]
```

**Flags:**

- `app-name`: Application name (optional, defaults to shipyard.toml)
- `--host <host>`: Host name (optional, defaults to interactive selection)
- `--port <port>`: View logs for specific port (optional, defaults to active port)
- `--lines <N>`: Show last N lines (default: 500)
- `--follow` / `-f`: Follow log output in real-time
- `--color`: Enable color output (default: enabled)
- `--no-color`: Disable color output

**Examples:**

```bash
# View logs with interactive host selection
shipyard-cli logs

# View logs for specific app
shipyard-cli logs chat-app

# View logs for specific host
shipyard-cli logs --host vps-frankfurt

# View last 100 lines
shipyard-cli logs --lines 100

# Follow logs in real-time
shipyard-cli logs --follow
shipyard-cli logs -f

# View logs for specific port
shipyard-cli logs --port 12345

# View logs without colors (useful for piping to files)
shipyard-cli logs --no-color > app.log

# Combined example
shipyard-cli logs chat-app --host vps-frankfurt --lines 200 --follow
```

**Output (static mode):**

```
Fetching logs for chat-app@vps-frankfurt...
Target: Active Port 12345
Mode: Static view (last 500 lines)
--------------------------------------------------------------------------------
2024-01-20 10:30:45 [info] Starting application...
2024-01-20 10:30:46 [info] Database connection established
2024-01-20 10:30:47 [info] Server started on port 12345
2024-01-20 10:31:00 [info] GET /api/health 200 OK
--------------------------------------------------------------------------------
Displayed last 500 lines of logs
```

**Output (follow mode):**

```
Fetching logs for chat-app@vps-frankfurt...
Target: Active Port 12345
Mode: Real-time tracking (press Ctrl+C to exit)
--------------------------------------------------------------------------------
[Real-time log output streaming...]
```

**Notes:**

- Follow mode (`-f`) uses WebSocket connection for real-time streaming
- Static mode fetches logs via SSH using journalctl
- Color output helps distinguish log levels (info=green, warn=yellow, error=red)
- Press Ctrl+C to exit follow mode

---

## Build Artifacts

### build

Manage build artifacts for applications.

**Usage:**

```bash
shipyard-cli build <subcommand> [--app <name>]
```

**Subcommands:**

- `list`: List all build artifacts for an application

**Flags:**

- `--app <name>`: Application name (optional, defaults to shipyard.toml)

**Examples:**

```bash
# List builds for app in shipyard.toml
shipyard-cli build list

# List builds for specific app
shipyard-cli build list --app chat-app
```

**Output:**

```
--- Build Artifacts for app 'chat-app' ---

VERSION          MD5 (short)  GIT COMMIT SHA                           CREATED AT          
--------------------------------------------------------------------------------
1.2.0            a1b2c3d4e5   abc123def456789012345678901234567890   2024-01-20 10:30:45
1.1.0            f5e4d3c2b1   def456789012345678901234567890abc123   2024-01-19 15:20:30
1.0.0            0123456789   789012345678901234567890abc123def456   2024-01-18 09:15:00

Total: 3 build(s)
```

**Notes:**

- Build artifacts are cached on the server
- Use the identifiers (version, git SHA, or MD5) with `deploy --use-build` to reuse builds
- This speeds up deployments by skipping the build step

---

## Domain Management

### domain

Manage domain configurations and check Caddy settings.

**Usage:**

```bash
shipyard-cli domain <subcommand> [--app <name>] [--host <name>]
```

**Subcommands:**

- `check`: Check Caddy configuration for the domain

**Flags:**

- `--app <name>`: Application name (optional, defaults to shipyard.toml)
- `--host <name>`: Host name (optional, defaults to interactive selection)

**Examples:**

```bash
# Check domain config with interactive selection
shipyard-cli domain check

# Check specific app and host
shipyard-cli domain check --app chat-app --host vps-frankfurt
```

**Output:**

```
Connecting to host vps-frankfurt (192.168.1.100)...
Getting Caddy config...
{
  "apps": {
    "http": {
      "servers": {
        "srv0": {
          "routes": [
            {
              "match": [
                {
                  "host": [
                    "chat.example.com"
                  ]
                }
              ],
              "handle": [
                {
                  "handler": "reverse_proxy",
                  "upstreams": [
                    {
                      "dial": "localhost:12345"
                    }
                  ]
                }
              ]
            }
          ]
        }
      }
    }
  }
}
```

**Notes:**

- Shows the complete Caddy configuration as JSON
- Useful for debugging domain routing issues
- Verify that your domains are correctly mapped to ports

---

## Utility

### version

Display the CLI version.

**Usage:**

```bash
shipyard-cli version
```

**Example:**

```bash
shipyard-cli version
```

**Output:**

```
dev
```

or in release builds:

```
v1.0.0
```

---

### help

Display help information about available commands.

**Usage:**

```bash
shipyard-cli help
shipyard-cli --help
shipyard-cli -h
```

**Example:**

```bash
shipyard-cli help
```

**Output:**

```
Shipyard CLI (Client Mode)
Usage: shipyard-cli [global flags] <command> [command flags]

Global Flags:
  --config <path>   Config file path (default: shipyard.toml)
  --server <url>    Shipyard Server URL

Commands:
  login             Login to Shipyard Server
  logout            Logout
  deploy            Deploy application
  launch            Initialize and deploy a new application
  status            Show status of current project application
  vars              Manage application environment variables (list, set, unset)
  logs              View application instance logs
  app               App management commands (restart, stop, status)
  build             Build artifact management commands (list)
  domain            Domain management commands (check)
  version           Show version
  help              Show help
...
```

---

## Common Workflows

### First-Time Application Setup

```bash
# 1. Login to server
shipyard-cli login --endpoint http://localhost:8080

# 2. Navigate to your project directory
cd /path/to/your/app

# 3. Launch the application (handles setup and first deployment)
shipyard-cli launch

# 4. Check status
shipyard-cli status
```

### Regular Deployment

```bash
# Navigate to project directory
cd /path/to/your/app

# Deploy latest changes
shipyard-cli deploy
```

### Managing Environment Variables

```bash
# Set production secrets
shipyard-cli vars set DATABASE_URL="postgres://..."
shipyard-cli vars set SECRET_KEY_BASE="long-random-string"
shipyard-cli vars set API_KEY="your-api-key"

# View all variables
shipyard-cli vars list

# Deploy with new variables
shipyard-cli deploy
```

### Troubleshooting

```bash
# Check app status
shipyard-cli app status

# View recent logs
shipyard-cli logs --lines 200

# Follow logs in real-time
shipyard-cli logs -f

# Restart app if needed
shipyard-cli app restart

# Check Caddy configuration
shipyard-cli domain check
```

### Fast Deployment with Build Reuse

```bash
# List available builds
shipyard-cli build list

# Deploy using existing build (by version)
shipyard-cli deploy --use-build 1.2.0

# Or by git commit
shipyard-cli deploy --use-build abc123def456
```

---

## Configuration File (shipyard.toml)

The `shipyard.toml` file in your project directory defines application settings:

```toml
# Application name (required)
app = "chat-app"

# Domain configuration (optional)
domains = ["example.com", "www.example.com"]
primary_domain = "example.com"

# Environment variables (optional, non-sensitive only)
[env]
MIX_ENV = "prod"
PORT = "4000"
PHX_SERVER = "true"

# Build hooks (optional)
[hooks]
pre_build = ["mix deps.get", "mix assets.deploy"]
post_build = []
pre_deploy = []
post_deploy = []
```

---

## Tips and Best Practices

1. **Always login first**: Run `shipyard-cli login` before using other commands
2. **Use build reuse**: For hotfixes or rollbacks, use `--use-build` to skip the build step
3. **Check logs regularly**: Use `shipyard-cli logs -f` during deployment to catch issues early
4. **Manage secrets properly**: Never commit secrets to `shipyard.toml`, use `vars set` instead
5. **Test in staging first**: Use `--config` to manage multiple environments
6. **Monitor app status**: Run `shipyard-cli app status` after deployments
7. **Keep tokens secure**: Login tokens are stored in `~/.shipyard/config.json` with restricted permissions

---

## Testing Status

| Command | Status | Notes |
|---------|--------|-------|
| login | ✅ Tested | Working correctly |
| logout | ✅ Tested | Working correctly |
| deploy | ✅ Tested | Working correctly |
| launch | ✅ Tested | Working correctly |
| status/info | ✅ Tested | Working correctly |
| app restart | ✅ Tested | Working correctly |
| app stop | ✅ Tested | Working correctly |
| app status | ✅ Tested | Working correctly |
| vars list | ✅ Tested | Working correctly |
| vars set | ✅ Tested | Working correctly |
| vars unset | ✅ Tested | Working correctly |
| logs | ✅ Tested | Both static and follow modes working |
| build list | ✅ Tested | Working correctly |
| domain check | ✅ Tested | Working correctly |
| version | ✅ Tested | Working correctly |
| help | ✅ Tested | Working correctly |

All commands have been verified to work as documented.
