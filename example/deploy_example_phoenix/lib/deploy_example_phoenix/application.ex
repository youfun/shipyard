defmodule DeployExamplePhoenix.Application do
  # See https://hexdocs.pm/elixir/Application.html
  # for more information on OTP Applications
  @moduledoc false

  use Application

  @impl true
  def start(_type, _args) do
    children = [
      DeployExamplePhoenixWeb.Telemetry,
      DeployExamplePhoenix.Repo,
      {DNSCluster, query: Application.get_env(:deploy_example_phoenix, :dns_cluster_query) || :ignore},
      {Phoenix.PubSub, name: DeployExamplePhoenix.PubSub},
      # Start a worker by calling: DeployExamplePhoenix.Worker.start_link(arg)
      # {DeployExamplePhoenix.Worker, arg},
      # Start to serve requests, typically the last entry
      DeployExamplePhoenixWeb.Endpoint
    ]

    # See https://hexdocs.pm/elixir/Supervisor.html
    # for other strategies and supported options
    opts = [strategy: :one_for_one, name: DeployExamplePhoenix.Supervisor]
    Supervisor.start_link(children, opts)
  end

  # Tell Phoenix to update the endpoint configuration
  # whenever the application is updated.
  @impl true
  def config_change(changed, _new, removed) do
    DeployExamplePhoenixWeb.Endpoint.config_change(changed, removed)
    :ok
  end
end
