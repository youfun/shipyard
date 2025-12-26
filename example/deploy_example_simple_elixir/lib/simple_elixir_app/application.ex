defmodule SimpleElixirApp.Application do
  use Application

  @impl true
  def start(_type, _args) do
    port = String.to_integer(System.get_env("PORT") || "4000")

    children = [
      {Bandit, plug: SimpleElixirApp.Router, port: port}
    ]

    opts = [strategy: :one_for_one, name: SimpleElixirApp.Supervisor]
    IO.puts("🚀 Starting SimpleElixirApp on port #{port}")
    Supervisor.start_link(children, opts)
  end
end
