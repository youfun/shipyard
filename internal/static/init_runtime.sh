#!/usr/bin/env bash
set -euo pipefail

APP="${APP:-chat_room}"
USER="${USER:-phoenix}"
RUNTIME="${RUNTIME:-phoenix}" # phoenix|elixir|node|golang|static
START_CMD="${START_CMD:-}" # Optional: Override start command

# 1) Create user and directories
if ! id -u "$USER" >/dev/null 2>&1; then
  useradd --system --home "/var/www/$APP" --shell /usr/sbin/nologin "$USER"
fi
mkdir -p "/var/www/$APP/releases" "/var/www/$APP/instances"
chown -R "$USER:$USER" "/var/www/$APP"

# 2) Environment file (common variables)
mkdir -p "/etc/$APP"
if [ ! -f "/etc/$APP/env" ]; then
  touch "/etc/$APP/env"
fi
chown root:"$USER" "/etc/$APP/env"
chmod 0640 "/etc/$APP/env"

# 3) Generate systemd template unit
UNIT_PATH="/etc/systemd/system/$APP@.service"
cat > "$UNIT_PATH" <<'UNIT'
[Unit]
Description=%APP% instance %i
After=network.target
PartOf=%APP%.target

[Service]
User=%USER%
Group=%USER%
WorkingDirectory=/var/www/%APP%/instances/%i
EnvironmentFile=/etc/%APP%/env
Environment=PORT=%i
Environment=PHX_SERVER=true
Restart=always
RestartSec=5s
Type=simple
LimitNOFILE=65536
StandardOutput=journal
StandardError=journal
SyslogIdentifier=%APP%-%i

# ExecStart/ExecStop will be replaced below based on RUNTIME

[Install]
WantedBy=multi-user.target
UNIT

# 4) Inject ExecStart/ExecStop based on runtime
case "$RUNTIME" in
  phoenix)
    # Phoenix releases have a 'foreground' command that runs in foreground
    sed -i "/# ExecStart\/ExecStop/a ExecStart=/bin/sh -lc 'if [ -n \"$START_CMD\" ]; then exec $START_CMD; elif [ -x ./bin/server ]; then exec ./bin/server; else exec ./bin/$APP foreground; fi'\nExecStop=/var/www/$APP/instances/%i/bin/$APP stop" "$UNIT_PATH"
    ;;
  elixir)
    # Plain Elixir releases use 'start' which runs in foreground when called directly
    sed -i "/# ExecStart\/ExecStop/a ExecStart=/bin/sh -lc 'if [ -n \"$START_CMD\" ]; then exec $START_CMD; elif [ -x ./bin/server ]; then exec ./bin/server; else exec ./bin/$APP start; fi'\nExecStop=/var/www/$APP/instances/%i/bin/$APP stop" "$UNIT_PATH"
    ;;
  node)
    sed -i "/# ExecStart\/ExecStop/a ExecStart=/bin/sh -lc 'if [ -n \"$START_CMD\" ]; then exec $START_CMD; elif command -v pnpm >/dev/null 2>&1 && [ -f package.json ]; then exec pnpm start; elif command -v yarn >/dev/null 2>&1 && [ -f package.json ]; then exec yarn start; elif command -v npm >/dev/null 2>&1 && [ -f package.json ]; then exec npm start -- --port=$PORT; elif [ -f server.js ]; then exec node server.js; else echo \"No start command found (set START_CMD or provide scripts.start/server.js)\"; exit 1; fi" "$UNIT_PATH"
    ;;
  golang)
    sed -i "/# ExecStart\/ExecStop/a ExecStart=/bin/sh -lc 'if [ -n \"$START_CMD\" ]; then exec $START_CMD; elif [ -x ./bin/server ]; then exec ./bin/server -port $PORT; elif [ -x ./bin/$APP ]; then exec ./bin/$APP -port $PORT; elif [ -x ./$APP ]; then exec ./$APP -port $PORT; else echo \"No Go binary found (set START_CMD or provide ./bin/server|./bin/$APP|./$APP)\"; exit 1; fi" "$UNIT_PATH"
    ;;
  static)
    # Static sites use a single Go-compiled binary server
    # Multi-stage Docker build has embedded static files into the Go binary
    # VPS needs no Node.js, Go, or Docker, just run the binary
    sed -i "/# ExecStart\/ExecStop/a ExecStart=/bin/sh -lc 'if [ -n \"$START_CMD\" ]; then exec $START_CMD; elif [ -x ./server ]; then exec ./server; else echo \"No static server binary found (expected ./server)\"; exit 1; fi" "$UNIT_PATH"
    ;;
  *)
    echo "Unsupported RUNTIME: $RUNTIME" >&2; exit 1;
    ;;
 esac

# 5) Replace template variables
sed -i "s/%APP%/$APP/g" "$UNIT_PATH"
sed -i "s/%USER%/$USER/g" "$UNIT_PATH"

# If START_CMD is provided and contains ' start', adjust Type to forking (better for daemonized processes)
if [ -n "$START_CMD" ] && echo "$START_CMD" | grep -qE "\bstart\b"; then
  sed -i "s/^Type=simple/Type=forking/" "$UNIT_PATH"
fi

# 6) Apply configuration
systemctl daemon-reload
# Optional: systemctl enable $APP@.service (for instances use systemctl enable $APP@PORT)

echo "Initialized runtime=$RUNTIME unit=$UNIT_PATH"
