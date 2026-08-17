#!/bin/sh
set -eu

listen_host=127.0.0.1
allowed_hosts=

usage() {
  cat <<'EOF'
Usage: ./install-user-service.sh [--listen-lan] [--allowed-host HOST]

  --listen-lan          listen on 0.0.0.0 instead of loopback
  --allowed-host HOST   allow stream URLs from this TorrServer host; repeatable
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --listen-lan)
      listen_host=0.0.0.0
      shift
      ;;
    --allowed-host)
      [ "$#" -ge 2 ] || { echo "--allowed-host requires a value" >&2; exit 2; }
      if [ -n "$allowed_hosts" ]; then
        allowed_hosts="$allowed_hosts,$2"
      else
        allowed_hosts=$2
      fi
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

command -v python3 >/dev/null 2>&1 || { echo "python3 is required" >&2; exit 1; }
player=$(command -v vlc || true)
[ -n "$player" ] || { echo "VLC is required" >&2; exit 1; }
command -v systemctl >/dev/null 2>&1 || { echo "systemd is required" >&2; exit 1; }

source_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
agent_dir="$HOME/.local/lib/torrserver-vlc-agent"
unit_dir="$HOME/.config/systemd/user"
env_file="$HOME/.config/torrserver-vlc-agent.env"
unit_file="$unit_dir/torrserver-vlc-agent.service"

install -d -m 0755 "$agent_dir" "$unit_dir"
install -m 0755 "$source_dir/torrserver_vlc_agent.py" "$agent_dir/torrserver_vlc_agent.py"
install -m 0644 "$source_dir/torrserver-vlc-agent.service" "$unit_file"

if [ ! -e "$env_file" ]; then
  token=$(python3 -c 'import secrets; print(secrets.token_urlsafe(32))')
  umask 077
  cat >"$env_file" <<EOF
VLC_AGENT_HOST=$listen_host
VLC_AGENT_PORT=8092
VLC_AGENT_TOKEN=$token
VLC_AGENT_ALLOWED_HOSTS=$allowed_hosts
VLC_AGENT_PLAYER=$player
VLC_AGENT_PLAYER_ARGS="--no-one-instance --no-video-title-show --network-caching=3000 --http-reconnect"
VLC_AGENT_STOP_TIMEOUT=3
EOF
  chmod 0600 "$env_file"
else
  echo "Keeping existing configuration: $env_file"
fi

systemctl --user daemon-reload
systemctl --user enable --now torrserver-vlc-agent.service

echo "Installed TorrServer VLC agent."
echo "Configuration: $env_file"
echo "Status: systemctl --user status torrserver-vlc-agent.service"
echo "Use the VLC_AGENT_TOKEN value from the configuration when registering this device in TorrServer."
