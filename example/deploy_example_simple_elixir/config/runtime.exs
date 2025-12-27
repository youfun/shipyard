import Config

# Runtime configuration for production
if config_env() == :prod do
  # Port is read from environment variable in application.ex
  # This allows Shipyard to dynamically assign ports
  config :simple_elixir_app,
    port: String.to_integer(System.get_env("PORT") || "4000")
end
