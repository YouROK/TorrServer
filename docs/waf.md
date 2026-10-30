# Web Application Firewall (WAF)

TorrServer includes an HTTP access WAF that filters clients by IP address and by the `Referer` and `Origin` request headers. Configure it from **Settings → Access** or through the authenticated `/waf` API.

## Configuration

WAF configuration is stored in the top-level **`waf`** object in **`settings.json`**. Each rule is a separate array entry:

```json
{
  "waf": {
    "version": 1,
    "whitelist": [
      "127.0.0.1",
      "::1",
      "10.0.0.0/8"
    ],
    "blacklist": [
      "203.0.113.0/24"
    ],
    "referers": [
      "example.com"
    ]
  }
}
```

On first start, if `settings.json` has **no** `waf` key yet and legacy ACL files **`wip.txt`** (whitelist) / **`bip.txt`** (blacklist) exist in the config directory (same place as `config.db`), TorrServer imports them into `waf` arrays and renames the sources to **`wip.txt.bak`** / **`bip.txt.bak`**. Those backups are not read again. If a `waf` key already exists (even with empty lists), legacy files are left untouched.

Changes saved through the web UI or API are applied immediately. After editing `settings.json` manually, restart TorrServer to load the changes.

## IP rules

Rules:

- If the whitelist is **not empty**, the client IP must match it.
- If the blacklist is **not empty**, a matching client IP is banned even when it is also on the whitelist.
- An empty whitelist or blacklist disables that IP check.
- Invalid entries are skipped and reported as warnings; valid entries remain active.
- Banned responses use HTTP **403** with body `Banned`.
- Client IP is taken from the TCP peer address (`RemoteAddr`). Reverse-proxy headers are not trusted by default.

Supported array-entry formats include IPv4, IPv6, ranges, CIDR blocks, comments, and optional descriptions:

```json
[
  "# comment",
  "127.0.0.1",
  "local:127.0.0.1",
  "127.0.0.0-127.0.0.255",
  "local:127.0.0.0-127.0.0.255",
  "10.0.0.0/8",
  "lan:10.0.0.0/8",
  "2001:db8::1",
  "local:2001:db8::1",
  "2001:db8::/32"
]
```

## Referer and Origin rules

Block HTTP requests that come from unwanted sites (for example mirror pages that embed your TorrServer streams).

- Each entry is a hostname. URLs with only an HTTP/HTTPS scheme and host are also accepted.
- A rule blocks the hostname and all its subdomains.
- Both `Referer` and `Origin` are checked before the IP allowlist, so an IP whitelist match cannot bypass a referer rule.
- Requests without either header are allowed.
- Protected internal rules remain active and are not displayed in the web UI.

```json
{
  "referers": [
    "example.com",
    "evil.example.org",
    "# comment"
  ]
}
```

## API

`GET /waf` returns the active editable lists, status flags, and parse warnings. `POST /waf` atomically replaces all three editable lists and hot-reloads the WAF. The API uses newline-delimited strings for compatibility with the web text editors; all fields are required and an empty string clears a list.

```shell
curl -u USER:PASSWORD http://127.0.0.1:8090/waf

curl -u USER:PASSWORD \
  -H 'Content-Type: application/json' \
  -d '{"whitelist":"127.0.0.1\n::1\n10.0.0.0/8","blacklist":"","referers":"example.com"}' \
  http://127.0.0.1:8090/waf
```

In read-only mode, `GET /waf` remains available but `POST /waf` returns HTTP **403**.

> **Note:** BitTorrent peer IP filtering uses a separate PeerGuardian-style file named `blocklist` in the config directory. That list is not managed by Settings → Access / `/waf`.
