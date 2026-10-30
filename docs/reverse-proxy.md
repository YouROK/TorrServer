# Reverse proxy

If TorrServer is exposed to the internet under a domain, the simplest option is a reverse proxy that handles certificates itself. Never expose the plain HTTP port (`--port`) to the internet: Basic auth credentials and stream URLs would travel unencrypted. Run TorrServer without `--ssl`, enable `--httpauth`, and bind it to loopback with `--ip 127.0.0.1` when the proxy runs on the same host. TorrServer honours `X-Forwarded-Proto`/`X-Forwarded-Host`, so playlist links use the public `https://` name. Disable response buffering, or playback will stutter.

Caddy (obtains and renews Let's Encrypt certificates automatically):

```
tv.example.com {
    reverse_proxy 127.0.0.1:8090 {
        flush_interval -1
    }
}
```

nginx (certificate from certbot or similar):

```nginx
location / {
    proxy_pass http://127.0.0.1:8090;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $host;
    proxy_buffering off;
    proxy_request_buffering off;
    proxy_read_timeout 24h;
    client_max_body_size 50m;
}
```
