## TorrServer

After starting the container, the latest server is downloaded from GitHub.\
If you need update server to latest, repull container

Source code: https://github.com/YouROK/TorrServer

--------

Author of docker file and scripts [butaford (aka Pavel)](https://github.com/butaford)

--------

### Support platforms
* TorrServer-linux-386
* TorrServer-linux-amd64
* TorrServer-linux-arm5
* TorrServer-linux-arm64
* TorrServer-linux-arm7

--------
### Support env
TS_PORT: TS web port\
TS_PATH: config path and other\
TS_LOGPATHDIR: log path\
TS_LOGFILE: log file name\
TS_WEBLOGFILE: web log file name\
TS_RDB: read only config\
TS_HTTPAUTH: auth for server, accs.db should be in the TS_PATH\
TS_DONTKILL: don't kill server by signal\
TS_TORRENTSDIR: torrents listen directory\
TS_TORRENTADDR: torrents peer listen port\
TS_PUBIPV4: the IP addresses as our peers should see them. May differ from the local interfaces due to NAT or other network configurations\
TS_PUBIPV6: the IP addresses as our peers should see them. May differ from the local interfaces due to NAT or other network configurations\
TS_SEARCHWA: disable auth for search torrents if auth is enable

--------
### Docker run example
```
docker run -p 8090:8090 \
-e TS_PORT=8090 \
-e TS_PATH="/opt/torrserver/config" \
-e TS_LOGPATHDIR="/opt/torrserver/log/" \
-e TS_LOGFILE="ts.log" \
-e TS_WEBLOGFILE="tsweb.log" \
-e TS_RDB=true \
-e TS_HTTPAUTH=true \
-e TS_DONTKILL=true \
-e TS_TORRENTSDIR="/opt/torrserver/torrents" \
-e TS_TORRENTADDR=32000 \
-e TS_PUBIPV4=publicIP \
-e TS_PUBIPV6=publicIP \
-e TS_SEARCHWA=true \
yourok/torrserver
```

--------
### Docker compose example
```
version: '3.6'
services:
  torrserver:
    container_name: torrserver
    image: ghcr.io/yourok/torrserver
    restart: unless-stopped
    environment:
      - TS_PORT=8090
      - TS_PATH=/opt/torrserver/config
      - TS_LOGPATHDIR=/opt/torrserver/log
      - TS_LOGFILE=ts.log
      - TS_WEBLOGFILE=tsweb.log
      - TS_RDB=false
      - TS_HTTPAUTH=true
      - TS_DONTKILL=true
      - TS_TORRENTSDIR=/opt/torrserver/torrents
      - TS_TORRENTADDR=:32000
      - TS_PUBIPV4=publicIP
      - TS_PUBIPV6=publicIP
      - TS_SEARCHWA=true
    ports:
      - 8090:8090
    volumes:
      - ./torrserver/config:/opt/torrserver/config
      - ./torrserver/log:/opt/torrserver/log
      - ./torrserver/torrents:/opt/torrserver/torrents
```
