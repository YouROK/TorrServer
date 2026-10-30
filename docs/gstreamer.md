# GStreamer

GStreamer adds HLS output for Matroska/WebM torrents and, when enabled, AVI files. Compatible video and AAC audio can be remuxed without quality loss; unsupported streams can be transcoded to H.264/AAC.

## The `-gst` binary

GStreamer support is available only in a binary compiled with the `gst` build tag. There is no runtime switch that can add `/gst/*` routes to a standard binary.

| Binary | GStreamer |
| --- | --- |
| `TorrServer-gst-<os>-<arch>` | Yes - `/gst/*` routes, remuxing, and transcoding |
| `TorrServer-<os>-<arch>` | No - `GET /gst/settings` reports `built_in: false` |

Use the `--gst` option with the Linux/macOS install scripts, download a matching asset from [releases](https://github.com/YouROK/TorrServer/releases), or build from the repository root:

```bash
cd server
CGO_ENABLED=0 go build -tags="nosqlite gst" -trimpath -o TorrServer-gst ./cmd
```

Supported gst targets are Windows amd64, Linux amd64/arm64, and macOS amd64/arm64. A Windows single-file build with an embedded runtime additionally uses the `embed_gstlib` tag and `server/gstreamer/gst-libs/win-x86_64`; see [`.github/scripts/build-gst-windows.ps1`](../.github/scripts/build-gst-windows.ps1).

## Runtime loading

The gst build uses `purego` to load GStreamer at runtime. It is built with `CGO_ENABLED=0` and does not link GStreamer into the executable at build time.

| Runtime mode | Platforms | How it works |
| --- | --- | --- |
| System or custom path | Windows, Linux, macOS | Loads libraries and plugins from `GSTPath` or a platform installation |
| Portable `gst-lib/` beside the executable | Windows amd64 | Uses the normal GStreamer directory layout without embedding it |
| Embedded `gst-lib` | Windows amd64 | Extracts the runtime from the executable into a versioned TorrServer cache directory on first use |

Runtime roots are tried in this order:

1. `GSTPath` from settings.
2. Platform defaults such as `/usr`, `/usr/local`, the macOS framework/Homebrew prefixes, or the Windows MinGW install directory.
3. Windows only: `gst-lib/` beside the executable.
4. Windows only: the embedded runtime, when compiled with `embed_gstlib`.
5. The operating-system loader search path (`PATH`, `LD_LIBRARY_PATH`, or `DYLD_LIBRARY_PATH`).

Linux and macOS do not use a portable `gst-lib` directory. Install GStreamer system-wide or set `GSTPath`. When GStreamer loads successfully, its real version from `gst_version()` takes precedence over the configured fallback version.

## Runtime requirements

TorrServer requires **GStreamer 1.22 or newer** and `gst-discoverer-1.0`. The full package set is recommended because the pipeline may need HTTP/TLS support, container demuxers, codec parsers, audio decoders, `avenc_aac`, and `x264enc` depending on the source and configuration.

`HDRToSDR` additionally requires the custom `hdrtonemap` GStreamer element. It is included in the embedded Windows runtime. A system/custom runtime must provide a compatible plugin through its normal plugin directory or `GST_PLUGIN_PATH`.

## Installing GStreamer

Installation is required for Linux/macOS gst builds and Windows gst builds that do not embed or ship `gst-lib`.

**Debian / Ubuntu**

```bash
sudo apt-get update

sudo apt-get install -y --no-install-recommends \
  libgstreamer1.0-0 \
  libgstreamer-plugins-base1.0-0 \
  gstreamer1.0-plugins-base \
  gstreamer1.0-plugins-good \
  gstreamer1.0-plugins-bad \
  gstreamer1.0-plugins-base-apps \
  gstreamer1.0-plugins-ugly \
  gstreamer1.0-libav \
  gstreamer1.0-tools \
  ocl-icd-libopencl1 \
  ca-certificates
```

Package roles:

| Package group | Purpose |
| --- | --- |
| `libgstreamer1.0-0`, `libgstreamer-plugins-base1.0-0` | Core and GstApp libraries loaded by TorrServer |
| `gstreamer1.0-plugins-base`, `gstreamer1.0-plugins-base-apps` | Base elements and `gst-discoverer-1.0` |
| `gstreamer1.0-plugins-good` | HTTP source, Matroska/WebM demuxing, and common media elements |
| `gstreamer1.0-plugins-bad` | Modern codec parsers, timestamp helpers, and additional format support |
| `gstreamer1.0-plugins-ugly` | `x264enc` CPU fallback for H.264 transcoding |
| `gstreamer1.0-libav` | `avenc_aac` and FFmpeg-based codec support |
| `gstreamer1.0-tools` | `gst-inspect-1.0` diagnostics |
| `ocl-icd-libopencl1` | OpenCL loader for optional GPU HDR tone mapping; CPU fallback is used when unavailable |
| `ca-certificates` | TLS certificate validation for HTTPS sources |

The selected repository must provide GStreamer 1.22 or newer.

Hardware encoders are optional. They also require a compatible vendor driver and GStreamer encoder plugin; TorrServer tests the available candidates and falls back to `x264enc` when none can start. Candidates include NVENC, QSV, AMF, D3D11/MF, D3D12, and VAAPI (`vah264enc`) on Linux. On Linux, install something like `gstreamer1.0-vaapi` (or the distro equivalent) and a working GPU driver (`vainfo` is useful for checks); TorrServer probes what can start and uses `x264enc` if none succeed.

**Fedora / RHEL / Rocky / AlmaLinux**

```bash
sudo dnf install -y \
  gstreamer1 \
  gstreamer1-plugins-base \
  gstreamer1-plugins-base-tools \
  gstreamer1-plugins-good \
  gstreamer1-plugins-bad-free \
  gstreamer1-plugins-ugly-free \
  gstreamer1-plugin-libav \
  ocl-icd \
  ca-certificates
```

`gstreamer1` provides `gst-inspect-1.0` and related tools; `gstreamer1-plugins-base-tools` provides `gst-discoverer-1.0`. On RHEL and derivatives, `gstreamer1-plugin-libav` may require [EPEL](https://docs.fedoraproject.org/en-US/epel/). Full `x264enc` support may require [RPM Fusion](https://rpmfusion.org/) or the equivalent repository for the distribution.

**Arch Linux**

```bash
sudo pacman -S --needed \
  gstreamer \
  gst-plugins-base \
  gst-plugins-good \
  gst-plugins-bad \
  gst-plugins-ugly \
  gst-libav \
  ocl-icd \
  ca-certificates
```

**macOS**

Install the [official GStreamer Runtime](https://gstreamer.freedesktop.org/download/#macos), or use the current Homebrew formula, which includes the GStreamer plugin sets:

```bash
brew install gstreamer
```

Set `GSTPath` when auto-detection cannot find the installation. Common roots are `/Library/Frameworks/GStreamer.framework/Versions/1.0`, `/opt/homebrew`, and `/usr/local`.

**Windows**

The embedded `TorrServer-gst-windows-amd64.exe` needs no separate GStreamer installation. For a dynamic build, install the **MinGW x86_64 Runtime** from [gstreamer.freedesktop.org](https://gstreamer.freedesktop.org/download/#windows). The runtime installer is sufficient; development files are not required to run TorrServer.

The default root is:

```text
C:\Program Files\gstreamer\1.0\mingw_x86_64
```

Alternatively, place the same runtime layout in `gst-lib/` beside the executable. Do not mix an MSVC installation path with the MinGW libraries or the bundled MinGW `hdrtonemap` plugin.

## Verifying the installation

Check the runtime and discoverer first:

```bash
gst-inspect-1.0 --version
gst-discoverer-1.0 --version
gst-inspect-1.0 souphttpsrc
gst-inspect-1.0 matroskademux
gst-inspect-1.0 mp4mux
gst-inspect-1.0 appsink
gst-inspect-1.0 avenc_aac
gst-inspect-1.0 x264enc
```

`avenc_aac` is needed when the selected audio is not already AAC. `x264enc` is the CPU fallback when video transcoding is enabled. Check `gst-inspect-1.0 hdrtonemap` only when `HDRToSDR` is required.

The shell commands require the GStreamer `bin` directory on `PATH`. For an embedded Windows build, use `GET /gst/echo` or the **GStreamer** settings tab instead. The health response reports `found`, `available`, and `works` for the native runtime, `gst-discoverer`, HDR tone mapping, and the embedded runtime where applicable.

## Web UI configuration

1. Open **Settings**.
2. Enable **PRO mode**.
3. Open the **GStreamer** tab.
4. Adjust options and click **Save GStreamer Settings**.

Settings are stored separately from the main BitTorrent settings and take effect for new tasks. Existing tasks keep the configuration with which they were created.

## `settings.json` block

GStreamer options are stored under the top-level `gstreamer` key in `settings.json`. Legacy keys `gst` and `GStreamer` are still read on load; saving from the web UI or API writes the `gstreamer` key.

Example (`settings.json`):

```json
{
  "gstreamer": {
    "GSTVersion": 1.22,
    "GSTPath": "",
    "Source": "stream",
    "MaxTasks": 0,
    "InactiveMinutes": 5,
    "AACBitrateKbps": 256,
    "AACChannels": 0,
    "AACSamplerate": 0,
    "SegmentSeconds": 6,
    "SegmentDiff": 20,
    "Subtitles": true,
    "TranscodeH264": false,
    "TranscodeH265": false,
    "TranscodeAV1": false,
    "TranscodeVP9": false,
    "TranscodeVP8": false,
    "TranscodeAVI": false,
    "HDRToSDR": false,
    "HardwareAcceleration": true,
    "UseGPU": true,
    "X264Ultrafast": false,
    "VideoBitrate": 10000
  }
}
```

On Windows, platform defaults use `GSTVersion: 1.28` and `GSTPath: "C:\\Program Files\\gstreamer\\1.0\\mingw_x86_64"`. Other platforms default to version `1.22` and an empty path.

| Field | Description |
| --- | --- |
| `GSTVersion` | Pipeline compatibility fallback, minimum `1.22`. A successfully detected runtime version takes precedence. |
| `GSTPath` | GStreamer installation root. Empty uses platform auto-detection. |
| `Source` | `stream` accepts any info hash; `play` requires a torrent already listed in TorrServer. |
| `MaxTasks` | Maximum concurrent GStreamer tasks. `0` is unlimited; the least recently active task is removed when the limit is exceeded. |
| `InactiveMinutes` | Freeze and release an idle pipeline after this timeout. The task is removed 20 minutes later if it stays inactive. |
| `AACBitrateKbps` | Bitrate for non-AAC audio transcoding. It is doubled for more than two channels. |
| `AACChannels` | Output channels for non-AAC audio. `0` uses the source value, clamped to 1-8; fallback is 2. |
| `AACSamplerate` | Output sample rate for non-AAC audio. `0` selects the nearest supported source rate; fallback is 48000 Hz. |
| `SegmentSeconds` | Target HLS duration. Copy mode uses Matroska Cue boundaries when available. |
| `SegmentDiff` | Keyframe alignment tolerance in copy mode. `0` disables the limit. |
| `Subtitles` | Expose supported embedded text subtitles as segmented WebVTT. Bitmap subtitles are not converted. |
| `TranscodeH264` | Convert H.264 video to H.264 instead of copying it. |
| `TranscodeH265` | Convert H.265/HEVC video to H.264 instead of copying it. |
| `TranscodeAV1` | Convert AV1 video to H.264 instead of copying it. |
| `TranscodeVP9` | Convert VP9 video to H.264 instead of copying it. |
| `TranscodeVP8` | Allow VP8 input and convert it to H.264. VP8 is rejected when disabled. |
| `TranscodeAVI` | Allow AVI input and convert its video to H.264. AVI is rejected when disabled. |
| `HDRToSDR` | Tone-map detected PQ/HLG HDR to SDR and convert the video to H.264. Requires `hdrtonemap`. |
| `HardwareAcceleration` | Use a tested hardware H.264 encoder when available; otherwise use `x264enc`. |
| `UseGPU` | Allow GPU video encoding and HDR processing. CPU fallbacks are used when unavailable. |
| `X264Ultrafast` | Use the x264 `ultrafast` preset instead of `veryfast`, reducing CPU load at the cost of compression efficiency. |
| `VideoBitrate` | Target H.264 video bitrate in kbps when video transcoding is active. |

With all `Transcode*` options disabled, H.264, H.265/HEVC, AV1, and VP9 video is copied. AAC audio is also copied; other audio codecs are decoded and encoded to AAC.

## API

**Settings** (requires authentication when `--httpauth` is enabled; read/write only in `-gst` builds):

- `GET /gst/settings` — on `-gst` builds: `built_in`, current config, and platform defaults; on standard builds: `{ "built_in": false }` only
- `POST /gst/settings` — update or reset config (`404` on standard builds)

  ```json
  { "action": "set", "config": { "GSTVersion": 1.22, "Source": "stream" } }
  ```

  Reset to defaults:

  ```json
  { "action": "def" }
  ```

**Streaming** (available in `-gst` builds):

| Endpoint | Description |
| --- | --- |
| `GET /gst/echo` | GStreamer / gst-discoverer health check |
| `GET /gst/:hash/probe` | Probe torrent file metadata (`index`, `id`, or `fileID` query); successful probes are cached for one hour |
| `GET /gst/:hash/master.m3u8` | Create/reuse a task and return the HLS master playlist (`index`, `audio`, and optional `seconds` query) |
| `GET /gst/:hash/video.m3u8` | HLS media playlist referenced by the master playlist |
| `GET /gst/:hash/init.mp4` | Initialization segment |
| `GET /gst/:hash/seg/*segment` | Media segment |
| `GET /gst/:hash/subs/:track.m3u8` | Segmented WebVTT subtitle playlist |
| `GET /gst/:hash/subs/:track/:segment.vtt` | WebVTT subtitle segment |
| `GET /gst/:hash/heartbeat` | Keep the task and its torrent active; returns torrent state details |
| `GET /gst/remove` | Dispose the task and drop the torrent cache (`hash` or `id` query) |
