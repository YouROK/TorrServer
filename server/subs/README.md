# External subtitles

TorrServer can find subtitle files that lie next to a video inside a torrent and
serve them to clients: converted to WebVTT (default) or in their original format,
in the requested character encoding.

This is useful for players that cannot read subtitle files from a torrent
themselves (web players, Android TV clients, external players that accept
a subtitle URL). It works in all builds and needs no external tools.

## API

```
GET /subs/{hash}/{id}                  list subtitle tracks for the video file {id}
GET /subs/{hash}/{id}?track={index}    return one subtitle file
```

`{hash}` and `{id}` are the same as in `/play/{hash}/{id}`. If the torrent
contains a single file, `{id}` is treated as `1`, as in `/play`. `{id}` must be
a video file (by extension), otherwise the response is `400`.

The same authentication rules as for `/play` apply.

### Track list

```json
[
  {"index": 3, "name": "ep01.eng.ass", "title": "ep01.eng.ass", "lang": "en", "format": "ass", "url": "/subs/<hash>/1?track=3"},
  {"index": 2, "name": "ep01.rus.srt", "title": "ep01.rus.srt", "lang": "ru", "format": "srt", "url": "/subs/<hash>/1?track=2"}
]
```

- `index` - index of the subtitle file in the torrent (use it as `track`);
- `name` - file name;
- `title` - display name: the file name; if several tracks have the same file name, the
  name of their folder is prepended (`Subs [Exmel & Reddog]/ep01.ass` becomes
  `Exmel & Reddog - ep01.ass`; square brackets and the words `Subs`, `Sub`, `Subtitles`
  and `Субтитры` are removed from the folder name);
- `lang` - detected language code or `und` (see [Language detection](#language-detection));
- `format` - format of the source file (`srt`, `ass`, `ssa`, `vtt`);
- `url` - relative link to the track (prepend scheme and host).

The list is sorted by language code, then by title. It is empty (`[]`) if
nothing was found.

### Track

Query parameters (used together with `track`):

| Parameter  | Values                        | Default | Description |
|------------|-------------------------------|---------|-------------|
| `format`   | `vtt`, `original`             | `vtt`   | `vtt` converts the file to WebVTT, `original` returns the file as is (no conversion) |
| `encoding` | `utf8`, `windows-1251` (`cp1251`) | `utf8`  | Character encoding of the response |

The response `Content-Type` is `text/vtt; charset=...` for `format=vtt` and
`text/plain; charset=...` for `format=original`.

Browsers require UTF-8 for WebVTT, so change `encoding` for `format=vtt` only
deliberately. Characters that do not exist in the target encoding are replaced
with `?`.

Examples:

```
/subs/<hash>/1?track=5                                   WebVTT, UTF-8
/subs/<hash>/1?track=5&format=original                   original ASS/SRT, UTF-8
/subs/<hash>/1?track=5&format=original&encoding=cp1251   original, Windows-1251
```

### Status codes

| Code | Reason |
|------|--------|
| 400  | invalid file or track index, `id` is not a video, unknown `format` or `encoding` |
| 401  | authentication required |
| 404  | torrent, video file or track not found |
| 422  | the subtitle file could not be converted (e.g. no dialogue lines) |
| 500  | torrent error (e.g. no metadata, read error) |
| 504  | the subtitle file could not be read from the torrent within 60 seconds |

Reading a subtitle file does not mark it (or the video) as viewed.

## Supported source formats

The format is detected by file extension, not by content.

| Extension      | Conversion to WebVTT | Lost in conversion |
|----------------|----------------------|--------------------|
| `.srt`         | Adds the `WEBVTT` header, converts timestamps (`00:00:01,000` to `00:00:01.000`). Cue numbers and inline tags such as `<i>` and `<b>` are kept as is | nothing |
| `.ass`, `.ssa` | Reads the `[Events]` section: field positions come from the `Format:` line (Start, End, Text), cues are built from `Dialogue:` lines and sorted by start time | styles, colors, fonts, positioning, karaoke |
| `.vtt`         | Returned as is; the `WEBVTT` header is added if missing | nothing |

ASS/SSA details:

- `{...}` override blocks are removed; `\N` and `\n` become line breaks, `\h` a
  space; `&`, `<` and `>` are escaped;
- `Comment:` lines, empty cues, cues with `end <= start` and vector drawings
  (`\p1`) are skipped;
- times are expected as `h:mm:ss.cc`;
- "sign" cues (on-screen text with their own styles) are not filtered out and
  appear as regular cues.

With `format=original` any of the four formats is returned without conversion.

Not supported: `.sub`/`.idx` (VobSub), `.sup` (PGS), `.smi`, `.txt`, and
subtitle tracks embedded in the video container. Files larger than 2 MiB are
ignored.

### Source encodings

Detected automatically:

- UTF-8, with or without BOM (the BOM is removed);
- UTF-8 with a few stray invalid bytes (they are replaced with `U+FFFD`);
- Windows-1251 (when the file is not UTF-8).

Not supported: UTF-16, cp866, KOI8-R and others.

## Finding external subtitles

Paths are the file paths inside the torrent (`path` in the file list).

**Which files are subtitles.** Files with the extensions `.srt`, `.ass`, `.ssa`
and `.vtt` (case-insensitive), up to 2 MiB. Subtitle files that lie in a folder with audio files
(`.mka`, `.aac`, `.ac3`, `.dts`, `.flac`, ...) but without any video are ignored: they are
companions of that audio (for example, "signs" for a voice-over), not tracks of the video.

**Where subtitles may lie relative to the video.** The subtitle directory must
be nested with the video directory in one of these ways:

- the same directory;
- a subdirectory of any depth (`Subs/`, `Subs/Rus/`, ...);
- any parent directory of the video, including the torrent root.

If the video or the subtitle file is in the torrent root, any directory
matches. Sibling directories (neither contains the other) are not matched.

**Which of the found files belong to the video:**

1. If the torrent contains one video file, all found subtitles belong to it.
2. If it contains several videos (files with the extensions `.mkv .mp4 .avi
   .m4v .ts .mov .webm` are counted, including samples), a subtitle matches if
   either:
   - its name without the extension starts with the video name without the
     extension (case-insensitive), and the next character is not a letter or a
     digit: `ep01.mkv` matches `ep01.rus.srt` and `ep01.eng.ass`, but not
     `ep010.srt`;
   - the episode number is the same. Recognized: `S01E02` (also with a space,
     `.`, `_` or `-` between `S01` and `E02`), `1x02`, and a standalone number
     of 1-3 digits delimited by spaces, `.`, `_`, `-` or brackets (the last such
     number in the name is used: `Show - 02` is episode 2). The season is
     compared only if both names contain it. Numbers that belong to a codec name
     (`x.265`, `H.264`) or to a decimal (`5.1`) are ignored, and the episode number is
     used only if it tells the videos apart (no other video has the same number).

### Language detection

The language is taken from the file name first, then from the folder names,
nearest first. Names are split into words on any non-letter character,
case-insensitive.

| `lang` | Words |
|--------|-------|
| `ru`   | `ru`, `rus`, `russian`, `рус`, `русский`, `русские` |
| `en`   | `en`, `eng`, `english`, `англ`, `английский`, `английские` |
| `ja`   | `ja`, `jpn`, `jap`, `japanese`, `японский`, `японские` |

If none of these words is found and the path (the file name or any folder name)
contains Cyrillic letters, the language is assumed to be `ru`. An explicit marker
always wins: `Субтитры/en/ep01.srt`, `Субтитры/ep01.eng.srt` and
`Сериал.ENG/Subs/ep01.srt` are English. Otherwise `lang` is `und`.

### Layout examples

Matched:

```
Movie/Movie.mkv            + Movie/Movie.rus.srt, Movie/Movie.eng.srt   same directory
Movie.mkv                  + Subs/Movie.srt                             video in the root
Show/ep01.mkv              + Show/Subs/Rus/ep01.ass                     nested directory
Show/Season 1/ep01.mkv     + Show/ep01.srt                              parent directory
Show/S01E02.mkv (+ others) + Show/Subs/Show.S01E02.rus.ass              by episode number
```

Not matched:

```
Video/ep01.mkv  + Subs/ep01.srt                             sibling directories
Show/ep01.mkv   + Show/ep01.sub                             unsupported format
Show/ep01.mkv, Show/ep02.mkv + Show/Subs/commentary.srt     several videos, no name or episode match
```

## HLS in GStreamer builds

In `-gst` builds, external subtitles are also offered in the HLS master playlist
(`/gst/{hash}/master.m3u8`) as additional WebVTT renditions, next to the embedded
subtitle tracks. They are controlled by the existing `Subtitles` option of the
GStreamer module.

- The tracks are the ones listed by `GET /subs/{hash}/{id}` for the played video file;
  the module reads them through the loopback address, like it reads the video.
- External tracks get the indexes `1000`, `1001`, ... (embedded tracks use small
  indexes), so their playlists are `/gst/{hash}/subs/1000.m3u8` and so on.
- The whole subtitle file is read once per task (at most 16 tracks, 4 MiB each,
  15 seconds in total) and cut into segments aligned with the video segments.
  A track that cannot be read for a temporary reason (timeout, a server error) is skipped
  and the next master request tries again, at most three times. A track that cannot be
  used at all (for example, an ASS file without dialogue lines, or a file larger than the
  limit) is skipped for good.

External audio tracks (voice-overs in separate files) are described in
[../extfiles/README.md](../extfiles/README.md).
