# Deploy Example Phoenix

This is a **test project** for the [Shipyard](https://github.com/youfun/shipyard) CLI tool. It provides a minimal Phoenix application used to validate and demonstrate the deployment workflow.

## Purpose

This project serves as an integration test target for the Shipyard CLI, verifying:
- Blue-green deployment strategy
- Caddy reverse proxy configuration
- Database migrations via hooks
- Environment variable injection
- SQLite database path permissions
- Zero-downtime deployments

## Features

- 🎯 **Phoenix Framework** - Full-featured web application framework
- 🗄️ **SQLite Database** - Embedded database with migrations
- 🔄 **Database Hooks** - Automated migration execution
- 💚 **Health Check** - Built-in health endpoint for deployments
- 🌐 **Multi-domain** - Support for multiple domains
- 🔧 **Environment Variables** - Production configuration

## Prerequisites

- A configured Shipyard server (running on `http://127.0.0.1:8080` by default)
- A registered SSH host in Shipyard
- Phoenix and Elixir installed locally (for development)

## Deploy with Shipyard

### 1. Authenticate with the Shipyard Server

```bash
shipyard-cli login --server http://127.0.0.1:8080
```

### First-time Deployment

For the **first deployment**, use the `launch` command to create and deploy the application:

```bash
cd deploy_example_phoenix

# First-time deployment with default configuration
shipyard-cli launch

# Or specify a custom configuration file
shipyard-cli --config shipyard.staging.toml launch
```

The `launch` command will:
- Create the application in Shipyard
- Build the Phoenix release
- Upload the artifact to the remote host
- Set up domains and SSL certificates
- Configure environment variables
- Ensure database directory permissions
- Start the new version on a standby port
- Run the `migrate` hook
- Perform a health check
- Configure the reverse proxy

### Subsequent Deployments

After the first deployment, use the `deploy` command for updates:

```bash
# Deploy updates to existing application
shipyard-cli deploy

# Or with custom configuration
shipyard-cli --config shipyard.staging.toml deploy
```

The `deploy` command will:
1. Build the Phoenix release
2. Upload the artifact to the remote host
3. Sync domains from `shipyard.toml` to the database
4. Set up environment variables
5. Ensure database directory permissions
6. Start the new version on a standby port
7. Run the `migrate` hook
8. Perform a health check
9. Switch traffic via Caddy

### Deployment Commands Summary

| Command | Usage | Description |
|---------|-------|-------------|
| `launch` | First deployment only | Creates app + initial deployment |
| `deploy` | Subsequent deployments | Updates existing app |

### 3. Check Deployment Status

```bash
shipyard-cli info
```

## Configuration

The deployment is configured via `shipyard.toml`:

```toml
app = 'deploy_example_phoenix'

domains = ["test1.exmaple.com", "test2.exmaple.com"]

keep_releases = 3

[env]
  PHX_SERVER = true
  DATABASE_PATH = "/var/lib/app1/deploy_example_phoenix.db"

[hooks]
  [[hooks.migrate]]
    name = "migrate_database"
    type = "shell"
    command = "bin/migrate"
```

## Endpoints

- `GET /` - Phoenix home page
- `GET /health` - Health check endpoint (returns "OK")
- Phoenix LiveView and other routes as configured

## Local Development

To run the Phoenix server locally:

```bash
# Install dependencies and setup database
mix setup

# Start the Phoenix server
mix phx.server
```

Visit [`localhost:4000`](http://localhost:4000) in your browser.

## Build Release Locally

```bash
# Build a production release
MIX_ENV=prod mix release

# Run the release
PORT=4000 DATABASE_PATH=./dev.db _build/prod/rel/deploy_example_phoenix/bin/deploy_example_phoenix start
```


## Project Structure

```
deploy_example_phoenix/
├── mix.exs                      # Project configuration with release setup
├── shipyard.toml                # Shipyard deployment configuration
├── config/
│   ├── config.exs               # Base configuration
│   ├── dev.exs                  # Development configuration
│   ├── prod.exs                 # Production configuration
│   └── runtime.exs              # Production runtime config
├── priv/
│   └── repo/migrations/         # Database migrations
└── lib/
    └── deploy_example_phoenix/
        ├── application.ex       # OTP Application
        └── ...                  # Phoenix application code
```

## Learn More

- [Phoenix Framework](https://www.phoenixframework.org/)
- [Phoenix Deployment Guides](https://hexdocs.pm/phoenix/deployment.html)
- [Shipyard CLI Documentation](../../README.md)
