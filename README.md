# Shipyard - Automated Deployment Tool for elixir project

`Shipyard` is a command-line tool written in Go, similar to `flyctl`, designed to automate modern smooth deployment (blue-green deployment) workflows. It achieves zero-downtime traffic switching through the Caddy Admin API and supports complex deployment scenarios with multiple applications and hosts.

> 📖 **Quick Reference**: For detailed CLI command documentation and examples, see [CLI_COMMANDS.md](CLI_COMMANDS.md)

## 🏗️ Architecture Overview


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

- **Native Server Efficiency**: Optimized for Elixir/Phoenix projects. Runs native BEAM releases directly on the target server, avoiding the memory footprint of container runtimes. Perfect for maximizing performance on small VPS instances (e.g., 512MB RAM).
- **Docker-Based Local  Builds**: The CLI uses Docker locally to build releases, ensuring a consistent build environment regardless of the developer's OS. You can customize the build process and base environment to ensure consistency with your deployment server, while keeping the production server free of Docker and build tools.
- **Smooth Deployment**: Uses blue-green deployment strategy with Caddy  for traffic switching, achieving zero-downtime updates
- **Multi-App/Multi-Host**: Flexible database structure supporting deployment of multiple independent applications to different VPS hosts, laying groundwork for future load balancing
- **Multi-Domain Support**: Single application instance can be bound to multiple domains (e.g., `example.com` and `www.example.com`), with automatic Caddy reverse proxy configuration
- **Multi-Environment Support**: Use `--config` parameter to support different configuration files, achieving physical-level isolation of Staging/Production environments
- **Global CLI Tool**: Can be installed as a system-level command, callable from any project directory
- **Configuration Separation**: Global configuration (hosts, app definitions) separated from project-local configuration (`shipyard.toml`), clear and organized


## 📚 Multi-Domain and Multi-Environment

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

For detailed documentation, see [Multi-Domain and Multi-Environment Examples](example\deploy_example_phoenix\shipyard.toml).

## 🚀 Quick Start

For detailed installation instructions, environment configuration, and troubleshooting, please refer to the **[Installation Scripts Guide](install-guide.en.md)**.

### 1. Installation Overview

**Shipyard Server:**
The server handles storage, API, and the Web UI. It can be installed automatically via script (recommended) or manually.
- **Quick Script:** `curl -fsSL .../install-shipyard-server.sh | bash`
- **Default Port:** `15678` (API & Web UI)

**Shipyard CLI:**
The lightweight client for managing deployments.
- **Quick Script:** `curl -fsSL .../install-shipyard-cli.sh | bash`
- **First Step:** `shipyard-cli login --endpoint http://your-server:15678`

> 📖 See the **[Installation Guide](scripts/install-guide.en.md)** for full script usage and manual build instructions.

### 2. Configuration Workflow (Three Steps)

Before using the `deploy` command, you need to define your "hosts" and "applications", then "link" them together.

**Step One: Add Host (`add-host`)**

Tell `shipyard-cli` which VPS you want to deploy to. This information will be encrypted and stored in the server database.

```bash
# Add a host named "vps-frankfurt"
shipyard-cli add-host --name vps-frankfurt --addr 192.168.1.100 --user root --password "your_password"
```

**Step Two: Configure Project**

Navigate to the root directory of your Phoenix project and create a `shipyard.toml` file:

`shipyard.toml`:
```toml
# Declare which application this project directory corresponds to
app = "todolist-app"
domains = ["todolist.example.com"]
```

**Step Three: First Deployment (`launch`)**

Use the `launch` command to create the application on the server and perform the first deployment.

```bash
# Creates the app and deploys it to the specified host
shipyard-cli launch --host vps-frankfurt
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

Use `deploy` for subsequent updates after the initial `launch`.

```bash
# Deploy updates to "todolist-app"
# The tool reads the app name from shipyard.toml
shipyard-cli deploy

# You can also explicitly specify the app name, which overrides the setting in shipyard.toml
shipyard-cli deploy --app todolist-app
```

**Deployment Process Overview**:

1.  **Local Build**: The CLI builds the release locally using Docker (ensuring consistent environment).
2.  **Upload & Start**: Artifacts are uploaded to the remote host and started on a new available port.
3.  **Zero-Downtime Switch**: Once healthy, Caddy switches traffic to the new version instantly via API.
4.  **Cleanup**: The old version is gracefully shut down.

> 📚 **Advanced CLI Usage**: For more deployment options, troubleshooting commands, and advanced workflows, refer to [CLI_COMMANDS.md](CLI_COMMANDS.md)

## 🧪 Testing with Example Projects

The repository includes two example projects to help you test and understand the deployment workflow.

### 1. Simple Elixir Project (`example/deploy_example_simple_elixir`)
A basic Elixir application without Phoenix, suitable for testing the fundamental build and release process.

```bash
cd example/deploy_example_simple_elixir

# 1. Initialize configuration (if not already present)
# Ensure shipyard.toml contains: app = "your-app-name"

# 2. Deploy
shipyard-cli deploy --host your-host-name
```

### 2. Phoenix Project (`example/deploy_example_phoenix`)
A complete Phoenix web application, suitable for testing:
- Static asset compilation
- Database migrations (if configured)
- Web server startup and port binding
- Caddy reverse proxy integration

```bash
cd example/deploy_example_phoenix

# 1. Initialize configuration
# Ensure shipyard.toml contains: app = "your-phoenix-app-name"

# 2. Deploy
shipyard-cli deploy --host your-host-name
```

**Note:** Before deploying these examples, ensure you have completed the "Configuration Workflow" (Add Host -> Add App -> Link App) for the respective application names defined in their `shipyard.toml` files.

## 🌐 Web UI

The server provides a modern, feature-rich web management interface built with SolidJS, accessible at `http://your-server:15678`.

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
shipyard-cli login --endpoint http://new-server

# Configuration is saved in ~/.shipyard-cli/config.json
```

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
