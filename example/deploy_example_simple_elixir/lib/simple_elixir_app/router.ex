defmodule SimpleElixirApp.Router do
  import Plug.Conn

  def init(options), do: options

  def call(conn, _opts) do
    case conn.request_path do
      "/" ->
        handle_root(conn)
      "/health" ->
        handle_health(conn)
      "/api/info" ->
        handle_info(conn)
      _ ->
        handle_not_found(conn)
    end
  end

  defp handle_root(conn) do
    html = """
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="UTF-8">
      <title>Simple Elixir App</title>
      <style>
        body {
          font-family: system-ui;
          max-width: 800px;
          margin: 50px auto;
          padding: 20px;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          color: white;
        }
        .card {
          background: rgba(255,255,255,0.1);
          backdrop-filter: blur(10px);
          border-radius: 16px;
          padding: 30px;
          box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.37);
        }
        h1 { margin-top: 0; }
        .badge {
          display: inline-block;
          background: rgba(255,255,255,0.2);
          padding: 4px 12px;
          border-radius: 12px;
          font-size: 14px;
          margin-right: 8px;
        }
        a { color: #ffd700; text-decoration: none; }
        a:hover { text-decoration: underline; }
      </style>
    </head>
    <body>
      <div class="card">
        <h1>✨ Simple Elixir App</h1>
        <p><span class="badge">🚀 Deployed with Shipyard</span> <span class="badge">⚡ Powered by Elixir</span></p>
        <p>This is a minimal Elixir application demonstrating that Shipyard works with <strong>any</strong> Elixir release, not just Phoenix!</p>

        <h2>Features</h2>
        <ul>
          <li>✅ No Phoenix framework - just pure Elixir + Bandit</li>
          <li>✅ Mix release with configurable PORT</li>
          <li>✅ Health check endpoint for zero-downtime deployments</li>
          <li>✅ JSON API endpoint</li>
        </ul>

        <h2>API Endpoints</h2>
        <ul>
          <li><a href="/health">/health</a> - Health check (returns OK)</li>
          <li><a href="/api/info">/api/info</a> - JSON info about the app</li>
        </ul>

        <p style="margin-top: 30px; opacity: 0.8;">
          Version: #{Application.spec(:simple_elixir_app, :vsn)}<br>
          Port: #{System.get_env("PORT") || "4000"}<br>
          Node: #{Node.self()}
        </p>
      </div>
    </body>
    </html>
    """

    conn
    |> put_resp_content_type("text/html")
    |> send_resp(200, html)
  end

  defp handle_health(conn) do
    conn
    |> put_resp_content_type("text/plain")
    |> send_resp(200, "OK")
  end

  defp handle_info(conn) do
    info = %{
      app: "simple_elixir_app",
      version: to_string(Application.spec(:simple_elixir_app, :vsn)),
      elixir_version: System.version(),
      otp_release: :erlang.system_info(:otp_release) |> to_string(),
      node: Node.self(),
      port: System.get_env("PORT") || "4000",
      uptime_seconds: :erlang.statistics(:wall_clock) |> elem(0) |> div(1000),
      deployed_with: "Shipyard 🚢"
    }

    json = Jason.encode!(info, pretty: true)

    conn
    |> put_resp_content_type("application/json")
    |> send_resp(200, json)
  end

  defp handle_not_found(conn) do
    conn
    |> put_resp_content_type("text/plain")
    |> send_resp(404, "Not Found")
  end
end
