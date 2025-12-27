defmodule DeployExamplePhoenix.Repo do
  use Ecto.Repo,
    otp_app: :deploy_example_phoenix,
    adapter: Ecto.Adapters.SQLite3
end
