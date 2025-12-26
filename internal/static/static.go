package static

import _ "embed"

//go:embed Dockerfile.build.phoenix
var DockerfileBuildPhoenix string

//go:embed Dockerfile.build.elixir
var DockerfileBuildElixir string

//go:embed install_caddy_github.sh
var InstallCaddyScript string

//go:embed init_runtime.sh
var InitRuntimeScript string

// DockerfileStatic is a multi-stage Dockerfile for static sites that require frontend build.
// Uses Node.js to build frontend, and Go to compile into a single binary.
//
//go:embed Dockerfile.static
var DockerfileStatic string

// DockerfileStaticSimple is a Dockerfile for pure static files (no build required).
// Directly embeds static files into the Go binary.
//
//go:embed Dockerfile.static.simple
var DockerfileStaticSimple string
