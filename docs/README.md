# TorrServer documentation

Start with [Installation](installation.md) or [Docker](docker.md), then open <http://127.0.0.1:8090>. The project overview is in the [main README](../README.md).

## Run it

- [Installation](installation.md): Windows, the Linux installer script and its options, macOS, IOCage plugin, NAS systems.
- [Docker](docker.md): `docker run`, the `TS_*` environment variables, Docker Compose.
- [Server arguments](cli.md): every command-line flag.
- [Smart TV](smart-tv.md): Media Station X.

## Secure and expose it

- [HTTPS](https.md): HTTP and HTTPS modes, self-signed and Let's Encrypt certificates, running on an Android TV box.
- [Reverse proxy](reverse-proxy.md): Caddy and nginx in front of TorrServer.
- [Authentication](auth.md): HTTP Basic auth.
- [Web Application Firewall](waf.md): IP and Referer/Origin rules, and their API.

## Tune and operate it

- [Retrackers](retrackers.md): adding trackers to torrents.
- [Memory use](memory.md): why systemd shows more memory than the process holds, and how to check.

## Integrations

- [Torznab](torznab.md): search through Jackett, Prowlarr and similar indexer managers.
- [GStreamer](gstreamer.md): HLS remux and transcoding, the `-gst` builds, runtime requirements.
- [API](api.md): HTTP API, Swagger, and the [MCP server](../server/mcp/README.md).
- [Telegram bot](../server/tgbot/README.md).
- [External subtitles](subtitles.md): formats, encodings and the file layouts that are recognised.

## Develop it

- [Development](development.md): running from source and the web UI ([web/README.md](../web/README.md)).
- [Building](BUILD.md): Makefile targets, GoReleaser, Docker builder, cross-compiling, CI and release.

## Credits

- [Credits and support](credits.md).
