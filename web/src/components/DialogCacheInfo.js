import React, { useEffect, useRef } from 'react'
import Typography from '@material-ui/core/Typography'

import { getPeerString, humanizeSize } from '../utils/Utils'
import DialogTitle from '@material-ui/core/DialogTitle'
import DialogContent from '@material-ui/core/DialogContent'
import { cacheHost } from '../utils/Hosts'

export default function DialogCacheInfo(props) {
    const [hash] = React.useState(props.hash)
    const [cache, setCache] = React.useState({})
    const timerID = useRef(-1)
    const canvasRef = useRef(null)
    const dialogRef = useRef(null)

    useEffect(() => {
        const canvas = canvasRef.current
        const context = canvas.getContext('2d')

        if (hash)
            timerID.current = setInterval(() => {
                getCache(hash, (cache) => {
                    setCache(cache);
                    redraw(cache, canvas, context, dialogRef)
                })
            }, 100)
        else clearInterval(timerID.current)

        return () => {
            clearInterval(timerID.current)
        }
    }, [hash, props.open])

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
            <DialogContent ref={dialogRef}>
                <canvas ref={canvasRef} />
            </DialogContent>
        </div>
    )
}

const pieceSize = 12;

const colors = {
    0: "#eef2f4", // empty piece
    1: "#00d0d0", // donwloading
    2: "#009090", // donwloading fill color
    3: "#3fb57a", // downloaded
    4: "#9a9aff", // reader range color
    5: "#000000", // reader current position
}


const map = new Map();
const savedCanvas = document.createElement("canvas");
savedCanvas.ctx = savedCanvas.getContext("2d");



function redraw(cache, canvas, ctx, dialogRef) {
    if (!cache || !cache.PiecesCount || !dialogRef.current) return;
    if(dialogRef.current.offsetWidth !== canvas.width + 50 || cache.PiecesCount !== map.size) {
        canvas.width = dialogRef.current.offsetWidth - 50;
        renderPieces(canvas, ctx, cache.PiecesCount);
    }
    ctx.drawImage(savedCanvas, 0, 0);
    if (cache.Pieces) {
        Object.values(cache.Pieces).forEach(piece => {
            const cords = map.get(piece.Id);
            if (piece.Completed && piece.Size >= piece.Length) {
                ctx.fillStyle = colors[3];
                ctx.beginPath();
                ctx.rect(cords.x, cords.y, pieceSize, pieceSize);
                ctx.fill();
            } else {
                ctx.fillStyle = colors[1];
                ctx.beginPath();
                ctx.rect(cords.x, cords.y, pieceSize, pieceSize);
                ctx.fill();
                const percent = piece.Size / piece.Length
                fillPiece(ctx, piece.Id, percent)
            }
        })
    }
    cache.Readers.forEach(r => setReader(ctx, r.Start, r.Reader, r.End ))
}


function renderPieces(canvas, ctx, count) {
    const horizont = ~~(canvas.width / (pieceSize + 1));
    canvas.height = ~~(count / horizont) * (pieceSize + 1) + pieceSize + 1;
    const vertical = ~~(canvas.height / pieceSize);

    ctx.fillStyle = colors[0];

    map.clear();

    for(let y = 0; y < vertical; y++) {
        for(let x = 0; x < horizont; x++) {
            if(map.size >= count) break;
            map.set(map.size, { x: (pieceSize + 1) * x, y: (pieceSize + 1) * y});
            ctx.beginPath();
            ctx.rect((pieceSize + 1) * x, (pieceSize + 1) * y, pieceSize, pieceSize);
            ctx.fill();
        }
    }
    savedCanvas.width = canvas.width;
    savedCanvas.height = canvas.height;
    savedCanvas.ctx.drawImage(canvas, 0, 0);
}

function fillPiece(ctx, piece, percent) {
    let cords = map.get(piece);
    let offest = pieceSize * percent;
    if(offest < 1) return; //if less than one pixel a spot remains under a piece
    ctx.fillStyle = colors[2];
    ctx.beginPath();
    ctx.rect(cords.x, cords.y + (pieceSize - offest), pieceSize, offest);
    ctx.fill();
}

function setReader(ctx, start, reader, end) {
    ctx.strokeStyle = colors[4];
    for (let i = start; i <= end; i++) {
        cords = map.get(i)
        ctx.beginPath();
        ctx.rect(cords.x + 0.5, cords.y + 0.5, pieceSize - 1, pieceSize - 1);
        ctx.stroke();
    }

    let cords = map.get(reader);
    ctx.strokeStyle = colors[5];
    ctx.beginPath();
    ctx.rect(cords.x + 0.5, cords.y + 0.5, pieceSize - 1, pieceSize - 1);
    ctx.stroke();
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