import Config

# This file is only used during development and testing
# Production configuration is in runtime.exs

if config_env() == :dev do
  config :simple_elixir_app,
    port: 4000
end
