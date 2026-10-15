# External audio tracks

TorrServer can find audio files that belong to a video file: voice-overs and other
dubbing tracks that lie next to the video in separate files (for example
`Sound [Group]/ep01.mka`). The list is meant for clients that can load an external audio
track by URL (players with a "load audio" function, web and TV clients).

The files are matched by package `extfiles`, which also decides which subtitle files belong
to a video (see [../subs/README.md](../subs/README.md)).

```
GET /audio/{hash}/{id}    list audio tracks for the video file {id}
```

`{hash}`, `{id}` and the authentication rules are the same as for `/subs`.

```json
[
  {"index": 31, "name": "ep01.mka", "title": "JAM & Nika Lenina - ep01.mka", "lang": "und", "format": "mka", "size": 34821530, "url": "/stream?link=<hash>&index=31&play"},
  {"index": 29, "name": "ep01.mka", "title": "Shuuma & Gomer - ep01.mka", "lang": "und", "format": "mka", "size": 35111119, "url": "/stream?link=<hash>&index=29&play"}
]
```

- `index` - index of the audio file in the torrent;
- `name`, `title` - as for subtitles; the words `Sound`, `Audio`, `Dub`, `Озвучка`, `Звук`
  and a few similar ones are removed from the folder name that is prepended to the title;
- `lang` - detected language code or `und` (same rules as for subtitles);
- `format` - the file extension: `mka`, `aac`, `ac3`, `eac3`, `dts`, `flac`, `mp3`, `ogg`,
  `opus`, `wav`, `m4a`, `thd`;
- `size` - file size in bytes;
- `url` - relative link to the file itself (prepend scheme and host).

The list is sorted by language code, then by title, and is empty (`[]`) if nothing was found.

**Which audio files belong to the video.** The folder rules are those of subtitles (the same
directory, a subdirectory, or a parent directory of the video), and with several videos in the
torrent the file name or the episode number must match too. With one video only the rules are
stricter than for subtitles, so that a folder with the soundtrack does not turn into dubbing
tracks. An audio file is accepted when

- its name starts with the video name (`Movie.mkv` and `Movie.Rus.ac3`), or
- it lies in the same folder as the video, or
- it lies below the folder of the video, in a folder named like an audio folder
  (`Audio/Rus/dub.ac3`, `Sound/dub.mka`, `Озвучка/dub.ac3`).

`Movie/OST/01 Theme.flac` and `Movie/Music/cover.mp3` are not matched.

Companion subtitles of a voice-over (for example, "signs" in `Sound [Group]/ep01.ass`) are still
not listed by `/subs`.

## Playback in the GStreamer HLS module

The tracks of this list can be played through the GStreamer HLS module
(`/gst/{hash}/master.m3u8?index={id}`). The `audio` parameter selects the track:
`audio=1000` is the first track of the `/audio/{hash}/{id}` list, `1001` the second, and so
on. Small values (`0`, `1`, ...) are still the audio tracks inside the video file.

- The audio file is read from the torrent by a source of its own and re-encoded to AAC; the
  video part of the pipeline is the same as for embedded tracks.
- A seek restarts the pipeline, because both sources have to be repositioned together.
- An `audio` value outside of the list is answered with 404.
