defmodule SimpleElixirApp.MixProject do
  use Mix.Project

  def project do
    [
      app: :simple_elixir_app,
      version: "0.1.0",
      elixir: "~> 1.14",
      start_permanent: Mix.env() == :prod,
      deps: deps(),
      releases: releases()
    ]
  end

  def application do
    [
      extra_applications: [:logger],
      mod: {SimpleElixirApp.Application, []}
    ]
  end

  defp deps do
    [
      {:bandit, "~> 1.0"},
      {:jason, "~> 1.4"}
    ]
  end

  defp releases do
    [
      simple_elixir_app: [
        include_executables_for: [:unix],
        applications: [runtime_tools: :permanent],
        steps: [:assemble, :tar]
      ]
    ]
  end
end
