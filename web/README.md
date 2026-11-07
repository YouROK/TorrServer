# TorrServer web client

### Google Cast

Open a video in the built-in player and select **Google Cast** to choose a receiver
using the browser's device picker. This uses Google's official
[Web Sender SDK](https://developers.google.com/cast/docs/web_sender/integrate) and
Default Media Receiver; video streams directly to the receiver, without screen mirroring.
The SDK is fetched from Google only when a player opens in a secure browser context.

Use HTTPS and a browser that supports Google Cast (for example, desktop Chrome).
Browsers also permit a secure localhost context for development. Insecure HTTP
LAN pages keep local playback and the existing browser-native remote playback fallback.
The receiver must be discoverable and able to reach the media URL returned by
TorrServer's external media address setting. A localhost media URL, embedded URL
credentials, browser cookies or Basic authentication cannot be used by this integration.
HTTPS media also needs a certificate the receiver trusts. Reverse-proxy path prefixes
and stream query parameters are preserved.

MP4, WebM, Ogg and existing HLS streams can be sent; actual codec/container support
depends on the receiver. Cast does not add transcoding or make every torrent playable.
Local video pauses only after the receiver accepts the stream. The player's play/pause,
timeline and ten-second seek controls then operate on that stream. Local volume, speed,
subtitles and picture-in-picture controls are disabled while casting; use the receiver
for volume. **Stop casting** ends the receiver session. Closing the player leaves the
receiver playing and detaches these controls. Reopening a player does not automatically
take over an existing session; select Cast again to send the selected stream.

### How to start project

0. ignore first two steps if the server is on `localhost`
1. duplicate `.env_example` and rename it to `.env`
2. in `.env` file add server address to `REACT_APP_SERVER_HOST` (without last "/")
> `http://192.168.78.4:8090` - correct
>
> `http://192.168.78.4:8090/` - wrong
3. in `.env` file add TMDB api key
4. `NODE_OPTIONS=--openssl-legacy-provider yarn start`

### Eslint
> Prettier will fix the code every time the code is saved

- `yarn lint` - to find all linting problems
- `yarn fix` - to fix code

### How images were generated
`npx pwa-asset-generator public/logo.png public -m public/site.webmanifest -p "calc(50vh - 25%) calc(50vw - 25%)" -b "linear-gradient(135deg, rgb(50,54,55), rgb(84,90,94))" -q 100 -i public/index.html -f`
