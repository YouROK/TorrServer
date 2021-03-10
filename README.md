## TorrServer
TorrServer, stream torrent to http

### Installation
Just download from releases and exec file
https://github.com/YouROK/TorrServer/releases
After open browser link http://127.0.0.1:8090


#
### Server args:
#### Usage
TorrServer [--port PORT] [--path PATH] [--logpath LOGPATH] [--rdb] [--httpauth] [--dontkill] [--ui]

#### Options
* --port PORT, -p PORT             web server port
* --path PATH, -d PATH             database and settings path
* --logpath LOGPATH, -l LOGPATH    log path
* --rdb, -r                        start in read-only DB mode
* --httpauth, -a                   http auth on all requests
* --dontkill, -k                   dont kill server on signal
* --ui, -u                         run page torrserver in browser
* --version                        display version and exit

###


### Http Api of TorrServer:
#### GET

###### /echo 
*Return version of server*

###### /shutdown 
*Shutdown server*

###### /stream...
#### args:
* link - magnet/hash/link to torrent
* index - index of file
* preload - preload torrent
* stat - return stat of torrent
* save - save to db
* m3u - return m3u
* fromlast - return m3u from last play
* play - start stream torrent
* title - set title of torrent
* poster - set poster link of torrent

##### Examples:
>**get stat**
>
>http://127.0.0.1:8090/stream/fname?link=...&stat
>
>**get m3u**
>
>http://127.0.0.1:8090/stream/fname?link=...&index=1&m3u
>http://127.0.0.1:8090/stream/fname?link=...&index=1&m3u&fromlast
>
>**stream torrent**
>
>http://127.0.0.1:8090/stream/fname?link=...&index=1&play
>http://127.0.0.1:8090/stream/fname?link=...&index=1&play&save
>http://127.0.0.1:8090/stream/fname?link=...&index=1&play&save&title=...&poster=...
>
>**only save**
>
>http://127.0.0.1:8090/stream/fname?link=...&save&title=...&poster=...

###### /playlistall/all.m3u
*Get all http links of all torrents in m3u list*

###### /playlist
*Get http link of torrent in m3u list*
#### args:
* hash - hash of torrent
* fromlast - from last play file

#
#### POST
###### /torrents
##### Send json:
{\
    "action": "add/get/rem/list/drop",\
    "link": "hash/magnet/link to torrent",\
    "hash": "hash of torrent",\
    "title": "title of torrent",\
    "poster": "link to poster of torrent",\
    "data": "custom data of torrent, may be json",\
    "save_to_db": true/false\
}
##### Return json of torrent(s)

###### /torrent/upload
##### Send multipart/form data
Only one file support
#### args:
* title - set title of torrent
* poster - set poster link of torrent
* data - set custom data of torrent, may be json
* save - save to db

###### /cache
##### Send json:
{\
    "action": "get"\
    "hash" : ""hash": "hash of torrent",\
}
##### Return cache stat 
https://github.com/YouROK/TorrServer/blob/d36d0c28f805ceab39adb4aac2869cd7a272085b/server/torr/storage/state/state.go

###### /settings
##### Send json:
{\
    "action": "get/set/def",\
    _fields of BTSets_\
}
##### Return json of BTSets
https://github.com/YouROK/TorrServer/blob/d36d0c28f805ceab39adb4aac2869cd7a272085b/server/settings/btsets.go

###### /viewed
##### Send json:
{\
    "action": "set/rem/list",\
    "hash": "hash of torrent",\
    "file_index": int, id of file,\
}
##### Return
if hash is empty, return all viewed files\
if hash is not empty, return viewed file of torrent 
##### Json struct see in
https://github.com/YouROK/TorrServer/blob/d36d0c28f805ceab39adb4aac2869cd7a272085b/server/settings/viewed.go

###

### Donate:
[PayPal](https://www.paypal.me/yourok)

[YooMoney](https://yoomoney.ru/to/410013733697114/200)  
YooMoney card: 5599 0050 6424 4747

SberBank card: 4276 4000 6707 2919

###

### Thanks to everyone who tested and helped

###### **Anacrolix Matt Joiner** [github.com/anacrolix](https://github.com/anacrolix/)

###### **tsynik** [github.com/tsynik](https://github.com/tsynik)

###### **Tw1cker Руслан Пахнев** [github.com/Nemiroff](https://github.com/Nemiroff)

###### **SpAwN_LMG**
