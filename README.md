# TorrServer

TorrServer, stream torrent to http

### Installation
Just download server from releases and exec file\
https://github.com/YouROK/TorrServer/releases \
After open browser link http://127.0.0.1:8090 \
On linux systems you may need to set the environment variable before run \
***export GODEBUG=madvdontneed=1***

#### macOS install / configure / uninstall script
Just run in Terminal: `curl -s https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerMac.sh -o installTorrserverMac.sh && chmod 755 installTorrServerMac.sh && sudo ./installTorrServerMac.sh`
Alternative install script for Intel Macs: https://github.com/dancheskus/TorrServerMacInstaller

#### Linux on VPS install / configure / uninstall script
Just run in console: `curl -s https://raw.githubusercontent.com/YouROK/TorrServer/master/installTorrServerLinux.sh | sudo bash`

#### Unofficial TorrServer iocage plugin
On FreeBSD (TrueNAS/FreeNAS) you can use this plugin

https://github.com/filka96/iocage-plugin-TorrServer

### Build
Install golang 1.16+ by instruction: https://golang.org/doc/install \
Goto dir to source\
Run build script under linux build-all.sh\
For build web page need install npm and yarn\
For instal yarn: _npm i -g yarn_ after install npm\
For build android server need android toolchain\
Download android ndk and change NDK_TOOLCHAIN in build.sh to\
path/to/Android/sdk/ndk/ver/toolchains/llvm/prebuilt/platform

#
### Server args:
#### Usage
TorrServer-darwin-arm64 [--port PORT] [--path PATH] [--logpath LOGPATH] [--weblogpath WEBLOGPATH] [--rdb] [--httpauth] [--dontkill] [--ui] [--torrentsdir TORRENTSDIR] [--torrentaddr TORRENTADDR] [--pubipv4 PUBIPV4] [--pubipv6 PUBIPV6] [--searchwa]

#### Options
* --port PORT, -p PORT   
  *                 web server port, default 8090
* --path PATH, -d PATH   
  *                 database dir path
* --logpath LOGPATH, -l LOGPATH
  *                 server log file path
* --weblogpath WEBLOGPATH, -w WEBLOGPATH
  *                 web access log file path
* --rdb, -r              
  *                 start in read-only DB mode
* --httpauth, -a         
  *                 enable http auth on all requests
* --dontkill, -k         
  *                 don't kill server on signal
* --ui, -u               
  *                 open torrserver page in browser
* --torrentsdir TORRENTSDIR, -t TORRENTSDIR
  *                 autoload torrents from dir
* --torrentaddr TORRENTADDR
  *                 Torrent client address (format [IP]:PORT, ex. :32000, 127.0.0.1:32768 etc)
* --pubipv4 PUBIPV4, -4 PUBIPV4
  *                 set public IPv4 addr
* --pubipv6 PUBIPV6, -6 PUBIPV6 
  *                 set public IPv6 addr
* --searchwa, -s         
  *                 search without auth
* --help, -h             
  *                 display this help and exit
* --version              
  *                 display version and exit

#### Development

`swag` must be installed on the system to [re]build Swagger documentation.

```bash
go install github.com/swaggo/swag/cmd/swag@latest
cd server; swag init -g web/server.go

# Documentation can be linted using
swag fmt
```

#
### API

#### API Docs

API documentation is hosted as Swagger format available at path `/swagger/index.html`.

#### API Authentication

The user data file should be located near to the settings. Basic auth, read more in wiki <https://en.wikipedia.org/wiki/Basic_access_authentication>.

* File name: *accs.db*
* JSON file format

```json
{
    "User1": "Pass1",
    "User2": "Pass2"
}
```

#
### Whitelist/Blacklist IP
The lists file should be located near to the settings.

whitelist file name: wip.txt\
blacklist file name: bip.txt

whitelist has prior

Example:\
local:127.0.0.0-127.0.0.255\
127.0.0.0-127.0.0.255\
local:127.0.0.1\
127.0.0.1\
\# at the beginning of the line, comment

#
### MSX Install:
Open msx and goto: Settings -> Start Parameter -> Setup \
Enter current ip address and port of server e.g. _127.0.0.1:8090_

#
### Running in docker
Just run:  `docker run --rm -d --name torrserver -p 8090:8090 ghcr.io/yourok/torrserver:latest` \
For running in persistence mode, just mount volume to container by adding `-v ~/ts:/opt/ts`, where `~/ts` folder path is just example, but you could use it anyway... Result example command: `docker run --rm -d --name torrserver -v ~/ts:/opt/ts -p 8090:8090 ghcr.io/yourok/torrserver:latest` \
Other options:
- add `-e TS_HTTPAUTH=1` and place [auth file](#authorization) into `~/ts/config` forlder for enabling basic auth
- add `-e TS_RDB=1` for enabling `--rdb` flag
- add `-e TS_DONTKILL=1` for enabling `--dontkill` flag
- add `-e TS_PORT=5555` for changind default port to 5555(example), also u need to change `-p 8090:8090` to `-p 5555:5555` (example)
- add `-e TS_CONF_PATH=/opt/tsss` for overriding torrserver config path inside container
- add `-e TS_TORR_DIR=/opt/torr_files` for overriding torrents directory
- add `-e TS_LOG_PATH=/opt/torrserver.log` for overriding log path


Example with full overrided command(on default values):
```
docker run --rm -d -e TS_PORT=5665 -e TS_DONTKILL=1 -e TS_HTTPAUTH=1 -e TS_RDB=1 -e TS_CONF_PATH=/opt/ts/config -e TS_LOG_PATH=/opt/ts/log -e TS_TORR_DIR=/opt/ts/torrents --name torrserver -v ~/ts:/opt/ts -p 5665:5665 ghcr.io/yourok/torrserver:latest

```


#
### Donate:
[PayPal](https://www.paypal.me/yourok) \
[QIWI]((https://qiwi.com/n/YOUROK85) \
[YooMoney](https://yoomoney.ru/to/410013733697114/200) 

SberBank card: **5484 4000 2285 7839**

YooMoney card: **4048 4150 1812 8179**


#
### Thanks to everyone who tested and helped

###### **Anacrolix Matt Joiner** [github.com/anacrolix](https://github.com/anacrolix/)

###### **tsynik** [github.com/tsynik](https://github.com/tsynik)

###### **dancheskus** [github.com/dancheskus](https://github.com/dancheskus)

###### **kolsys** [github.com/kolsys](https://github.com/kolsys)

###### **Tw1cker Руслан Пахнев** [github.com/Nemiroff](https://github.com/Nemiroff)

###### **SpAwN_LMG** [github.com/spawnlmg](https://github.com/spawnlmg)
