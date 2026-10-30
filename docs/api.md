# API

## API Docs

API documentation is hosted as Swagger format available at path `/swagger/index.html`.

## MCP (AI agents)

TorrServer exposes a native [Model Context Protocol](https://modelcontextprotocol.io/) server at **`/mcp`** on the same HTTP(S) port as the web UI (default `8090`). OpenClaw, Hermes, and other MCP clients can list, add, and manage torrents, and get a play URL for the next unwatched TV episode. The REST API is unchanged.

Endpoint: `http://<host>:8090/mcp` (or `https://` when `--ssl` is enabled).

When HTTP auth is on (`-a` / `TS_HTTPAUTH=1`), MCP uses the same Basic credentials as the rest of the API (`accs.db`). Play links returned by tools are ordinary HTTP URLs for VLC, mpv, or a browser. Reverse proxies that connect to `127.0.0.1` with a public `Host` header (Cloudflare Tunnel, nginx, Caddy) are supported.

**OpenClaw** (`openclaw.json`):

```json
{
  "mcp": {
    "servers": {
      "torrserver": {
        "url": "http://127.0.0.1:8090/mcp",
        "transport": "streamable-http"
      }
    }
  }
}
```

With auth, add `"headers": { "Authorization": "Basic <base64-user-pass>" }`.

**Hermes** (`~/.hermes/config.yaml`):

```yaml
mcp_servers:
  torrserver:
    url: "http://127.0.0.1:8090/mcp"
    headers:
      Authorization: "Basic <base64-user-pass>"
```

See [server/mcp/README.md](../server/mcp/README.md) for the tool list and next-unwatched behavior.
