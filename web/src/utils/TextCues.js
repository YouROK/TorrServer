// Shows the active cues of the video's text tracks (WebVTT from HLS, from <track> elements)
// in our own layer instead of the browser's native rendering. The native cues are drawn at the
// very bottom of the video, behind the custom controls of the player; the layer can be moved.
//
// Tracks that would be shown natively are switched to "hidden": their cues are still loaded and
// reported as active, but the browser draws nothing.
export default class TextCues {
  constructor(video, layer) {
    this.video = video
    this.layer = layer
    this.tracks = new Set()
    this.shown = []

    this.onCueChange = () => this.render()
    this.onTracksChange = () => this.sync()
    this.list = video.textTracks
    this.list.addEventListener('addtrack', this.onTracksChange)
    this.list.addEventListener('removetrack', this.onTracksChange)
    this.list.addEventListener('change', this.onTracksChange)
    this.sync()
  }

  sync() {
    const current = new Set()
    for (let i = 0; i < this.list.length; i++) {
      const track = this.list[i]
      if (track.kind === 'subtitles' || track.kind === 'captions') {
        if (track.mode === 'showing') track.mode = 'hidden'
        current.add(track)
        if (!this.tracks.has(track)) track.addEventListener('cuechange', this.onCueChange)
      }
    }
    this.tracks.forEach(track => {
      if (!current.has(track)) track.removeEventListener('cuechange', this.onCueChange)
    })
    this.tracks = current
    this.render()
  }

  render() {
    const cues = []
    this.tracks.forEach(track => {
      const active = track.mode === 'disabled' ? null : track.activeCues
      for (let i = 0; active && i < active.length; i++) cues.push(active[i])
    })
    if (cues.length === this.shown.length && cues.every((cue, i) => cue === this.shown[i])) return
    this.shown = cues

    this.layer.textContent = ''
    cues.forEach(cue => {
      const line = document.createElement('div')
      if (typeof cue.getCueAsHTML === 'function') line.appendChild(cue.getCueAsHTML())
      else line.textContent = String(cue.text || '').replace(/<[^>]*>/g, '')
      this.layer.appendChild(line)
    })
  }

  destroy() {
    this.list.removeEventListener('addtrack', this.onTracksChange)
    this.list.removeEventListener('removetrack', this.onTracksChange)
    this.list.removeEventListener('change', this.onTracksChange)
    this.tracks.forEach(track => track.removeEventListener('cuechange', this.onCueChange))
    this.tracks = new Set()
    this.shown = []
    this.layer.textContent = ''
  }
}
