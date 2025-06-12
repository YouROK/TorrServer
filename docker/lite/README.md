## TorrServer

A lightweight container that contains a single TorrServer file

Source code: https://github.com/YouROK/TorrServer

--------

### Support platforms
* TorrServer-linux-386
* TorrServer-linux-amd64
* TorrServer-linux-arm5
* TorrServer-linux-arm64
* TorrServer-linux-arm7

--------
### Docker run example
```
docker run -p 8090:8090 yourok/torrlite:TAG [ ARGS ]
```

TAG - tag of version in docker hub eg MatriX.134 \
ARGS - args of torrserver

You can mount a directory like -v /your/local/path/:/cfg and write logs etc there

Example of run with args:
```
docker run -p 8099:8099 yourok/torrlite:MatriX.134 --port=8099
```