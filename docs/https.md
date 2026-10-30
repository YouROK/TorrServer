# HTTPS

Start with `--ssl` to serve the web UI and API over HTTPS on `--sslport` (default 8091) alongside plain HTTP on `--port`. Plain HTTP requests sent to the HTTPS port are redirected to `https://`. There are three ways to get a certificate:

| Option | Trusted by browsers | Trusted by players/TVs | Needs |
|---|---|---|---|
| Self-signed (default) | after accepting a warning | no | nothing |
| Let's Encrypt via DNS-01, LAN only | yes | yes (except Android ≤ 7.0) | a domain or free DuckDNS name |
| Reverse proxy (Caddy, nginx) | yes | yes | a domain and a port open to the internet |

## HTTP and HTTPS modes

Four flags decide what the plain HTTP port (`--port`, default 8090) and the HTTPS port (`--sslport`, default 8091) serve:

| Mode | Flags | HTTP port | HTTPS port | Use it when |
|---|---|---|---|---|
| HTTP only (default) | none | everything | not opened | a trusted home network, or behind a [reverse proxy](reverse-proxy.md) |
| HTTP and HTTPS | `--ssl` | everything | everything | browsers use HTTPS, players and TVs keep using HTTP |
| HTTPS, media also on HTTP | `--ssl --force-https --http-media` | media only; everything else redirects to HTTPS | everything | the self-signed certificate, with players and TVs that can't use it |
| HTTPS preferred | `--ssl --force-https` | redirects everything to HTTPS | everything | a trusted certificate; clients that type `http://` are sent to HTTPS |
| HTTPS only | `--ssl --https-only` | not opened | everything | a trusted certificate, and nothing should ever travel unencrypted |

- **Media** means `/stream`, `/play`, `/playlist`, `/playlistall` (also used by DLNA) and GStreamer HLS under `/gst/<hash>/`. The web UI, the API and GStreamer control endpoints (`/gst/settings`, `/gst/remove`, `/gst/echo`) are not media.
- **Redirects** are `307 Temporary Redirect` to the same path and query on `https://<host>:<sslport>`. A request still reaches the HTTP port before it's redirected, so its URL and any Basic auth credentials have already been sent unencrypted. `--https-only` removes that path, because the HTTP port is never opened. The same applies to plain HTTP sent to the HTTPS port (below) in every mode, so always configure clients with the `https://` address.
- **Plain HTTP sent to the HTTPS port** (for example `http://host:8091`) is redirected to `https://` in every mode, instead of failing with "Client sent an HTTP request to an HTTPS server".
- **Links handed to players:** playlists and the web UI's external-player and copy-link buttons use the address the page was opened on. The exception is a page opened over the self-signed certificate while the HTTP port serves media: then they point at the HTTP port, because players reject that certificate. DLNA links point at the HTTP port when it serves media, and at the HTTPS port otherwise. With `--https-only`, Bonjour advertises `_torrserver` on the HTTPS port and doesn't advertise `_http`.
- **TorrServer's own requests** (ffprobe, GStreamer) use an internal listener on a random `127.0.0.1` port that is never redirected, so they work in every mode, including when `--ip` excludes loopback.
- **Self-signed certificate with `--force-https` or `--https-only`:** startup logs a warning, because most players and TVs won't play. Use a trusted certificate, or `--force-https --http-media` on a trusted network.
- **Invalid combinations stop startup:** `--force-https` or `--https-only` without `--ssl`, `--http-media` without `--force-https`, and `--https-only` with `--http-media`. `--force-https` with `--https-only` is allowed and behaves like `--https-only`.
- **Apps that use the API** (Lampa and other TorrServer clients that add torrents, list them or call `/gst/remove`) must be configured with the `https://` address when `--force-https` or `--https-only` is on: HTTP redirects are unreliable for browser API requests, particularly those requiring CORS preflight. `--http-media` only helps players that receive stream links.
- **Docker:** `TS_SSL_ENABLE`, `TS_FORCE_HTTPS_ENABLE`, `TS_HTTP_MEDIA_ENABLE` and `TS_HTTPS_ONLY_ENABLE` set to `1` enable the matching flags.

HTTPS is only on when TorrServer is started with `--ssl`; the choice isn't saved in the settings. The HTTPS port, certificate and key paths are saved, and reused on later starts with `--ssl`.

Examples:

```bash
# HTTPS only, with a trusted certificate
TorrServer --ssl --https-only --sslcert /opt/torrserver/tls/fullchain.pem --sslkey /opt/torrserver/tls/key.pem

# self-signed certificate for browsers, players and TVs on HTTP
TorrServer --ssl --force-https --http-media
```

## Certificate in the web UI

With `--ssl`, **Settings → Additional → HTTPS** (turn on **PRO mode** to see the Additional tab) shows the HTTPS mode and ports, the certificate in use (names, issuer, expiry, whether a trusted CA issued it) and the certificate and key files it's served from. The certificate can be changed there without a restart. The mode and ports stay startup flags, and without `--ssl` the section isn't shown.

- **Upload** a PEM certificate (full chain) and its unencrypted private key from the computer you're browsing on, e.g. `fullchain.pem` and `privkey.pem`. The pair must match and be currently valid. TorrServer copies them to `ssl/uploaded.crt` and `ssl/uploaded.key` in its config folder (the `--path` folder, e.g. `/opt/torrserver` with the Linux install script or `/opt/ts/config` in Docker; the HTTPS section shows the full path) and serves them within a few seconds. The upload sends the private key, so do it over HTTPS or on the TorrServer machine itself. The copy isn't renewed: upload again after renewing, or use the next option.
- **Use these files** takes the full paths of a certificate and key that are already on the TorrServer machine, as TorrServer sees them (inside the container for Docker); nothing is uploaded. Use it for certificates renewed automatically by acme.sh, certbot or another ACME client: TorrServer notices when the files change and serves the renewed certificate without a restart.
- **Use self-signed** switches back to TorrServer's self-signed certificate (the existing one is reused, so devices that trust it keep working) and deletes an uploaded copy.
- **Regenerate** creates a new self-signed certificate and key.
- **Download certificate** saves the certificate in use (never the key), e.g. to trust the self-signed one on your devices.

The same actions are available in the API under `/ssl/` (see `/swagger`). They are not available with `--rdb`, and the certificate can't be changed here when `--sslcert`/`--sslkey` are given, because those flags are applied again on every start; start without them to manage the certificate from the web UI.

## Self-signed certificate

Without `--sslcert`/`--sslkey`, TorrServer generates a self-signed certificate for `localhost`, the hostname, `hostname.local` and the local IPs, and renews it before it expires or when the host moves to a new IP. Global IPv6 addresses are included but don't cause a renewal when they change, as IPv6 privacy extensions rotate them every few hours. Browsers show a warning you can accept once. Most media players, TVs and DLNA renderers reject it, so give them HTTP links: don't use `--force-https`, or add `--http-media` on a trusted network. When a playlist is requested over the self-signed HTTPS port and HTTP still serves media, its links point to the HTTP port. The self-signed certificate is only ever regenerated if it is one TorrServer created; your own certificate is never touched, even at the default location.

## Trusted certificate on your LAN (Let's Encrypt DNS-01)

Let's Encrypt can issue a certificate for a name that points to a **private** IP. It verifies you own the name through a DNS record, so nothing has to be reachable from the internet. Example with [DuckDNS](https://www.duckdns.org) (free) and [acme.sh](https://github.com/acmesh-official/acme.sh):

1. Create a DuckDNS subdomain, e.g. `mytorr.duckdns.org`, and set its IP to TorrServer's LAN IP (e.g. `192.168.1.10`). Give the host a DHCP reservation so that IP doesn't change.
2. Check it resolves on your LAN: `dig +short mytorr.duckdns.org`. If it returns nothing, your router's DNS rebinding protection blocks public names with private IPs; allow the domain there.
3. Issue the certificate and install it where TorrServer reads it:

   ```bash
   DuckDNS_Token=YOUR_TOKEN acme.sh --issue --dns dns_duckdns -d mytorr.duckdns.org --server letsencrypt --dnssleep 180
   acme.sh --install-cert -d mytorr.duckdns.org --key-file /opt/torrserver/tls/key.pem --fullchain-file /opt/torrserver/tls/fullchain.pem
   ```

4. Start TorrServer with the certificate. Use the **full chain** file: TVs and Android players reject a certificate without its intermediate.

   ```bash
   TorrServer --ssl --sslcert /opt/torrserver/tls/fullchain.pem --sslkey /opt/torrserver/tls/key.pem
   ```

   Or start with just `--ssl` and set the same two paths in the web UI under **Use these files** (see [Certificate in the web UI](#certificate-in-the-web-ui)).

5. Open `https://mytorr.duckdns.org:8091` on any device on the LAN. Use the name, not the IP: the IP isn't in the certificate.
6. Optional: add `--https-only` so TorrServer doesn't open the plain HTTP port at all, or `--force-https` to keep it open but redirect it to HTTPS.

acme.sh renews the certificate automatically every ~60 days and rewrites the files; TorrServer picks up the new files within seconds, without a restart. Any ACME client with DNS-01 support (certbot, lego, Caddy with a DNS plugin) works the same way. Note that certificate names are published in public Certificate Transparency logs.

## TorrServer on an Android TV box

A certificate belongs to a name, not to a device or IP, so it doesn't have to be issued on the box. Two practical setups:

**Issue on a computer, push to the box.** Run the DNS-01 steps above on a Mac, Linux or Windows (WSL) machine, point the DuckDNS name at the **box's** LAN IP, and let acme.sh copy each renewed certificate to the box over ADB (enable network debugging on the box):

```bash
acme.sh --install-cert -d mytorr.duckdns.org --key-file ~/tls/key.pem --fullchain-file ~/tls/fullchain.pem --reloadcmd "adb connect 192.168.1.50 && adb push ~/tls/fullchain.pem ~/tls/key.pem /sdcard/torrserver/tls/"
```

Start TorrServer on the box with `--ssl --sslcert /sdcard/torrserver/tls/fullchain.pem --sslkey /sdcard/torrserver/tls/key.pem`. TorrServer reloads the pushed files within seconds, without a restart. The computer has to be on around renewal time (every ~60 days).

**Let another always-on device handle HTTPS.** If you have a NAS, Raspberry Pi or router that can run [Caddy](https://caddyserver.com) (built with the [DuckDNS plugin](https://github.com/caddy-dns/duckdns)), point the name at that device and forward to the box. Caddy obtains and renews the certificate by itself, and the box runs TorrServer without `--ssl`:

```
mytorr.duckdns.org {
    tls {
        dns duckdns YOUR_TOKEN
    }
    reverse_proxy 192.168.1.50:8090 {
        flush_interval -1
    }
}
```

Running acme.sh on the box itself (e.g. in Termux) also works, but Android's storage restrictions and aggressive background-app killing make renewals unreliable. The box's own Android version doesn't matter, because TorrServer brings its own TLS stack; only the devices that *connect* to it must trust Let's Encrypt.
