# Simple Elixir App - Shipyard Deployment Example

This is a minimal Elixir application (without Phoenix) that demonstrates Shipyard's ability to deploy **any** Elixir release.

## Purpose

This project serves as a lightweight deployment example for the Shipyard CLI, demonstrating:
- Pure Elixir application deployment
- Zero-downtime blue-green deployments
- Health check integration
- Simple HTTP API serving
- Minimal resource usage

## Features

- 🎯 **Pure Elixir** - No Phoenix framework
- 🚀 **Bandit Web Server** - Fast HTTP/1.1 and HTTP/2 support
- 🔧 **Mix Release** - Standard Elixir release configuration
- 💚 **Health Check** - `/health` endpoint for zero-downtime deployments
- 📊 **JSON API** - Simple `/api/info` endpoint

## Prerequisites

- A configured Shipyard server (running on `http://127.0.0.1:15678` by default)
- A registered SSH host in Shipyard
- Elixir installed locally (for development)

## Deploy with Shipyard

### 1. Authenticate with the Shipyard Server

```bash
shipyard-cli login --server http://127.0.0.1:15678
```

### First-time Deployment

For the **first deployment**, use the `launch` command to create and deploy the application:

```bash
cd deploy_example_simple_elixir

# First-time deployment with default configuration
shipyard-cli launch

# Or specify a custom configuration file
shipyard-cli --config shipyard.staging.toml launch
```

The `launch` command will:
- Create the application in Shipyard
- Build and deploy the first version
- Set up domains and SSL certificates
- Configure the reverse proxy
- Start the application on the target host

### Subsequent Deployments

After the first deployment, use the `deploy` command for updates:

```bash
# Deploy updates to existing application
shipyard-cli deploy

# Or with custom configuration
shipyard-cli --config shipyard.staging.toml deploy
```

The `deploy` command will:
- Build the new version
- Perform zero-downtime blue-green deployment
- Update the running application
- Switch traffic seamlessly

### Deployment Commands Summary

| Command | Usage | Description |
|---------|-------|-------------|
| `launch` | First deployment only | Creates app + initial deployment |
| `deploy` | Subsequent deployments | Updates existing app |

### Check Deployment Status

```bash
shipyard-cli info
```

## Configuration

The deployment is configured via `shipyard.toml`:

```toml
app = 'deploy_example_simple_elixir'

domains = ["simple.example.com"]

keep_releases = 3

[env]
  PORT = 4000
  MIX_ENV = "prod"
```

## Local Development

```bash
# Install dependencies
mix deps.get

# Run in development
mix run --no-halt

# Or use iex
iex -S mix
```

Visit http://localhost:4000

## Build Release Locally

```bash
# Build a production release
MIX_ENV=prod mix release

# Run the release
PORT=4000 _build/prod/rel/simple_elixir_app/bin/simple_elixir_app start
```

## Endpoints

- `GET /` - Home page with app info
- `GET /health` - Health check (returns "OK")
- `GET /api/info` - JSON information about the running app

## What This Demonstrates

✅ **Shipyard works with any Elixir release**, not just Phoenix  
✅ **First-time setup** with `launch` command  
✅ **Automatic port management**  
✅ **Zero-downtime blue-green deployments** with `deploy`  
✅ **Health check integration**  
✅ **Simple configuration** with `shipyard.toml`  
✅ **Lightweight deployment** for microservices

## Project Structure

```
deploy_example_simple_elixir/
├── mix.exs                      # Project configuration with release setup
├── shipyard.toml                # Shipyard deployment configuration
├── config/
│   ├── config.exs               # Base configuration
│   └── runtime.exs              # Production runtime config
└── lib/
    └── simple_elixir_app/
        ├── application.ex       # OTP Application
        └── router.ex            # HTTP request router
```


Both deploy exactly the same way with Shipyard!

## Learn More

- [Elixir Releases](https://hexdocs.pm/mix/Mix.Tasks.Release.html)
- [Bandit Web Server](https://github.com/mtrudel/bandit)
- [Shipyard CLI Documentation](../../README.md)
