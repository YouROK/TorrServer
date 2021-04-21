import React, { useEffect, useRef } from 'react'
import Typography from '@material-ui/core/Typography'

import { getPeerString, humanizeSize } from '../utils/Utils'
import DialogTitle from '@material-ui/core/DialogTitle'
import DialogContent from '@material-ui/core/DialogContent'
import { cacheHost } from '../utils/Hosts'

const style = {
    cache: {
        paddingLeft: "6px",
        paddingRight: "2px",
        lineHeight: "11px",
    },
    piece: {
        width: "12px",
        height: "12px",
        backgroundColor: "#eef2f4",
        border: "1px solid #eef2f4",
        display: "inline-block",
        marginRight: "1px",
    },
    pieceComplete: {
        backgroundColor: "#3fb57a",
        borderColor: "#3fb57a",
    },
    pieceLoading: {
        backgroundColor: "#00d0d0",
        borderColor: "#00d0d0",
    },
    readerRange: {
        borderColor: "#9a9aff !important",
    },
    pieceReader: {
        borderColor: "#000000 !important",
    },
    pieceProgress: {
        position: "relative",
        zIndex: "1",
        backgroundColor: "#009090",

        left: "-1px",
        top: "-1px",
        width: "12px",
    },
}

export default function DialogCacheInfo(props) {
    const [hash] = React.useState(props.hash)
    const [cache, setCache] = React.useState({})
    const timerID = useRef(-1)
    const [pMap, setPMap] = React.useState([])

    useEffect(() => {
        if (hash)
            timerID.current = setInterval(() => {
                getCache(hash, (cache) => {
                    setCache(cache)
                })
            }, 100)
        else clearInterval(timerID.current)

        return () => {
            clearInterval(timerID.current)
        }
    }, [hash, props.open])

    useEffect(()=>{
        if (cache && cache.PiecesCount && cache.Pieces){
            var map = [];
            for (let i = 0; i < cache.PiecesCount; i++) {
                var reader = 0
                var cls = "piece"
                var prc = 0
                if (cache.Pieces[i]) {
                    if (cache.Pieces[i].Completed && cache.Pieces[i].Size >= cache.Pieces[i].Length)
                        cls += " piece-complete"
                    else
                        cls += " piece-loading"
                    prc = (cache.Pieces[i].Size / cache.Pieces[i].Length * 100).toFixed(2)
                }

                cache.Readers.forEach((r, k) => {
                    if (i >= r.Start && i <= r.End && i !== r.Reader)
                        cls += " reader-range"
                    if (i === r.Reader) {
                        cls += " piece-reader"
                    }
                })
                map.push({
                    prc: prc,
                    class: cls,
                    info: i,
                    reader: reader,
                })
            }
            setPMap(map)
        }
    },[cache.Pieces])

    return (
        <div>
            <DialogTitle id="form-dialog-title">
                <Typography fullWidth>
                    <b>Hash </b> {cache.Hash}
                    <br />
                    <b>Capacity </b> {humanizeSize(cache.Capacity)}
                    <br />
                    <b>Filled </b> {humanizeSize(cache.Filled)}
                    <br />
                    <b>Torrent size </b> {cache.Torrent && cache.Torrent.torrent_size && humanizeSize(cache.Torrent.torrent_size)}
                    <br />
                    <b>Pieces length </b> {humanizeSize(cache.PiecesLength)}
                    <br />
                    <b>Pieces count </b> {cache.PiecesCount}
                    <br />
                    <b>Peers: </b> {getPeerString(cache.Torrent)}
                    <br />
                    <b>Download speed </b> {cache.Torrent && cache.Torrent.download_speed ? humanizeSize(cache.Torrent.download_speed) + '/sec' : ''}
                    <br />
                    <b>Status </b> {cache.Torrent && cache.Torrent.stat_string && cache.Torrent.stat_string}
                </Typography>
            </DialogTitle>
            <DialogContent>
                <div className="cache">
                    {pMap.map((itm) => <span className={itm.class} title={itm.info}>{itm.prc>0 && itm.prc<100 && (<div className="piece-progress" style={{height: itm.prc/100*12+"px"}}></div>)}</span>)}
                </div>
            </DialogContent>
        </div>
    )
}

function getCacheMap(cache) {
    if (!cache || !cache.PiecesCount) return ''
    var html = ''
    for (let i = 0; i < cache.PiecesCount; i++) {
        html += "<span class='piece"
        let info = i
        var prcDiv = ""
        if (cache.Pieces && cache.Pieces[i]) {
            let prc = (cache.Pieces[i].Size/cache.Pieces[i].Length*100).toFixed(2)
            let piece = cache.Pieces[i]
            if (piece.Completed && piece.Size >= piece.Length) {
                html += ' piece-complete'
                info += ' 100%'
            }else {
                html += ' piece-loading'
                info += ' ' + prc + '%'
                prcDiv = "<div class='piece-progress' style='height: "+prc+"%;'></div>"
            }
        }
        cache.Readers.forEach((r,k)=> {
            if (i >= r.Start && i <= r.End && i !== r.Reader)
                html += ' reader-range'
            if (i === r.Reader) {
                html += ' piece-reader'
                info += ' reader'
            }
        })
        html += "' title='" + info + "'>"
        html += prcDiv
        html += "</span>"
    }
    return html
}

function getCache(hash, callback) {
    try {
        fetch(cacheHost(), {
            method: 'post',
            body: JSON.stringify({ action: 'get', hash: hash }),
            headers: {
                Accept: 'application/json, text/plain, */*',
                'Content-Type': 'application/json',
            },
        })
            .then((res) => res.json())
            .then(
                (json) => {
                    callback(json)
                },
                (error) => {
                    callback({})
                    console.error(error)
                }
            )
    } catch (e) {
        console.error(e)
        callback({})
    }
}
/*
{
	"Hash": "41e36c8de915d80db83fc134bee4e7e2d292657e",
	"Capacity": 209715200,
	"Filled": 2914808,
	"PiecesLength": 4194304,
	"PiecesCount": 2065,
	"DownloadSpeed": 32770.860273455524,
	"Pieces": {
		"2064": {
			"Id": 2064,
			"Length": 2914808,
			"Size": 162296,
			"Completed": false
		}
	}
}
 */
