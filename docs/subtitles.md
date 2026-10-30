# External subtitles

TorrServer can find subtitle files that lie next to a video inside a torrent and serve them to clients, converted to WebVTT or in their original format. It works in all builds and needs no external tools.

```
GET /subs/{hash}/{id}                  list subtitle tracks for the video file {id}
GET /subs/{hash}/{id}?track={index}    return one subtitle file (WebVTT by default)
```

Query parameters for `track`: `format` (`vtt` or `original`) and `encoding` (`utf8` or `windows-1251`).

See [server/subs/README.md](../server/subs/README.md) for the supported formats and encodings, and for the file layouts inside a torrent that are recognized.
