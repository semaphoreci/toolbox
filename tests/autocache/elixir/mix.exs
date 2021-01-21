defmodule Toolbox.Mixfile do
  use Mix.Project

  def project do
    [
      app: :toolbox,
      version: "0.0.1",
      deps: deps()
    ]
  end

  def application do
    []
  end

  defp deps do
    [
      {:uuid, "~> 1.1"}
    ]
  end
end
