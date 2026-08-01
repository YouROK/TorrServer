# TorrServer VLC agent for Linux

This optional agent lets TorrServer start VLC on a Linux computer connected to a TV while the Web UI is controlled from a phone or another computer.

The agent is intentionally small:

- Python standard library only;
- opens VLC in a normal window or fullscreen according to the selected device setting in TorrServer;
- replaces only the VLC process started by this agent;
- supports a bearer token;
- can restrict accepted TorrServer hostnames;
- validates the torrent hash, file index, filename, and stream URL before launching VLC;
- never executes a shell command.

## Requirements

- Linux with a graphical desktop;
- Python 3.10 or newer;
- VLC;
- systemd user services for the installer below.

## Install

Run this as the desktop user who should own the VLC window:

```bash
cd extras/vlc-agent
./install-user-service.sh --listen-lan --allowed-host 192.168.1.10
```

Replace `192.168.1.10` with the hostname or IP address used in TorrServer stream URLs. Repeat `--allowed-host` when the same TorrServer is reached by several names.

The installer:

- copies the agent to `~/.local/lib/torrserver-vlc-agent/`;
- installs a user service in `~/.config/systemd/user/`;
- creates `~/.config/torrserver-vlc-agent.env` with mode `0600`;
- generates a random bearer token;
- enables and starts the service.

Read the generated token from the environment file and register the device in **TorrServer → Settings → Application → Playback devices**:

```text
Name: Living room TV
Agent URL: http://PLAYER_IP:8092
Agent token: value from VLC_AGENT_TOKEN
TorrServer URL for this device: http://TORRSERVER_IP:8090
```

The last field is optional. Set it when the player reaches TorrServer through a different address than the browser, for example when the browser uses a public HTTPS name but the TV computer should stream over the LAN.

## Configuration

Edit:

```text
~/.config/torrserver-vlc-agent.env
```

Then restart:

```bash
systemctl --user restart torrserver-vlc-agent.service
```

Useful options:

```text
VLC_AGENT_HOST=0.0.0.0
VLC_AGENT_PORT=8092
VLC_AGENT_ALLOWED_HOSTS=192.168.1.10,movies.example.net
VLC_AGENT_PLAYER=/usr/bin/vlc
VLC_AGENT_PLAYER_ARGS="--no-one-instance --no-video-title-show --network-caching=3000 --http-reconnect"
```

TorrServer sends the device's **Open VLC in fullscreen** checkbox with every play request. The agent appends either `--fullscreen` or `--no-fullscreen` after the configured arguments, so the checkbox has an unambiguous result and remains off by default.

Additional VLC flags may be appended to `VLC_AGENT_PLAYER_ARGS`. Examples:

- `--qt-dark-palette` for VLC builds that support the dark Qt palette;
- `--aout=pulse --no-spdif --stereo-mode=1` when HDMI passthrough causes audio problems.

## Security

The default listener is `127.0.0.1`. Direct LAN access requires `VLC_AGENT_HOST=0.0.0.0` or `--listen-lan`.

A non-loopback listener refuses to start without a bearer token unless `--allow-unauthenticated-network` is explicitly used. That override is intended only for isolated test networks.

Also restrict TCP port `8092` with the host firewall so only the TorrServer machine can reach it. `VLC_AGENT_ALLOWED_HOSTS` limits the hostname accepted inside `stream_url`; it does not replace a firewall or bearer token.

## Desktop session troubleshooting

The service runs as a systemd **user** service so VLC can connect to that user's display and audio session. On desktops that do not import graphical variables into the user manager, run this once from a terminal inside the desktop session:

```bash
systemctl --user import-environment DISPLAY WAYLAND_DISPLAY DBUS_SESSION_BUS_ADDRESS PULSE_SERVER
systemctl --user restart torrserver-vlc-agent.service
```

Logs and status:

```bash
systemctl --user status torrserver-vlc-agent.service
journalctl --user -u torrserver-vlc-agent.service -f
```

Health check:

```bash
curl -H "Authorization: Bearer TOKEN_FROM_ENV_FILE" http://127.0.0.1:8092/health
```

## Run without installing

The agent can be started directly:

```bash
VLC_AGENT_TOKEN="TOKEN" \
VLC_AGENT_ALLOWED_HOSTS="192.168.1.10" \
python3 torrserver_vlc_agent.py --host 0.0.0.0
```

## Tests

```bash
python3 -m unittest -v test_agent.py
```
