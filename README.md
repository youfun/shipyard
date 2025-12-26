# Shipyard - Automated Deployment Tool for Go Projects

`Shipyard` is a command-line tool written in Go, similar to `flyctl`, designed to automate modern smooth deployment (blue-green deployment) workflows. It achieves zero-downtime traffic switching through the Caddy Admin API and supports complex deployment scenarios with multiple applications and hosts.

> 📖 **Quick Reference**: For detailed CLI command documentation and examples, see [CLI_COMMANDS.md](CLI_COMMANDS.md)

## 🏗️ Architecture Overview

Shipyard uses a client-server architecture (see [REFACTOR_PLAN.md](docs/REFACTOR_PLAN.md) for details):

- **shipyard-server**: Server component responsible for data storage (SQLite/PostgreSQL/Turso), API endpoints, Web UI hosting, and secret management
- **shipyard-cli**: Lightweight client that communicates with the server to retrieve configurations and execute deployments

### Build Instructions

```bash
# Build the server
go build -o shipyard-server ./cmd/shipyard-server

# Build the client
go build -o shipyard-cli ./cmd/shipyard-cli
```

### Deployment Architecture

```
Developer's Local Environment
    │
    ├─ shipyard-cli (Client)
    │       │
    │       └─ Communicates via HTTP API
    │              │
    │              ▼
    └─ shipyard-server (Server)
            │
            ├─ SQLite/PostgreSQL/Turso (Data Storage)
            ├─ Web UI (Browser Access)
            └─ SSH Connection to Remote Hosts for Deployment
```

## ✨ Key Features

- **Smooth Deployment**: Uses blue-green deployment strategy with Caddy Admin API for traffic switching, achieving zero-downtime updates
- **Multi-App/Multi-Host**: Flexible database structure supporting deployment of multiple independent applications to different VPS hosts, laying groundwork for future load balancing
- **Multi-Domain Support**: Single application instance can be bound to multiple domains (e.g., `example.com` and `www.example.com`), with automatic Caddy reverse proxy configuration
- **Multi-Environment Support**: Use `--config` parameter to support different configuration files, achieving physical-level isolation of Staging/Production environments
- **Port Management**: Automatically finds available ports on remote hosts to avoid port conflicts
- **Global CLI Tool**: Can be installed as a system-level command, callable from any project directory
- **Configuration Separation**: Global configuration (hosts, app definitions) separated from project-local configuration (`shipyard.toml`), clear and organized
- **Deployment History**: Automatically records detailed information for each deployment (version, status, logs) for auditing and troubleshooting

## 📚 New Features: Multi-Domain and Multi-Environment

### Multi-Environment Support

Use different configuration files to manage different environments:

```bash
# Production environment
shipyard deploy

# Staging environment  
shipyard --config shipyard.staging.toml deploy

# Development environment
shipyard --config shipyard.dev.toml deploy
```

Each configuration file corresponds to an independent application with its own database records, Secrets, and host bindings.

### Multi-Domain Support

Configure multiple domains in `shipyard.toml`:

```toml
app = "myapp"
domains = ["example.com", "www.example.com", "api.example.com"]
primary_domain = "example.com"
```

During deployment, automatically:
- Syncs domains to the database
- Configures Caddy reverse proxy for all domains
- Sets PHX_HOST environment variable for Phoenix multi-domain support

For detailed documentation, see [Multi-Domain and Multi-Environment Examples](docs/MULTI_DOMAIN_ENV_EXAMPLES.md).

## 🚀 Quick Start

> 💡 **Tip**: For detailed installation instructions and troubleshooting, see [Installation Scripts Guide](scripts/README.md)

### 1. Server Installation and Configuration

**Method 1: One-line Installation (Recommended)**

```bash
# Download and install from GitHub Release
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-server.sh | bash

# Or specify a version
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-server.sh | bash -s v1.0.0
```

> 📖 See [scripts/README.md](scripts/README.md) for installation modes, configuration details, and troubleshooting.

**Method 2: Manual Installation**

**Prerequisites**:
- Go 1.24+
- Caddy v2 installed and running on your remote host
- Caddy Admin API port (`:2019`) not blocked by firewall or network policies

**Step 1: Configure Environment Variables (Required for First Use)**

```bash
# Copy the environment variable example file
cp .env_example .env

# Edit the .env file and set JWT_SECRET (used for token signing, ensures login state persists after restarts)
# Generate a random key (recommended)
openssl rand -base64 32

# .env file example:
# JWT_SECRET=your-random-secret-key-here
# DATABASE_TYPE=sqlite  # or postgres, turso
# DATABASE_URL=./deploy.db
```

**Step 2: Start the Server**

```bash
# Build and start the server
go build -o shipyard-server ./cmd/shipyard-server
./shipyard-server --port 8080

# The server will be accessible at:
# - API: http://localhost:8080
# - Web UI: http://localhost:8080
```

### 2. Client Installation and Login

**Method 1: One-line Installation (Recommended)**

```bash
# Automatic detection of OS and architecture, download and install
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-cli.sh | bash

# Or specify a version
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-cli.sh | bash -s v1.0.0

> 📖 See [scripts/README.md](scripts/README.md) for installation paths, PATH configuration, and platform support.

# After installation, you can use shipyard-cli directly
shipyard-cli --version
```

**Method 2: Manual Build**

**Step 1: Build the Client**

```bash
go build -o shipyard-cli ./cmd/shipyard-cli
```

**Step 2: Log in to the Server**

```bash
# First-time use requires login
./shipyard-cli login --endpoint http://your-server:8080

# Test connection
./shipyard-cli ping
```

The client saves authentication information to `~/.shipyard-cli/config.json`, and subsequent commands will use it automatically.

### 3. Configuration Workflow (Three Steps)

Before using the `deploy` command, you need to define your "hosts" and "applications", then "link" them together.

**Step One: Add Host (`add-host`)**

Tell `shipyard-cli` which VPS you want to deploy to. This information will be encrypted and stored in the server database.

```bash
# Add a host named "vps-frankfurt"
shipyard-cli add-host --name vps-frankfurt --addr 192.168.1.100 --user root --password "your_password"
```

**Step Two: Add Application (`add-app`)**

Define an abstract application and its domain.

```bash
# Define an application named "todolist-app" with domain todolist.example.com
shipyard-cli add-app --name todolist-app --domain todolist.example.com
```

**Step Three: Link Application to Host (`link-app`)**

This is the key step that creates a deployable "instance", telling `shipyard-cli` that "todolist-app" will run on "vps-frankfurt".

```bash
shipyard-cli link-app --app todolist-app --host vps-frankfurt
```

### 4. Configure in Your Phoenix Project

Navigate to the root directory of your Phoenix project you want to deploy and create a `shipyard.toml` file with the following content:

`shipyard.toml`:
```toml
# Declare which application this project directory corresponds to
app = "todolist-app"
```

## 💡 Usage Guide

After completing the above configuration, you can perform various operations in your Phoenix project root directory.

> 💡 **Complete Command Reference**: This section covers basic usage. For comprehensive CLI documentation, advanced options, and detailed examples, see [CLI_COMMANDS.md](CLI_COMMANDS.md)

### View Application Status (`info`)

Run this command in the project directory to quickly view the application's deployment status.

```bash
shipyard-cli info
```

The tool will read `shipyard.toml` and display the application's domain, all deployment instances, and their latest deployment status.

### Manage Sensitive Variables (`secrets`)

`shipyard-cli` provides a secure secret management system where secret values are AES-encrypted before being stored in the database.

**Set Secrets:**
```bash
# Set database connection string and API key for "todolist-app"
shipyard-cli secrets set --app todolist-app DATABASE_URL="postgres://..." API_KEY="abc123xyz"
```

**List Secrets (Only shows keys, not values):**
```bash
shipyard-cli secrets list --app todolist-app
```

**Delete Secrets:**
```bash
shipyard-cli secrets unset --app todolist-app API_KEY
```

During deployment, these secrets are automatically written to the `.env` file in the remote release directory for your application to load.

### Deploy Application (`deploy`)

```bash
# Deploy "todolist-app" to the "vps-frankfurt" instance
# The tool reads the app name from shipyard.toml
shipyard-cli deploy --host vps-frankfurt

# You can also explicitly specify the app name, which overrides the setting in shipyard.toml
shipyard-cli deploy --app todolist-app --host vps-frankfurt
```

**Deployment Process Overview**:

1.  **Read Configuration**: `shipyard-cli` reads `shipyard.toml` to get the application name and retrieves detailed information about the application and host from the server API
2.  **Read Version**: Extracts the version number from the `mix.exs` file in the current directory
3.  **Build**: Executes `mix release` to build the project
4.  **Upload**: Packages and uploads the build artifacts to a versioned directory on the remote host (e.g., `/var/www/todolist-app/releases/0.1.0-1667888888`)
5.  **Port Selection**: Finds an available port on the remote host (e.g., `12345`) as the "green" port
6.  **Start New Version**: Starts the new version on the green port (`export PORT=12345 && ./bin/server`)
7.  **Health Check**: Performs a health check on the new version
8.  **Traffic Switching**: If healthy, switches traffic for `todolist.example.com` to the new green port `12345` via Caddy Admin API
9.  **Status Update**: Updates the database via API, setting the instance's `active_port` to `12345`
10. **Decommission Old Version**: Gracefully shuts down processes running on the old port

> 📚 **Advanced CLI Usage**: For more deployment options, troubleshooting commands, and advanced workflows, refer to [CLI_COMMANDS.md](CLI_COMMANDS.md)

## 🌐 Web UI

The server provides a modern, feature-rich web management interface built with SolidJS, accessible at `http://your-server:8080`.

**Key Features:**

- **Dashboard**: Overview of your system status.
- **Application Management**:
    - Manage applications (creation currently supported via CLI only).
    - **Deployment History**: View detailed logs and status of past deployments.
    - **Environment Variables**: Securely manage encrypted Secrets and encrypted environment config.
    - **Domain Management**: View application domain bindings.
    - **View Logs**: View real-time logs of deployment instances.
    <!-- - **Tokens**: Generate and manage API tokens for applications. -->
- **Host Management**:
    - Add and manage SSH hosts (VPS).
    - Connection status monitoring.
- **User Settings**: Manage your profile, password, and Web UI domain configuration.

## 🧪 Running Tests

Run in the `shipyard` project root directory:

```bash
go test ./...
```

## ❓ FAQ

### Why do I need to log in again every time I restart the server?

**Problem Cause**: The JWT_SECRET environment variable is not set, causing the server to generate a new signing key each time it restarts, invalidating all previously issued tokens.

**Solution**:
1. Copy the environment variable example file:
   ```bash
   cp .env_example .env
   ```

2. Edit the `.env` file and set JWT_SECRET:
   ```bash
   # Generate a random key (recommended method)
   openssl rand -base64 32
   
   # Or use any long string
   JWT_SECRET=your-very-secure-random-string-here
   ```

3. After restarting the server, as long as JWT_SECRET remains unchanged, user login status will remain valid (within the token validity period).

**Notes**:
- JWT_SECRET must be kept confidential and not leaked or committed to version control systems
- Production environments should use a random string of at least 32 bytes
- Changing JWT_SECRET will require all existing users to log in again

### How does the client connect to different servers?

```bash
# Change server address
shipyard-cli login --endpoint http://new-server8080

# Configuration is saved in ~/.shipyard-cli/config.json
```
